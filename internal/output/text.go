package output

import (
	"fmt"
	"strings"

	"github.com/natikgadzhi/fm/internal/jmap"
)

const (
	// dateFormat is the human-friendly date format used in text output.
	dateFormat = "Jan 02, 2006 3:04 PM"

	// Column widths for tabular output.
	idWidth      = 15
	threadWidth  = 15
	dateWidth    = 20
	fromWidth    = 25
	subjectWidth = 40
	previewWidth = 50
	nameWidth    = 30
	roleWidth    = 15
)

// TextFormatter formats JMAP objects as plain text tables.
type TextFormatter struct{}

// FormatEmailList formats a list of emails as a text table with
// ID, ThreadId, Date, From, and Subject columns.
func (f *TextFormatter) FormatEmailList(emails []jmap.Email) (string, error) {
	if len(emails) == 0 {
		return "No emails found.\n", nil
	}

	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "%-*s  %-*s  %-*s  %-*s  %-*s\n",
		idWidth, "ID",
		threadWidth, "THREAD ID",
		dateWidth, "DATE",
		fromWidth, "FROM",
		subjectWidth, "SUBJECT",
	)
	b.WriteString(strings.Repeat("-", idWidth+threadWidth+dateWidth+fromWidth+subjectWidth+8) + "\n")

	for _, email := range emails {
		date := email.Date.Format(dateFormat)
		from := formatAddress(email.From)
		subject := truncate(email.Subject, subjectWidth)

		fmt.Fprintf(&b, "%-*s  %-*s  %-*s  %-*s  %-*s\n",
			idWidth, email.Id,
			threadWidth, email.ThreadId,
			dateWidth, truncate(date, dateWidth),
			fromWidth, truncate(from, fromWidth),
			subjectWidth, subject,
		)
	}

	return b.String(), nil
}

// FormatEmail formats a single email with headers and body as plain text.
func (f *TextFormatter) FormatEmail(email jmap.Email) (string, error) {
	var b strings.Builder

	fmt.Fprintf(&b, "Date:    %s\n", email.Date.Format(dateFormat))
	fmt.Fprintf(&b, "From:    %s\n", formatAddressList(email.From))
	fmt.Fprintf(&b, "To:      %s\n", formatAddressList(email.To))
	if len(email.Cc) > 0 {
		fmt.Fprintf(&b, "Cc:      %s\n", formatAddressList(email.Cc))
	}
	fmt.Fprintf(&b, "Subject: %s\n", email.Subject)

	if email.HasAttachment && len(email.Attachments) > 0 {
		fmt.Fprintf(&b, "Attachments: %d\n", len(email.Attachments))
	}

	b.WriteString("\n")

	body := email.TextBody
	if body == "" {
		body = email.Preview
	}
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}

	return b.String(), nil
}

// FormatMailboxes formats a list of mailboxes as a text table with
// Name, Role, Unread, and Total columns.
func (f *TextFormatter) FormatMailboxes(mailboxes []jmap.Mailbox) (string, error) {
	if len(mailboxes) == 0 {
		return "No mailboxes found.\n", nil
	}

	var b strings.Builder

	fmt.Fprintf(&b, "%-*s  %-*s  %8s  %8s\n",
		nameWidth, "NAME",
		roleWidth, "ROLE",
		"UNREAD",
		"TOTAL",
	)
	b.WriteString(strings.Repeat("-", nameWidth+roleWidth+8+8+6) + "\n")

	for _, mb := range mailboxes {
		role := mb.Role
		if role == "" {
			role = "-"
		}
		fmt.Fprintf(&b, "%-*s  %-*s  %8d  %8d\n",
			nameWidth, truncate(mb.Name, nameWidth),
			roleWidth, truncate(role, roleWidth),
			mb.UnreadEmails,
			mb.TotalEmails,
		)
	}

	return b.String(), nil
}

// truncate shortens s to maxLen characters, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatAddress returns a display string for the first address in the list.
func formatAddress(addrs []jmap.Address) string {
	if len(addrs) == 0 {
		return ""
	}
	a := addrs[0]
	if a.Name != "" {
		return a.Name
	}
	return a.Email
}

// formatAddressList returns a comma-separated display string for all addresses.
func formatAddressList(addrs []jmap.Address) string {
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
