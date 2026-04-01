package cache

import (
	"fmt"

	"github.com/natikgadzhi/cli-kit/derived"
)

// Frontmatter holds the metadata for a cached email file.
type Frontmatter struct {
	Tool      string   `yaml:"tool"`
	Object    string   `yaml:"object"`
	Id        string   `yaml:"id"`
	ThreadId  string   `yaml:"thread_id"`
	MessageId string   `yaml:"message_id"`
	From      string   `yaml:"from"`
	To        []string `yaml:"to"`
	Subject   string   `yaml:"subject"`
	Date      string   `yaml:"date"`
	Mailbox   string   `yaml:"mailbox"`
	CachedAt  string   `yaml:"cached_at"`
	SourceURL string   `yaml:"source_url"`
	Command   string   `yaml:"command"`
}

// Marshal renders a Frontmatter struct as YAML frontmatter bytes,
// including the leading and trailing "---" delimiters.
// Uses cli-kit/derived.Render with a map[string]any representation.
func Marshal(fm Frontmatter) ([]byte, error) {
	meta := map[string]any{
		"tool":       fm.Tool,
		"object":     fm.Object,
		"id":         fm.Id,
		"thread_id":  fm.ThreadId,
		"message_id": fm.MessageId,
		"from":       fm.From,
		"to":         fm.To,
		"subject":    fm.Subject,
		"date":       fm.Date,
		"mailbox":    fm.Mailbox,
		"cached_at":  fm.CachedAt,
		"source_url": fm.SourceURL,
		"command":    fm.Command,
	}
	return derived.Render(meta, nil), nil
}

// Unmarshal parses YAML frontmatter and the Markdown body from file content.
// It returns the parsed Frontmatter, the body text after the frontmatter,
// and any error encountered.
// Uses cli-kit/derived.Parse for the parsing.
func Unmarshal(data []byte) (*Frontmatter, string, error) {
	meta, body, err := derived.Parse(data)
	if err != nil {
		return nil, "", err
	}
	if meta == nil {
		return nil, "", fmt.Errorf("missing frontmatter")
	}

	fm := &Frontmatter{
		Tool:      getString(meta, "tool"),
		Object:    getString(meta, "object"),
		Id:        getString(meta, "id"),
		ThreadId:  getString(meta, "thread_id"),
		MessageId: getString(meta, "message_id"),
		From:      getString(meta, "from"),
		Subject:   getString(meta, "subject"),
		Date:      getString(meta, "date"),
		Mailbox:   getString(meta, "mailbox"),
		CachedAt:  getString(meta, "cached_at"),
		SourceURL: getString(meta, "source_url"),
		Command:   getString(meta, "command"),
		To:        getStringSlice(meta, "to"),
	}

	return fm, string(body), nil
}

// getString extracts a string value from a map.
func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

// getStringSlice extracts a string slice from a map.
func getStringSlice(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	switch val := v.(type) {
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			result = append(result, fmt.Sprintf("%v", item))
		}
		return result
	case []string:
		return val
	default:
		return nil
	}
}
