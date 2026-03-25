// Package cache manages the local Markdown file cache for fetched emails.
package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/natikgadzhi/cli-kit/derived"
	"github.com/natikgadzhi/fm/internal/jmap"
)

// Cache provides read and write access to the local Markdown email cache.
type Cache struct {
	dir string
}

// NewCache creates a Cache that stores files under the given directory.
// The directory is created automatically on the first write if it does not exist.
func NewCache(dir string) *Cache {
	return &Cache{dir: dir}
}

// Exists reports whether an email with the given ID is cached on disk.
func (c *Cache) Exists(id string) bool {
	_, err := os.Stat(c.path(id))
	return err == nil
}

// Get reads a cached email from disk and returns the reconstructed Email.
// If the file does not exist, Get returns (nil, nil).
func (c *Cache) Get(id string) (*jmap.Email, error) {
	data, err := os.ReadFile(c.path(id))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading cache file: %w", err)
	}

	fm, body, err := Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("parsing cache file: %w", err)
	}

	date, _ := time.Parse(time.RFC3339, fm.Date)

	// Extract the email body from the rendered markdown.
	// The body starts after the header block (From/To/Date lines and blank line).
	emailBody := extractBody(body)

	email := &jmap.Email{
		Id:        fm.Id,
		ThreadId:  fm.ThreadId,
		MessageId: fm.MessageId,
		From:      []jmap.Address{{Email: fm.From}},
		Subject:   fm.Subject,
		Date:      date,
		TextBody:  emailBody,
	}

	// Reconstruct To addresses.
	for _, addr := range fm.To {
		email.To = append(email.To, jmap.Address{Email: addr})
	}

	// Reconstruct MailboxIds from the mailbox name stored in frontmatter.
	if fm.Mailbox != "" {
		email.MailboxIds = map[string]bool{fm.Mailbox: true}
	}

	return email, nil
}

// Put writes an email to the cache as a Markdown file with YAML frontmatter.
// The command string records which fm command produced this cache entry.
func (c *Cache) Put(email jmap.Email, command string) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	// Build the Markdown body.
	body := renderBody(email)

	// Create the derived frontmatter using cli-kit.
	dfm := derived.NewFrontmatter("fm", "email", email.Id, "https://api.fastmail.com/jmap/api/", command)

	// Also build our own extended frontmatter for the cache file.
	// We use our custom Frontmatter that includes email-specific fields,
	// and keep the created_at/updated_at from the derived frontmatter.
	fm := Frontmatter{
		Tool:      "fm",
		Object:    "email",
		Id:        email.Id,
		ThreadId:  email.ThreadId,
		MessageId: email.MessageId,
		From:      formatAddress(email.From),
		To:        formatAddresses(email.To),
		Subject:   email.Subject,
		Date:      email.Date.UTC().Format(time.RFC3339),
		Mailbox:   firstMailbox(email.MailboxIds),
		CachedAt:  dfm.CreatedAt,
		SourceURL: "https://api.fastmail.com/jmap/api/",
		Command:   command,
	}

	header, err := Marshal(fm)
	if err != nil {
		return err
	}

	content := append(header, []byte(body)...)

	if err := os.WriteFile(c.path(email.Id), content, 0o644); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	return nil
}

// path returns the filesystem path for the given email ID.
func (c *Cache) path(id string) string {
	return filepath.Join(c.dir, sanitizeID(id)+".md")
}

// sanitizeID replaces characters that are unsafe in filenames.
func sanitizeID(id string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"<", "_",
		">", "_",
		"\"", "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)
	return replacer.Replace(id)
}

// formatAddress returns the email string of the first address, or empty string.
func formatAddress(addrs []jmap.Address) string {
	if len(addrs) == 0 {
		return ""
	}
	return addrs[0].Email
}

// formatAddresses returns a slice of email strings.
func formatAddresses(addrs []jmap.Address) []string {
	out := make([]string, len(addrs))
	for i, a := range addrs {
		out[i] = a.Email
	}
	return out
}

// firstMailbox returns the first key from the MailboxIds map, or empty string.
func firstMailbox(ids map[string]bool) string {
	for k := range ids {
		return k
	}
	return ""
}

// renderBody produces the Markdown body for a cached email file.
func renderBody(email jmap.Email) string {
	var b strings.Builder

	b.WriteString("\n# ")
	b.WriteString(email.Subject)
	b.WriteString("\n\n")

	b.WriteString("**From:** ")
	b.WriteString(formatAddress(email.From))
	b.WriteString("\n")

	b.WriteString("**To:** ")
	b.WriteString(strings.Join(formatAddresses(email.To), ", "))
	b.WriteString("\n")

	b.WriteString("**Date:** ")
	b.WriteString(email.Date.Format("January 2, 2006 3:04 PM"))
	b.WriteString("\n")

	body := email.TextBody
	if body == "" {
		body = stripHTMLTags(email.HtmlBody)
	}

	if body != "" {
		b.WriteString("\n")
		b.WriteString(body)
		b.WriteString("\n")
	}

	return b.String()
}

// stripHTMLTags removes HTML tags from a string as a simple fallback
// when no text body is available.
func stripHTMLTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// extractBody extracts the email body text from the rendered Markdown.
// It skips the heading line (# Subject) and the header block (From/To/Date),
// returning only the actual email content.
func extractBody(body string) string {
	lines := strings.Split(body, "\n")
	// Skip leading empty lines, heading, header lines, and the blank line after headers.
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	// Skip the heading line (# Subject).
	if i < len(lines) && strings.HasPrefix(lines[i], "# ") {
		i++
	}
	// Skip blank line after heading.
	if i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	// Skip bold header lines (From, To, Date).
	for i < len(lines) && strings.HasPrefix(lines[i], "**") {
		i++
	}
	// Skip blank line after headers.
	if i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	result := strings.Join(lines[i:], "\n")
	return strings.TrimRight(result, "\n")
}
