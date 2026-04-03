package jmap

import (
	"fmt"
	"strings"
	"time"

	"github.com/natikgadzhi/cli-kit/output"
)

// Email represents a JMAP Email object with the fields we care about.
type Email struct {
	Id            string            `json:"id"`
	ThreadId      string            `json:"threadId"`
	MessageId     string            `json:"messageId"`
	From          []Address         `json:"from"`
	To            []Address         `json:"to"`
	Cc            []Address         `json:"cc"`
	Subject       string            `json:"subject"`
	Date          time.Time         `json:"date"`
	TextBody      string            `json:"textBody"`
	HtmlBody      string            `json:"htmlBody"`
	Preview       string            `json:"preview"`
	MailboxIds    map[string]bool   `json:"mailboxIds"`
	Size          int64             `json:"size"`
	HasAttachment bool              `json:"hasAttachment"`
	Attachments   []Attachment      `json:"attachments"`
}

// Address represents an email address with optional display name.
type Address struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Mailbox represents a JMAP Mailbox object (folder/label).
type Mailbox struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	TotalEmails  int    `json:"totalEmails"`
	UnreadEmails int    `json:"unreadEmails"`
	ParentId     string `json:"parentId"`
}

// Thread represents a JMAP Thread object — a group of related emails.
type Thread struct {
	Id       string   `json:"id"`
	EmailIds []string `json:"emailIds"`
}

// Attachment represents a file attached to an email.
type Attachment struct {
	BlobId  string `json:"blobId"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Size    int64  `json:"size"`
	Charset string `json:"charset"`
}

// SearchFilter holds the parsed search criteria for querying emails.
// Field values are populated from user query strings via ParseFilterQuery.
type SearchFilter struct {
	From          string     `json:"from,omitempty"`
	To            string     `json:"to,omitempty"`
	Subject       string     `json:"subject,omitempty"`
	Text          string     `json:"text,omitempty"`
	InMailbox     string     `json:"inMailbox,omitempty"`
	Before        *time.Time `json:"before,omitempty"`
	After         *time.Time `json:"after,omitempty"`
	HasAttachment bool       `json:"hasAttachment,omitempty"`
}

// dateFormat is the human-friendly date format used in table output.
const dateFormat = "Jan 02, 2006 3:04 PM"

// --- TableRenderer implementations for cli-kit/output ---

// EmailListRenderer wraps a slice of emails for table rendering.
type EmailListRenderer struct {
	Emails []Email
}

// RenderTable renders a list of emails as a table with ID, ThreadID, Date, From, Subject columns.
func (r *EmailListRenderer) RenderTable(t *output.Table) {
	t.Header("ID", "Thread ID", "Date", "From", "Subject")
	for _, email := range r.Emails {
		date := email.Date.Format(dateFormat)
		from := displayAddress(email.From)
		t.Row(email.Id, email.ThreadId, date, from, email.Subject)
	}
}

// EmailRenderer wraps a single email for table rendering.
type EmailRenderer struct {
	Email Email
}

// RenderTable renders a single email's full details as key-value rows.
func (r *EmailRenderer) RenderTable(t *output.Table) {
	e := r.Email
	t.Header("Field", "Value")
	t.Row("ID", e.Id)
	t.Row("Thread ID", e.ThreadId)
	t.Row("Date", e.Date.Format(dateFormat))
	t.Row("From", displayAddressList(e.From))
	t.Row("To", displayAddressList(e.To))
	if len(e.Cc) > 0 {
		t.Row("Cc", displayAddressList(e.Cc))
	}
	t.Row("Subject", e.Subject)
	if e.HasAttachment && len(e.Attachments) > 0 {
		t.Row("Attachments", fmt.Sprintf("%d", len(e.Attachments)))
	}
	// Add a blank row before body.
	body := e.TextBody
	if body == "" {
		body = e.Preview
	}
	t.Row("", "")
	t.Row("Body", body)
}

// MailboxListRenderer wraps a slice of mailboxes for table rendering.
type MailboxListRenderer struct {
	Mailboxes []Mailbox
}

// RenderTable renders mailboxes as a table with Name, Role, Unread, Total columns.
func (r *MailboxListRenderer) RenderTable(t *output.Table) {
	t.Header("Name", "Role", "Unread", "Total")
	for _, mb := range r.Mailboxes {
		role := mb.Role
		if role == "" {
			role = "-"
		}
		t.Row(mb.Name, role, fmt.Sprintf("%d", mb.UnreadEmails), fmt.Sprintf("%d", mb.TotalEmails))
	}
}

// CalendarListRenderer wraps a slice of calendars for table rendering.
type CalendarListRenderer struct {
	Calendars []Calendar
}

// RenderTable renders calendars as a table with Name, Color, Description, Read-only columns.
func (r *CalendarListRenderer) RenderTable(t *output.Table) {
	t.Header("Name", "Color", "Description", "Read-only")
	for _, cal := range r.Calendars {
		readOnly := "no"
		if cal.IsReadOnly {
			readOnly = "yes"
		}
		t.Row(cal.Name, cal.Color, cal.Description, readOnly)
	}
}

// CalendarEventListRenderer wraps a slice of calendar events for table rendering.
type CalendarEventListRenderer struct {
	Events []CalendarEvent
}

// RenderTable renders a list of calendar events as a table.
func (r *CalendarEventListRenderer) RenderTable(t *output.Table) {
	t.Header("ID", "Start", "Duration", "Title", "Location")
	for _, evt := range r.Events {
		loc := evt.Location
		if len(loc) > 40 {
			loc = loc[:37] + "..."
		}
		t.Row(evt.Id, evt.Start, evt.Duration, evt.Title, loc)
	}
}

// CalendarEventRenderer wraps a single calendar event for table rendering.
type CalendarEventRenderer struct {
	Event CalendarEvent
}

// RenderTable renders a single calendar event's full details as key-value rows.
func (r *CalendarEventRenderer) RenderTable(t *output.Table) {
	e := r.Event
	t.Header("Field", "Value")
	t.Row("ID", e.Id)
	t.Row("Title", e.Title)
	t.Row("Start", e.Start)
	t.Row("Time Zone", e.TimeZone)
	t.Row("Duration", e.Duration)
	if e.Location != "" {
		t.Row("Location", e.Location)
	}
	if e.Status != "" {
		t.Row("Status", e.Status)
	}

	// Attendees.
	if len(e.Participants) > 0 {
		var attendees []string
		for _, p := range e.Participants {
			if p.Email != "" {
				if p.Name != "" {
					attendees = append(attendees, fmt.Sprintf("%s <%s>", p.Name, p.Email))
				} else {
					attendees = append(attendees, p.Email)
				}
			}
		}
		if len(attendees) > 0 {
			t.Row("Attendees", strings.Join(attendees, ", "))
		}
	}

	// Recurrence.
	if len(e.RecurrenceRules) > 0 {
		var rules []string
		for _, rr := range e.RecurrenceRules {
			rule := rr.Frequency
			if rr.Interval > 1 {
				rule = fmt.Sprintf("Every %d %s", rr.Interval, rr.Frequency)
			}
			if rr.Until != "" {
				rule += " until " + rr.Until
			}
			if rr.Count > 0 {
				rule += fmt.Sprintf(" (%d times)", rr.Count)
			}
			rules = append(rules, rule)
		}
		t.Row("Recurrence", strings.Join(rules, "; "))
	}

	if e.Description != "" {
		t.Row("", "")
		t.Row("Description", e.Description)
	}
}

// displayAddress returns a display string for the first address in the list.
func displayAddress(addrs []Address) string {
	if len(addrs) == 0 {
		return ""
	}
	a := addrs[0]
	if a.Name != "" {
		return a.Name
	}
	return a.Email
}

// displayAddressList returns a comma-separated display string for all addresses.
func displayAddressList(addrs []Address) string {
	if len(addrs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(addrs))
	for _, a := range addrs {
		if a.Name != "" {
			parts = append(parts, fmt.Sprintf("%s <%s>", a.Name, a.Email))
		} else {
			parts = append(parts, a.Email)
		}
	}
	return strings.Join(parts, ", ")
}
