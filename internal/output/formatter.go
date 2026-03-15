// Package output provides formatters for displaying emails and mailboxes.
package output

import (
	"fmt"

	"github.com/natikgadzhi/fm/internal/jmap"
)

// Formatter defines the interface for formatting JMAP objects for display.
type Formatter interface {
	// FormatEmailList formats a list of emails for display.
	FormatEmailList(emails []jmap.Email) (string, error)

	// FormatEmail formats a single email for display.
	FormatEmail(email jmap.Email) (string, error)

	// FormatMailboxes formats a list of mailboxes for display.
	FormatMailboxes(mailboxes []jmap.Mailbox) (string, error)
}

// New creates a new Formatter for the given format string.
// Supported formats: "text", "json", "markdown".
func New(format string) (Formatter, error) {
	switch format {
	case "text":
		return &TextFormatter{}, nil
	case "json":
		return &JSONFormatter{}, nil
	case "markdown":
		return &MarkdownFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %q (supported: text, json, markdown)", format)
	}
}
