package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/natikgadzhi/fm/internal/jmap"
)

// sampleDate returns a fixed time for deterministic tests.
func sampleDate() time.Time {
	return time.Date(2025, time.March, 15, 10, 30, 0, 0, time.UTC)
}

func sampleEmails() []jmap.Email {
	return []jmap.Email{
		{
			Id:       "email-1",
			ThreadId: "thread-1",
			From: []jmap.Address{
				{Name: "Alice Smith", Email: "alice@example.com"},
			},
			To: []jmap.Address{
				{Name: "Bob Jones", Email: "bob@example.com"},
			},
			Subject:       "Meeting tomorrow",
			Date:          sampleDate(),
			TextBody:      "Hi Bob, let's meet tomorrow at 10am.",
			Preview:       "Hi Bob, let's meet tomorrow at 10am.",
			HasAttachment: false,
		},
		{
			Id:       "email-2",
			ThreadId: "thread-2",
			From: []jmap.Address{
				{Name: "Charlie Brown", Email: "charlie@example.com"},
			},
			To: []jmap.Address{
				{Name: "Alice Smith", Email: "alice@example.com"},
			},
			Subject:       "Project update: Q1 results are in and looking great for the team",
			Date:          sampleDate().Add(-2 * time.Hour),
			TextBody:      "Here are the Q1 results...",
			Preview:       "Here are the Q1 results for your review. Please take a look when you get a chance.",
			HasAttachment: true,
			Attachments: []jmap.Attachment{
				{BlobId: "blob-1", Name: "report.pdf", Type: "application/pdf", Size: 1024},
			},
		},
	}
}

func sampleEmail() jmap.Email {
	return jmap.Email{
		Id:       "email-1",
		ThreadId: "thread-1",
		From: []jmap.Address{
			{Name: "Alice Smith", Email: "alice@example.com"},
		},
		To: []jmap.Address{
			{Name: "Bob Jones", Email: "bob@example.com"},
		},
		Cc: []jmap.Address{
			{Name: "Charlie Brown", Email: "charlie@example.com"},
		},
		Subject:       "Meeting tomorrow",
		Date:          sampleDate(),
		TextBody:      "Hi Bob, let's meet tomorrow at 10am.\nSee you there.",
		Preview:       "Hi Bob, let's meet tomorrow at 10am.",
		HasAttachment: true,
		Attachments: []jmap.Attachment{
			{BlobId: "blob-1", Name: "agenda.pdf", Type: "application/pdf", Size: 2048},
		},
	}
}

func sampleMailboxes() []jmap.Mailbox {
	return []jmap.Mailbox{
		{
			Id:           "mb-1",
			Name:         "Inbox",
			Role:         "inbox",
			TotalEmails:  150,
			UnreadEmails: 12,
		},
		{
			Id:           "mb-2",
			Name:         "Sent",
			Role:         "sent",
			TotalEmails:  300,
			UnreadEmails: 0,
		},
		{
			Id:           "mb-3",
			Name:         "My Custom Folder With A Very Long Name That Should Be Truncated",
			Role:         "",
			TotalEmails:  5,
			UnreadEmails: 2,
		},
	}
}

// --- Factory tests ---

func TestNewValidFormats(t *testing.T) {
	for _, format := range []string{"text", "json", "markdown"} {
		f, err := New(format)
		if err != nil {
			t.Errorf("New(%q) returned error: %v", format, err)
		}
		if f == nil {
			t.Errorf("New(%q) returned nil formatter", format)
		}
	}
}

func TestNewInvalidFormat(t *testing.T) {
	f, err := New("xml")
	if err == nil {
		t.Error("New(\"xml\") should return an error")
	}
	if f != nil {
		t.Error("New(\"xml\") should return nil formatter")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error message should mention 'unsupported', got: %v", err)
	}
}

func TestNewEmptyFormat(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Error("New(\"\") should return an error")
	}
}

// --- TextFormatter tests ---

func TestTextFormatEmailList(t *testing.T) {
	f := &TextFormatter{}
	result, err := f.FormatEmailList(sampleEmails())
	if err != nil {
		t.Fatalf("FormatEmailList returned error: %v", err)
	}

	// Check header row
	if !strings.Contains(result, "ID") {
		t.Error("expected ID header")
	}
	if !strings.Contains(result, "THREAD ID") {
		t.Error("expected THREAD ID header")
	}
	if !strings.Contains(result, "DATE") {
		t.Error("expected DATE header")
	}
	if !strings.Contains(result, "FROM") {
		t.Error("expected FROM header")
	}
	if !strings.Contains(result, "SUBJECT") {
		t.Error("expected SUBJECT header")
	}

	// Check IDs are present (not truncated)
	if !strings.Contains(result, "email-1") {
		t.Error("expected email ID 'email-1' in output")
	}
	if !strings.Contains(result, "thread-1") {
		t.Error("expected thread ID 'thread-1' in output")
	}
	if !strings.Contains(result, "email-2") {
		t.Error("expected email ID 'email-2' in output")
	}
	if !strings.Contains(result, "thread-2") {
		t.Error("expected thread ID 'thread-2' in output")
	}

	// Check data
	if !strings.Contains(result, "Alice Smith") {
		t.Error("expected 'Alice Smith' in output")
	}
	if !strings.Contains(result, "Meeting tomorrow") {
		t.Error("expected 'Meeting tomorrow' in output")
	}
}

func TestTextFormatEmailListEmpty(t *testing.T) {
	f := &TextFormatter{}
	result, err := f.FormatEmailList(nil)
	if err != nil {
		t.Fatalf("FormatEmailList returned error: %v", err)
	}
	if !strings.Contains(result, "No emails found") {
		t.Errorf("expected empty message, got: %s", result)
	}
}

func TestTextFormatEmail(t *testing.T) {
	f := &TextFormatter{}
	result, err := f.FormatEmail(sampleEmail())
	if err != nil {
		t.Fatalf("FormatEmail returned error: %v", err)
	}

	if !strings.Contains(result, "Date:") {
		t.Error("expected Date header")
	}
	if !strings.Contains(result, "From:") {
		t.Error("expected From header")
	}
	if !strings.Contains(result, "To:") {
		t.Error("expected To header")
	}
	if !strings.Contains(result, "Cc:") {
		t.Error("expected Cc header")
	}
	if !strings.Contains(result, "Subject:") {
		t.Error("expected Subject header")
	}
	if !strings.Contains(result, "Alice Smith <alice@example.com>") {
		t.Error("expected formatted from address")
	}
	if !strings.Contains(result, "Hi Bob, let's meet tomorrow at 10am.") {
		t.Error("expected email body")
	}
	if !strings.Contains(result, "Attachments: 1") {
		t.Error("expected attachment count")
	}
}

func TestTextFormatMailboxes(t *testing.T) {
	f := &TextFormatter{}
	result, err := f.FormatMailboxes(sampleMailboxes())
	if err != nil {
		t.Fatalf("FormatMailboxes returned error: %v", err)
	}

	if !strings.Contains(result, "NAME") {
		t.Error("expected NAME header")
	}
	if !strings.Contains(result, "ROLE") {
		t.Error("expected ROLE header")
	}
	if !strings.Contains(result, "UNREAD") {
		t.Error("expected UNREAD header")
	}
	if !strings.Contains(result, "TOTAL") {
		t.Error("expected TOTAL header")
	}
	if !strings.Contains(result, "Inbox") {
		t.Error("expected Inbox in output")
	}
	if !strings.Contains(result, "inbox") {
		t.Error("expected inbox role in output")
	}
}

func TestTextFormatMailboxesEmpty(t *testing.T) {
	f := &TextFormatter{}
	result, err := f.FormatMailboxes(nil)
	if err != nil {
		t.Fatalf("FormatMailboxes returned error: %v", err)
	}
	if !strings.Contains(result, "No mailboxes found") {
		t.Errorf("expected empty message, got: %s", result)
	}
}

// --- Truncation tests ---

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a longer string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
		{"", 10, ""},
		{"hello", 0, ""},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

// --- JSONFormatter tests ---

func TestJSONFormatEmailList(t *testing.T) {
	f := &JSONFormatter{}
	result, err := f.FormatEmailList(sampleEmails())
	if err != nil {
		t.Fatalf("FormatEmailList returned error: %v", err)
	}

	// Verify it's valid JSON
	var parsed []map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, result)
	}
	if len(parsed) != 2 {
		t.Errorf("expected 2 emails in JSON output, got %d", len(parsed))
	}

	// Verify IDs are present in JSON output
	if parsed[0]["id"] != "email-1" {
		t.Errorf("expected id 'email-1', got %v", parsed[0]["id"])
	}
	if parsed[0]["threadId"] != "thread-1" {
		t.Errorf("expected threadId 'thread-1', got %v", parsed[0]["threadId"])
	}
	if parsed[1]["id"] != "email-2" {
		t.Errorf("expected id 'email-2', got %v", parsed[1]["id"])
	}
	if parsed[1]["threadId"] != "thread-2" {
		t.Errorf("expected threadId 'thread-2', got %v", parsed[1]["threadId"])
	}
}

func TestJSONFormatEmail(t *testing.T) {
	f := &JSONFormatter{}
	result, err := f.FormatEmail(sampleEmail())
	if err != nil {
		t.Fatalf("FormatEmail returned error: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, result)
	}

	if parsed["subject"] != "Meeting tomorrow" {
		t.Errorf("expected subject 'Meeting tomorrow', got %v", parsed["subject"])
	}
}

func TestJSONFormatMailboxes(t *testing.T) {
	f := &JSONFormatter{}
	result, err := f.FormatMailboxes(sampleMailboxes())
	if err != nil {
		t.Fatalf("FormatMailboxes returned error: %v", err)
	}

	var parsed []json.RawMessage
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, result)
	}
	if len(parsed) != 3 {
		t.Errorf("expected 3 mailboxes in JSON output, got %d", len(parsed))
	}
}

func TestJSONIsPrettyPrinted(t *testing.T) {
	f := &JSONFormatter{}
	result, err := f.FormatEmail(sampleEmail())
	if err != nil {
		t.Fatalf("FormatEmail returned error: %v", err)
	}

	// Pretty-printed JSON should have newlines and indentation
	if !strings.Contains(result, "\n") {
		t.Error("JSON output should contain newlines (pretty-printed)")
	}
	if !strings.Contains(result, "  ") {
		t.Error("JSON output should contain indentation (pretty-printed)")
	}
}

// --- MarkdownFormatter tests ---

func TestMarkdownFormatEmailList(t *testing.T) {
	f := &MarkdownFormatter{}
	result, err := f.FormatEmailList(sampleEmails())
	if err != nil {
		t.Fatalf("FormatEmailList returned error: %v", err)
	}

	// Check table structure includes ID and Thread ID columns
	if !strings.Contains(result, "| ID |") {
		t.Error("expected Markdown table header with ID")
	}
	if !strings.Contains(result, "| Thread ID |") {
		t.Error("expected Markdown table header with Thread ID")
	}
	if !strings.Contains(result, "| Date |") {
		t.Error("expected Markdown table header with Date")
	}
	if !strings.Contains(result, "| --- |") {
		t.Error("expected Markdown table separator")
	}
	if !strings.Contains(result, "Alice Smith") {
		t.Error("expected 'Alice Smith' in output")
	}

	// Check IDs are present in rows
	if !strings.Contains(result, "email-1") {
		t.Error("expected email ID 'email-1' in output")
	}
	if !strings.Contains(result, "thread-1") {
		t.Error("expected thread ID 'thread-1' in output")
	}

	// Count rows (header + separator + 2 data rows)
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines (header + separator + 2 rows), got %d", len(lines))
	}
}

func TestMarkdownFormatEmailListEmpty(t *testing.T) {
	f := &MarkdownFormatter{}
	result, err := f.FormatEmailList(nil)
	if err != nil {
		t.Fatalf("FormatEmailList returned error: %v", err)
	}
	if !strings.Contains(result, "No emails found") {
		t.Errorf("expected empty message, got: %s", result)
	}
}

func TestMarkdownFormatEmail(t *testing.T) {
	f := &MarkdownFormatter{}
	result, err := f.FormatEmail(sampleEmail())
	if err != nil {
		t.Fatalf("FormatEmail returned error: %v", err)
	}

	if !strings.Contains(result, "# Meeting tomorrow") {
		t.Error("expected Markdown heading with subject")
	}
	if !strings.Contains(result, "**Date:**") {
		t.Error("expected bold Date label")
	}
	if !strings.Contains(result, "**From:**") {
		t.Error("expected bold From label")
	}
	if !strings.Contains(result, "**To:**") {
		t.Error("expected bold To label")
	}
	if !strings.Contains(result, "**Cc:**") {
		t.Error("expected bold Cc label")
	}
	if !strings.Contains(result, "---") {
		t.Error("expected horizontal rule separator")
	}
	if !strings.Contains(result, "Hi Bob") {
		t.Error("expected email body content")
	}
	if !strings.Contains(result, "**Attachments:** 1") {
		t.Error("expected attachment count")
	}
}

func TestMarkdownFormatMailboxes(t *testing.T) {
	f := &MarkdownFormatter{}
	result, err := f.FormatMailboxes(sampleMailboxes())
	if err != nil {
		t.Fatalf("FormatMailboxes returned error: %v", err)
	}

	if !strings.Contains(result, "| Name |") {
		t.Error("expected Markdown table header with Name")
	}
	if !strings.Contains(result, "| --- |") {
		t.Error("expected Markdown table separator")
	}
	if !strings.Contains(result, "Inbox") {
		t.Error("expected 'Inbox' in output")
	}

	// The custom folder with no role should show "-"
	if !strings.Contains(result, "| - |") {
		t.Error("expected '-' for mailbox with no role")
	}
}

func TestMarkdownFormatMailboxesEmpty(t *testing.T) {
	f := &MarkdownFormatter{}
	result, err := f.FormatMailboxes(nil)
	if err != nil {
		t.Fatalf("FormatMailboxes returned error: %v", err)
	}
	if !strings.Contains(result, "No mailboxes found") {
		t.Errorf("expected empty message, got: %s", result)
	}
}

func TestMarkdownEscapesPipes(t *testing.T) {
	f := &MarkdownFormatter{}
	emails := []jmap.Email{
		{
			From: []jmap.Address{{Name: "Test | User", Email: "test@example.com"}},
			Subject: "Subject with | pipe",
			Date:    sampleDate(),
			Preview: "Preview text",
		},
	}
	result, err := f.FormatEmailList(emails)
	if err != nil {
		t.Fatalf("FormatEmailList returned error: %v", err)
	}
	// Pipes in content should be escaped
	if strings.Contains(result, "Test | User") {
		t.Error("pipe in name should be escaped")
	}
	if !strings.Contains(result, `Test \| User`) {
		t.Error("pipe in name should be escaped with backslash")
	}
}

// --- Helper function tests ---

func TestFormatAddress(t *testing.T) {
	tests := []struct {
		name  string
		addrs []jmap.Address
		want  string
	}{
		{"empty", nil, ""},
		{"name only", []jmap.Address{{Name: "Alice", Email: "alice@example.com"}}, "Alice"},
		{"email only", []jmap.Address{{Email: "alice@example.com"}}, "alice@example.com"},
	}
	for _, tt := range tests {
		got := formatAddress(tt.addrs)
		if got != tt.want {
			t.Errorf("formatAddress(%s) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestFormatAddressList(t *testing.T) {
	addrs := []jmap.Address{
		{Name: "Alice", Email: "alice@example.com"},
		{Email: "bob@example.com"},
	}
	got := formatAddressList(addrs)
	want := "Alice <alice@example.com>, bob@example.com"
	if got != want {
		t.Errorf("formatAddressList = %q, want %q", got, want)
	}
}
