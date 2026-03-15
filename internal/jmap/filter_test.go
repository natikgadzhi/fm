package jmap

import (
	"testing"
	"time"
)

// mustParse is a test helper that calls ParseFilterQuery and fails on error.
func mustParse(t *testing.T, query string) SearchFilter {
	t.Helper()
	filter, err := ParseFilterQuery(query)
	if err != nil {
		t.Fatalf("ParseFilterQuery(%q) returned unexpected error: %v", query, err)
	}
	return filter
}

func TestParseFilterQueryEmpty(t *testing.T) {
	filter := mustParse(t, "")
	if filter != (SearchFilter{}) {
		t.Errorf("empty query should produce zero-value filter, got %+v", filter)
	}
}

func TestParseFilterQueryWhitespace(t *testing.T) {
	filter := mustParse(t, "   ")
	if filter != (SearchFilter{}) {
		t.Errorf("whitespace query should produce zero-value filter, got %+v", filter)
	}
}

func TestParseFilterQueryFreeTextOnly(t *testing.T) {
	filter := mustParse(t, "hello world")
	if filter.Text != "hello world" {
		t.Errorf("Text: got %q, want %q", filter.Text, "hello world")
	}
	if filter.From != "" || filter.To != "" || filter.Subject != "" {
		t.Errorf("other fields should be empty: %+v", filter)
	}
}

func TestParseFilterQueryFrom(t *testing.T) {
	filter := mustParse(t, "from:alice@example.com")
	if filter.From != "alice@example.com" {
		t.Errorf("From: got %q, want %q", filter.From, "alice@example.com")
	}
}

func TestParseFilterQueryTo(t *testing.T) {
	filter := mustParse(t, "to:bob@example.com")
	if filter.To != "bob@example.com" {
		t.Errorf("To: got %q, want %q", filter.To, "bob@example.com")
	}
}

func TestParseFilterQuerySubjectSingleWord(t *testing.T) {
	filter := mustParse(t, "subject:meeting")
	if filter.Subject != "meeting" {
		t.Errorf("Subject: got %q, want %q", filter.Subject, "meeting")
	}
}

func TestParseFilterQuerySubjectMultipleWords(t *testing.T) {
	filter := mustParse(t, "subject:meeting notes tomorrow")
	if filter.Subject != "meeting notes tomorrow" {
		t.Errorf("Subject: got %q, want %q", filter.Subject, "meeting notes tomorrow")
	}
}

func TestParseFilterQuerySubjectStopsAtKeyword(t *testing.T) {
	filter := mustParse(t, "subject:meeting notes from:alice@example.com")
	if filter.Subject != "meeting notes" {
		t.Errorf("Subject: got %q, want %q", filter.Subject, "meeting notes")
	}
	if filter.From != "alice@example.com" {
		t.Errorf("From: got %q, want %q", filter.From, "alice@example.com")
	}
}

func TestParseFilterQueryInMailbox(t *testing.T) {
	filter := mustParse(t, "in:INBOX")
	if filter.InMailbox != "INBOX" {
		t.Errorf("InMailbox: got %q, want %q", filter.InMailbox, "INBOX")
	}
}

func TestParseFilterQueryBeforeDate(t *testing.T) {
	filter := mustParse(t, "before:2025-06-15")
	expected := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	if filter.Before == nil {
		t.Fatal("Before should not be nil")
	}
	if !filter.Before.Equal(expected) {
		t.Errorf("Before: got %v, want %v", *filter.Before, expected)
	}
}

func TestParseFilterQueryAfterDate(t *testing.T) {
	filter := mustParse(t, "after:2025-01-01")
	expected := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if filter.After == nil {
		t.Fatal("After should not be nil")
	}
	if !filter.After.Equal(expected) {
		t.Errorf("After: got %v, want %v", *filter.After, expected)
	}
}

func TestParseFilterQueryInvalidDateReturnsError(t *testing.T) {
	_, err := ParseFilterQuery("before:not-a-date")
	if err == nil {
		t.Error("expected error for invalid date format, got nil")
	}
}

func TestParseFilterQueryInvalidAfterDateReturnsError(t *testing.T) {
	_, err := ParseFilterQuery("after:2025/01/01")
	if err == nil {
		t.Error("expected error for invalid after date format, got nil")
	}
}

func TestParseFilterQueryHasAttachment(t *testing.T) {
	filter := mustParse(t, "has:attachment")
	if !filter.HasAttachment {
		t.Error("HasAttachment should be true")
	}
}

func TestParseFilterQueryHasAttachmentCaseInsensitive(t *testing.T) {
	filter := mustParse(t, "Has:Attachment")
	if !filter.HasAttachment {
		t.Error("HasAttachment should be true (case insensitive)")
	}
}

func TestParseFilterQueryHasUnknownValue(t *testing.T) {
	filter := mustParse(t, "has:something")
	if filter.HasAttachment {
		t.Error("HasAttachment should be false for unknown has: value")
	}
}

func TestParseFilterQueryCombined(t *testing.T) {
	query := "from:alice@example.com to:bob@example.com subject:quarterly review in:INBOX before:2025-12-31 after:2025-01-01 has:attachment important"
	filter := mustParse(t, query)

	if filter.From != "alice@example.com" {
		t.Errorf("From: got %q, want %q", filter.From, "alice@example.com")
	}
	if filter.To != "bob@example.com" {
		t.Errorf("To: got %q, want %q", filter.To, "bob@example.com")
	}
	if filter.Subject != "quarterly review" {
		t.Errorf("Subject: got %q, want %q", filter.Subject, "quarterly review")
	}
	if filter.InMailbox != "INBOX" {
		t.Errorf("InMailbox: got %q, want %q", filter.InMailbox, "INBOX")
	}
	if filter.Before == nil {
		t.Fatal("Before should not be nil")
	}
	expectedBefore := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	if !filter.Before.Equal(expectedBefore) {
		t.Errorf("Before: got %v, want %v", *filter.Before, expectedBefore)
	}
	if filter.After == nil {
		t.Fatal("After should not be nil")
	}
	expectedAfter := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !filter.After.Equal(expectedAfter) {
		t.Errorf("After: got %v, want %v", *filter.After, expectedAfter)
	}
	if !filter.HasAttachment {
		t.Error("HasAttachment should be true")
	}
	if filter.Text != "important" {
		t.Errorf("Text: got %q, want %q", filter.Text, "important")
	}
}

func TestParseFilterQueryMultipleFreeTextWords(t *testing.T) {
	filter := mustParse(t, "hello from:user@test.com world foo bar")
	if filter.From != "user@test.com" {
		t.Errorf("From: got %q, want %q", filter.From, "user@test.com")
	}
	if filter.Text != "hello world foo bar" {
		t.Errorf("Text: got %q, want %q", filter.Text, "hello world foo bar")
	}
}

func TestParseFilterQueryLastFromWins(t *testing.T) {
	// When the same keyword appears multiple times, the last value wins.
	filter := mustParse(t, "from:first@example.com from:second@example.com")
	if filter.From != "second@example.com" {
		t.Errorf("From: got %q, want %q (last should win)", filter.From, "second@example.com")
	}
}

func TestParseFilterQueryKeywordCaseInsensitive(t *testing.T) {
	filter := mustParse(t, "FROM:alice@example.com TO:bob@example.com")
	if filter.From != "alice@example.com" {
		t.Errorf("From: got %q, want %q", filter.From, "alice@example.com")
	}
	if filter.To != "bob@example.com" {
		t.Errorf("To: got %q, want %q", filter.To, "bob@example.com")
	}
}

func TestParseFilterQuerySubjectEmpty(t *testing.T) {
	// subject: with nothing after it (end of query)
	filter := mustParse(t, "from:test@test.com subject:")
	if filter.Subject != "" {
		t.Errorf("Subject: got %q, want empty string", filter.Subject)
	}
}

func TestParseFilterQueryDateRange(t *testing.T) {
	filter := mustParse(t, "after:2025-01-01 before:2025-06-30")
	if filter.After == nil || filter.Before == nil {
		t.Fatal("Both After and Before should be set")
	}
	expectedAfter := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedBefore := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)
	if !filter.After.Equal(expectedAfter) {
		t.Errorf("After: got %v, want %v", *filter.After, expectedAfter)
	}
	if !filter.Before.Equal(expectedBefore) {
		t.Errorf("Before: got %v, want %v", *filter.Before, expectedBefore)
	}
}

func TestParseFilterQuerySubjectFollowedByFreeText(t *testing.T) {
	// This is an interesting edge case: free text after subject absorbs into subject
	// because free text tokens are not keywords.
	filter := mustParse(t, "subject:hello world")
	if filter.Subject != "hello world" {
		t.Errorf("Subject: got %q, want %q", filter.Subject, "hello world")
	}
	if filter.Text != "" {
		t.Errorf("Text: got %q, want empty (absorbed into subject)", filter.Text)
	}
}

func TestParseFilterQueryFreeTextBeforeSubject(t *testing.T) {
	filter := mustParse(t, "important subject:meeting notes")
	if filter.Subject != "meeting notes" {
		t.Errorf("Subject: got %q, want %q", filter.Subject, "meeting notes")
	}
	if filter.Text != "important" {
		t.Errorf("Text: got %q, want %q", filter.Text, "important")
	}
}
