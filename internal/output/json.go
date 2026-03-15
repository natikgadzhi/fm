package output

import (
	"encoding/json"

	"github.com/natikgadzhi/fm/internal/jmap"
)

// JSONFormatter formats JMAP objects as pretty-printed JSON.
type JSONFormatter struct{}

// FormatEmailList formats a list of emails as pretty-printed JSON.
func (f *JSONFormatter) FormatEmailList(emails []jmap.Email) (string, error) {
	return marshalIndent(emails)
}

// FormatEmail formats a single email as pretty-printed JSON.
func (f *JSONFormatter) FormatEmail(email jmap.Email) (string, error) {
	return marshalIndent(email)
}

// FormatMailboxes formats a list of mailboxes as pretty-printed JSON.
func (f *JSONFormatter) FormatMailboxes(mailboxes []jmap.Mailbox) (string, error) {
	return marshalIndent(mailboxes)
}

// marshalIndent is a helper that JSON-encodes v with indentation.
func marshalIndent(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}
