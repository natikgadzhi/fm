package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/natikgadzhi/fm/internal/jmap"
)

// testEmail returns a sample email for use in tests.
func testEmail() jmap.Email {
	return jmap.Email{
		Id:        "M1234567890",
		ThreadId:  "T1234567890",
		MessageId: "<abc@fastmail.com>",
		From:      []jmap.Address{{Name: "Sender", Email: "sender@example.com"}},
		To:        []jmap.Address{{Name: "Recipient", Email: "recipient@example.com"}},
		Subject:   "Meeting tomorrow",
		Date:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		TextBody:  "Email body content here...",
		MailboxIds: map[string]bool{
			"INBOX": true,
		},
	}
}

func TestPutAndGet(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)
	email := testEmail()

	if err := c.Put(email, "fm fetch M1234567890"); err != nil {
		t.Fatalf("Put() error: %v", err)
	}

	got, err := c.Get("M1234567890")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got == nil {
		t.Fatal("Get() returned nil, want email")
	}

	if got.Id != email.Id {
		t.Errorf("Id = %q, want %q", got.Id, email.Id)
	}
	if got.ThreadId != email.ThreadId {
		t.Errorf("ThreadId = %q, want %q", got.ThreadId, email.ThreadId)
	}
	if got.MessageId != email.MessageId {
		t.Errorf("MessageId = %q, want %q", got.MessageId, email.MessageId)
	}
	if got.Subject != email.Subject {
		t.Errorf("Subject = %q, want %q", got.Subject, email.Subject)
	}
	if len(got.From) != 1 || got.From[0].Email != "sender@example.com" {
		t.Errorf("From = %v, want sender@example.com", got.From)
	}
	if len(got.To) != 1 || got.To[0].Email != "recipient@example.com" {
		t.Errorf("To = %v, want recipient@example.com", got.To)
	}
	if got.TextBody != "Email body content here..." {
		t.Errorf("TextBody = %q, want %q", got.TextBody, "Email body content here...")
	}
	if !got.Date.Equal(email.Date) {
		t.Errorf("Date = %v, want %v", got.Date, email.Date)
	}
}

func TestGetCacheMiss(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)

	got, err := c.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got != nil {
		t.Errorf("Get() = %v, want nil for cache miss", got)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)
	email := testEmail()

	if c.Exists("M1234567890") {
		t.Error("Exists() = true before Put, want false")
	}

	if err := c.Put(email, "fm fetch M1234567890"); err != nil {
		t.Fatalf("Put() error: %v", err)
	}

	if !c.Exists("M1234567890") {
		t.Error("Exists() = false after Put, want true")
	}
}

func TestFrontmatterRoundTrip(t *testing.T) {
	fm := Frontmatter{
		Tool:      "fm",
		Object:    "email",
		Id:        "M1234567890",
		ThreadId:  "T1234567890",
		MessageId: "<abc@fastmail.com>",
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Meeting tomorrow",
		Date:      "2025-01-15T10:30:00Z",
		Mailbox:   "INBOX",
		CachedAt:  "2025-01-15T12:00:00Z",
		SourceURL: "https://api.fastmail.com/jmap/api/",
		Command:   "fm fetch M1234567890",
	}

	data, err := Marshal(fm)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	// Append a body after the frontmatter.
	full := append(data, []byte("\nSome body text\n")...)

	got, body, err := Unmarshal(full)
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if got.Tool != fm.Tool {
		t.Errorf("Tool = %q, want %q", got.Tool, fm.Tool)
	}
	if got.Object != fm.Object {
		t.Errorf("Object = %q, want %q", got.Object, fm.Object)
	}
	if got.Id != fm.Id {
		t.Errorf("Id = %q, want %q", got.Id, fm.Id)
	}
	if got.ThreadId != fm.ThreadId {
		t.Errorf("ThreadId = %q, want %q", got.ThreadId, fm.ThreadId)
	}
	if got.MessageId != fm.MessageId {
		t.Errorf("MessageId = %q, want %q", got.MessageId, fm.MessageId)
	}
	if got.From != fm.From {
		t.Errorf("From = %q, want %q", got.From, fm.From)
	}
	if len(got.To) != 1 || got.To[0] != "recipient@example.com" {
		t.Errorf("To = %v, want %v", got.To, fm.To)
	}
	if got.Subject != fm.Subject {
		t.Errorf("Subject = %q, want %q", got.Subject, fm.Subject)
	}
	if got.Date != fm.Date {
		t.Errorf("Date = %q, want %q", got.Date, fm.Date)
	}
	if got.Mailbox != fm.Mailbox {
		t.Errorf("Mailbox = %q, want %q", got.Mailbox, fm.Mailbox)
	}
	if got.CachedAt != fm.CachedAt {
		t.Errorf("CachedAt = %q, want %q", got.CachedAt, fm.CachedAt)
	}
	if got.SourceURL != fm.SourceURL {
		t.Errorf("SourceURL = %q, want %q", got.SourceURL, fm.SourceURL)
	}
	if got.Command != fm.Command {
		t.Errorf("Command = %q, want %q", got.Command, fm.Command)
	}
	if !strings.Contains(body, "Some body text") {
		t.Errorf("body = %q, want to contain %q", body, "Some body text")
	}
}

func TestFileContentFormat(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)
	email := testEmail()

	if err := c.Put(email, "fm fetch M1234567890"); err != nil {
		t.Fatalf("Put() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "M1234567890.md"))
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	content := string(data)

	// Check that the file starts with frontmatter delimiters.
	if !strings.HasPrefix(content, "---\n") {
		t.Error("file does not start with frontmatter delimiter")
	}

	// Check key frontmatter fields.
	checks := []string{
		"tool: fm",
		"object: email",
		"id: M1234567890",
		"thread_id: T1234567890",
		`message_id: <abc@fastmail.com>`,
		"from: sender@example.com",
		"subject: Meeting tomorrow",
		"date: \"2025-01-15T10:30:00Z\"",
		"mailbox: INBOX",
		"command: fm fetch M1234567890",
		"source_url: https://api.fastmail.com/jmap/api/",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("file does not contain %q", want)
		}
	}

	// Check that the body contains the rendered headers and content.
	if !strings.Contains(content, "# Meeting tomorrow") {
		t.Error("file does not contain subject heading")
	}
	if !strings.Contains(content, "**From:** sender@example.com") {
		t.Error("file does not contain From header")
	}
	if !strings.Contains(content, "**To:** recipient@example.com") {
		t.Error("file does not contain To header")
	}
	if !strings.Contains(content, "**Date:** January 15, 2025 10:30 AM") {
		t.Error("file does not contain Date header")
	}
	if !strings.Contains(content, "Email body content here...") {
		t.Error("file does not contain email body")
	}
}

func TestCacheAutoCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "cache", "dir")
	c := NewCache(dir)
	email := testEmail()

	if err := c.Put(email, "fm fetch M1234567890"); err != nil {
		t.Fatalf("Put() error: %v", err)
	}

	if !c.Exists("M1234567890") {
		t.Error("Exists() = false after Put to auto-created dir")
	}
}

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"M1234567890", "M1234567890"},
		{"id/with/slashes", "id_with_slashes"},
		{"id:colon", "id_colon"},
		{"id<angle>bracket", "id_angle_bracket"},
		{"normal-id_123", "normal-id_123"},
	}

	for _, tt := range tests {
		got := sanitizeID(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestHTMLBodyFallback(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)
	email := jmap.Email{
		Id:        "Mhtml",
		ThreadId:  "Thtml",
		MessageId: "<html@test.com>",
		From:      []jmap.Address{{Email: "from@test.com"}},
		To:        []jmap.Address{{Email: "to@test.com"}},
		Subject:   "HTML only",
		Date:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		TextBody:  "",
		HtmlBody:  "<p>Hello <b>world</b></p>",
		MailboxIds: map[string]bool{
			"INBOX": true,
		},
	}

	if err := c.Put(email, "fm fetch Mhtml"); err != nil {
		t.Fatalf("Put() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "Mhtml.md"))
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Hello world") {
		t.Errorf("file does not contain stripped HTML body, got:\n%s", content)
	}
	if strings.Contains(content, "<p>") || strings.Contains(content, "<b>") {
		t.Error("file contains raw HTML tags")
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<p>Hello</p>", "Hello"},
		{"<b>Bold</b> and <i>italic</i>", "Bold and italic"},
		{"No tags here", "No tags here"},
		{"<div><p>Nested</p></div>", "Nested"},
		{"", ""},
	}

	for _, tt := range tests {
		got := stripHTMLTags(tt.input)
		if got != tt.want {
			t.Errorf("stripHTMLTags(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
