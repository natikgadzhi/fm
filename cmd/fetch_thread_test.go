package cmd

import (
	"testing"
	"time"

	"github.com/natikgadzhi/fm/internal/jmap"
)

func TestFetchThreadCommandRegistered(t *testing.T) {
	// Verify that fetch-thread is registered as a subcommand of root.
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "fetch-thread" {
			found = true
			break
		}
	}
	if !found {
		t.Error("fetch-thread command is not registered on rootCmd")
	}
}

func TestFetchThreadCommandArgs(t *testing.T) {
	cmd := fetchThreadCmd

	// The command should require exactly 1 argument.
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error when no args provided")
	}
	if err := cmd.Args(cmd, []string{"T12345"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"T12345", "extra"}); err == nil {
		t.Error("expected error when too many args provided")
	}
}

func TestFetchThreadWithAttachmentsFlag(t *testing.T) {
	// Verify the --with-attachments flag is registered.
	flag := fetchThreadCmd.Flags().Lookup("with-attachments")
	if flag == nil {
		t.Fatal("--with-attachments flag not registered")
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default value 'false', got %q", flag.DefValue)
	}
}

func TestCollectParticipantsSingleMessage(t *testing.T) {
	emails := []jmap.Email{
		{
			From: []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
			To:   []jmap.Address{{Name: "Bob", Email: "bob@example.com"}},
		},
	}

	participants := collectParticipants(emails)
	if len(participants) != 2 {
		t.Fatalf("expected 2 participants, got %d: %v", len(participants), participants)
	}

	// Check that names are included.
	found := map[string]bool{}
	for _, p := range participants {
		found[p] = true
	}
	if !found["Alice <alice@example.com>"] {
		t.Error("expected Alice in participants")
	}
	if !found["Bob <bob@example.com>"] {
		t.Error("expected Bob in participants")
	}
}

func TestCollectParticipantsMultiMessage(t *testing.T) {
	emails := []jmap.Email{
		{
			From: []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
			To:   []jmap.Address{{Name: "Bob", Email: "bob@example.com"}},
		},
		{
			From: []jmap.Address{{Name: "Bob", Email: "bob@example.com"}},
			To:   []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
			Cc:   []jmap.Address{{Name: "Charlie", Email: "charlie@example.com"}},
		},
		{
			From: []jmap.Address{{Name: "Charlie", Email: "charlie@example.com"}},
			To: []jmap.Address{
				{Name: "Alice", Email: "alice@example.com"},
				{Name: "Bob", Email: "bob@example.com"},
			},
		},
	}

	participants := collectParticipants(emails)
	if len(participants) != 3 {
		t.Fatalf("expected 3 unique participants, got %d: %v", len(participants), participants)
	}
}

func TestCollectParticipantsDeduplicates(t *testing.T) {
	emails := []jmap.Email{
		{
			From: []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
			To:   []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
		},
	}

	participants := collectParticipants(emails)
	if len(participants) != 1 {
		t.Fatalf("expected 1 unique participant, got %d: %v", len(participants), participants)
	}
}

func TestCollectParticipantsNoName(t *testing.T) {
	emails := []jmap.Email{
		{
			From: []jmap.Address{{Email: "noname@example.com"}},
			To:   []jmap.Address{{Name: "Bob", Email: "bob@example.com"}},
		},
	}

	participants := collectParticipants(emails)
	if len(participants) != 2 {
		t.Fatalf("expected 2 participants, got %d: %v", len(participants), participants)
	}

	found := map[string]bool{}
	for _, p := range participants {
		found[p] = true
	}
	if !found["noname@example.com"] {
		t.Error("expected bare email address for participant without name")
	}
}

func TestCollectParticipantsEmpty(t *testing.T) {
	participants := collectParticipants(nil)
	if len(participants) != 0 {
		t.Fatalf("expected 0 participants for nil input, got %d", len(participants))
	}

	participants = collectParticipants([]jmap.Email{})
	if len(participants) != 0 {
		t.Fatalf("expected 0 participants for empty input, got %d", len(participants))
	}
}

func TestCollectParticipantsPreservesOrder(t *testing.T) {
	emails := []jmap.Email{
		{
			From: []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
			To:   []jmap.Address{{Name: "Bob", Email: "bob@example.com"}},
		},
		{
			From: []jmap.Address{{Name: "Charlie", Email: "charlie@example.com"}},
			To:   []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
		},
	}

	participants := collectParticipants(emails)
	if len(participants) != 3 {
		t.Fatalf("expected 3 participants, got %d", len(participants))
	}

	// First-seen order: Alice (from msg1 From), Bob (from msg1 To), Charlie (from msg2 From).
	expected := []string{
		"Alice <alice@example.com>",
		"Bob <bob@example.com>",
		"Charlie <charlie@example.com>",
	}
	for i, want := range expected {
		if participants[i] != want {
			t.Errorf("participant[%d] = %q, want %q", i, participants[i], want)
		}
	}
}

func TestFetchThreadCommandUsage(t *testing.T) {
	if fetchThreadCmd.Use != "fetch-thread <thread-id>" {
		t.Errorf("unexpected Use: %q", fetchThreadCmd.Use)
	}
	if fetchThreadCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
	if fetchThreadCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

// Helper to make a sample thread email list for testing display logic.
func sampleThreadEmails() []jmap.Email {
	baseDate := time.Date(2025, time.March, 15, 10, 0, 0, 0, time.UTC)
	return []jmap.Email{
		{
			Id:       "email-1",
			ThreadId: "T12345",
			From:     []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
			To:       []jmap.Address{{Name: "Bob", Email: "bob@example.com"}},
			Subject:  "Project discussion",
			Date:     baseDate,
			TextBody: "Let's discuss the project.",
			Preview:  "Let's discuss the project.",
		},
		{
			Id:       "email-2",
			ThreadId: "T12345",
			From:     []jmap.Address{{Name: "Bob", Email: "bob@example.com"}},
			To:       []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
			Subject:  "Re: Project discussion",
			Date:     baseDate.Add(1 * time.Hour),
			TextBody: "Sounds good, let's chat tomorrow.",
			Preview:  "Sounds good, let's chat tomorrow.",
		},
		{
			Id:       "email-3",
			ThreadId: "T12345",
			From:     []jmap.Address{{Name: "Alice", Email: "alice@example.com"}},
			To: []jmap.Address{
				{Name: "Bob", Email: "bob@example.com"},
				{Name: "Charlie", Email: "charlie@example.com"},
			},
			Subject:       "Re: Project discussion",
			Date:          baseDate.Add(2 * time.Hour),
			TextBody:      "I've added Charlie to the thread.",
			Preview:       "I've added Charlie to the thread.",
			HasAttachment: true,
			Attachments: []jmap.Attachment{
				{BlobId: "blob-1", Name: "notes.pdf", Type: "application/pdf", Size: 4096},
			},
		},
	}
}

func TestCollectParticipantsThreadEmails(t *testing.T) {
	emails := sampleThreadEmails()
	participants := collectParticipants(emails)

	// Should have Alice, Bob, Charlie (3 unique participants).
	if len(participants) != 3 {
		t.Fatalf("expected 3 participants, got %d: %v", len(participants), participants)
	}
}
