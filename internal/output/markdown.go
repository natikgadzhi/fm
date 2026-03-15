package output

import (
	"fmt"
	"strings"

	"github.com/natikgadzhi/fm/internal/jmap"
)

// MarkdownFormatter formats JMAP objects as Markdown.
type MarkdownFormatter struct{}

// FormatEmailList formats a list of emails as a Markdown table.
func (f *MarkdownFormatter) FormatEmailList(emails []jmap.Email) (string, error) {
	if len(emails) == 0 {
		return "*No emails found.*\n", nil
	}

	var b strings.Builder

	b.WriteString("| ID | Thread ID | Date | From | Subject |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")

	for _, email := range emails {
		date := email.Date.Format(dateFormat)
		from := escapeMarkdown(formatAddress(email.From))
		subject := escapeMarkdown(truncate(email.Subject, 60))

		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", email.Id, email.ThreadId, date, from, subject)
	}

	return b.String(), nil
}

// FormatEmail formats a single email as Markdown with headers and body.
func (f *MarkdownFormatter) FormatEmail(email jmap.Email) (string, error) {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", escapeMarkdown(email.Subject))
	fmt.Fprintf(&b, "**Date:** %s\n\n", email.Date.Format(dateFormat))
	fmt.Fprintf(&b, "**From:** %s\n\n", escapeMarkdown(formatAddressList(email.From)))
	fmt.Fprintf(&b, "**To:** %s\n\n", escapeMarkdown(formatAddressList(email.To)))
	if len(email.Cc) > 0 {
		fmt.Fprintf(&b, "**Cc:** %s\n\n", escapeMarkdown(formatAddressList(email.Cc)))
	}

	if email.HasAttachment && len(email.Attachments) > 0 {
		fmt.Fprintf(&b, "**Attachments:** %d\n\n", len(email.Attachments))
	}

	b.WriteString("---\n\n")

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

// FormatMailboxes formats a list of mailboxes as a Markdown table.
func (f *MarkdownFormatter) FormatMailboxes(mailboxes []jmap.Mailbox) (string, error) {
	if len(mailboxes) == 0 {
		return "*No mailboxes found.*\n", nil
	}

	var b strings.Builder

	b.WriteString("| Name | Role | Unread | Total |\n")
	b.WriteString("| --- | --- | ---: | ---: |\n")

	for _, mb := range mailboxes {
		role := mb.Role
		if role == "" {
			role = "-"
		}
		fmt.Fprintf(&b, "| %s | %s | %d | %d |\n",
			escapeMarkdown(mb.Name),
			escapeMarkdown(role),
			mb.UnreadEmails,
			mb.TotalEmails,
		)
	}

	return b.String(), nil
}

// escapeMarkdown escapes pipe characters that would break Markdown tables.
func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}
