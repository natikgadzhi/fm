package jmap

import "time"

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
