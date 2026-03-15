package cache

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// frontmatterDelimiter is the YAML frontmatter delimiter used in Markdown files.
const frontmatterDelimiter = "---"

// Frontmatter holds the YAML metadata for a cached email file.
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
func Marshal(fm Frontmatter) ([]byte, error) {
	yamlData, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("marshaling frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(frontmatterDelimiter)
	buf.WriteByte('\n')
	buf.Write(yamlData)
	buf.WriteString(frontmatterDelimiter)
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

// Unmarshal parses YAML frontmatter and the Markdown body from file content.
// It returns the parsed Frontmatter, the body text after the frontmatter,
// and any error encountered.
func Unmarshal(data []byte) (*Frontmatter, string, error) {
	content := string(data)

	// The file must start with "---\n".
	if len(content) < 4 || content[:4] != frontmatterDelimiter+"\n" {
		return nil, "", fmt.Errorf("missing opening frontmatter delimiter")
	}

	// Find the closing delimiter.
	rest := content[4:]
	idx := bytes.Index([]byte(rest), []byte("\n"+frontmatterDelimiter+"\n"))
	if idx < 0 {
		return nil, "", fmt.Errorf("missing closing frontmatter delimiter")
	}

	yamlBlock := rest[:idx]
	body := rest[idx+len("\n"+frontmatterDelimiter+"\n"):]

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return nil, "", fmt.Errorf("parsing frontmatter YAML: %w", err)
	}

	return &fm, body, nil
}
