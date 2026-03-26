package cmd

import (
	"testing"
)

func TestListCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range emailCmd.Commands() {
		if cmd.Name() == "list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("list command should be registered on emailCmd")
	}
}

func TestListCommandFlags(t *testing.T) {
	flags := []struct {
		name     string
		defValue string
	}{
		{"limit", "20"},
	}

	for _, f := range flags {
		flag := listCmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag --%s should be registered on list command", f.name)
			continue
		}
		if flag.DefValue != f.defValue {
			t.Errorf("flag --%s default: got %q, want %q", f.name, flag.DefValue, f.defValue)
		}
	}
}

func TestListCommandRequiresExactlyOneArg(t *testing.T) {
	// The command should require exactly one argument (the mailbox).
	if listCmd.Args == nil {
		t.Fatal("list command should have an Args validator")
	}

	// Zero args should fail.
	if err := listCmd.Args(listCmd, []string{}); err == nil {
		t.Error("list command should reject zero arguments")
	}

	// One arg should succeed.
	if err := listCmd.Args(listCmd, []string{"INBOX"}); err != nil {
		t.Errorf("list command should accept one argument, got error: %v", err)
	}

	// Two args should fail.
	if err := listCmd.Args(listCmd, []string{"INBOX", "Sent"}); err == nil {
		t.Error("list command should reject two arguments")
	}
}
