package cmd

import (
	"testing"

	"github.com/natikgadzhi/fm/internal/jmap"
)

func TestSearchCommandRegistered(t *testing.T) {
	// Verify the search command is registered as a subcommand of root.
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "search" {
			found = true
			break
		}
	}
	if !found {
		t.Error("search command should be registered on rootCmd")
	}
}

func TestSearchCommandFlags(t *testing.T) {
	// Verify all expected flags exist on the search command.
	flags := []struct {
		name     string
		defValue string
	}{
		{"limit", "25"},
		{"from", ""},
		{"to", ""},
		{"has-attachments", "false"},
	}

	for _, f := range flags {
		flag := searchCmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag --%s should be registered on search command", f.name)
			continue
		}
		if flag.DefValue != f.defValue {
			t.Errorf("flag --%s default: got %q, want %q", f.name, flag.DefValue, f.defValue)
		}
	}
}

func TestMergeFilterFlagsFromOverrides(t *testing.T) {
	filter := jmap.SearchFilter{
		From: "query@example.com",
		Text: "hello",
	}

	merged := MergeFilterFlags(filter, "flag@example.com", "", false)

	if merged.From != "flag@example.com" {
		t.Errorf("From: got %q, want %q", merged.From, "flag@example.com")
	}
	// Text should be preserved.
	if merged.Text != "hello" {
		t.Errorf("Text: got %q, want %q", merged.Text, "hello")
	}
}

func TestMergeFilterFlagsToOverrides(t *testing.T) {
	filter := jmap.SearchFilter{
		To:   "query@example.com",
		Text: "hello",
	}

	merged := MergeFilterFlags(filter, "", "flag@example.com", false)

	if merged.To != "flag@example.com" {
		t.Errorf("To: got %q, want %q", merged.To, "flag@example.com")
	}
	if merged.Text != "hello" {
		t.Errorf("Text: got %q, want %q", merged.Text, "hello")
	}
}

func TestMergeFilterFlagsHasAttachmentOverrides(t *testing.T) {
	filter := jmap.SearchFilter{
		HasAttachment: false,
	}

	merged := MergeFilterFlags(filter, "", "", true)

	if !merged.HasAttachment {
		t.Error("HasAttachment should be true when flag is set")
	}
}

func TestMergeFilterFlagsDoesNotOverrideWithEmpty(t *testing.T) {
	filter := jmap.SearchFilter{
		From: "query@example.com",
		To:   "query-to@example.com",
	}

	// Empty flag values should not override query values.
	merged := MergeFilterFlags(filter, "", "", false)

	if merged.From != "query@example.com" {
		t.Errorf("From: got %q, want %q (should not override with empty)", merged.From, "query@example.com")
	}
	if merged.To != "query-to@example.com" {
		t.Errorf("To: got %q, want %q (should not override with empty)", merged.To, "query-to@example.com")
	}
}

func TestMergeFilterFlagsCombined(t *testing.T) {
	// Simulate: fm search --from alice@example.com "subject:report to:bob@example.com"
	filter, err := jmap.ParseFilterQuery("subject:report to:bob@example.com")
	if err != nil {
		t.Fatalf("ParseFilterQuery returned unexpected error: %v", err)
	}

	merged := MergeFilterFlags(filter, "alice@example.com", "", false)

	if merged.From != "alice@example.com" {
		t.Errorf("From: got %q, want %q", merged.From, "alice@example.com")
	}
	if merged.To != "bob@example.com" {
		t.Errorf("To: got %q, want %q", merged.To, "bob@example.com")
	}
	if merged.Subject != "report" {
		t.Errorf("Subject: got %q, want %q", merged.Subject, "report")
	}
}

func TestMergeFilterFlagsBothFlagsAndQueryFrom(t *testing.T) {
	// CLI flag should override query string from:
	filter, err := jmap.ParseFilterQuery("from:query@example.com important stuff")
	if err != nil {
		t.Fatalf("ParseFilterQuery returned unexpected error: %v", err)
	}

	merged := MergeFilterFlags(filter, "flag@example.com", "", false)

	if merged.From != "flag@example.com" {
		t.Errorf("From: got %q, want %q (CLI flag should override)", merged.From, "flag@example.com")
	}
	if merged.Text != "important stuff" {
		t.Errorf("Text: got %q, want %q", merged.Text, "important stuff")
	}
}

func TestMergeFilterFlagsHasAttachmentFalseDoesNotOverrideTrue(t *testing.T) {
	// has:attachment in query string should not be cleared by flag being false.
	filter := jmap.SearchFilter{
		HasAttachment: true,
	}

	merged := MergeFilterFlags(filter, "", "", false)

	if !merged.HasAttachment {
		t.Error("HasAttachment: should remain true when flag is false (false means not set)")
	}
}
