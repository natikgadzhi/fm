package cmd

import (
	"testing"
)

func TestListCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("list command should be registered on rootCmd")
	}
}

func TestListCommandFlags(t *testing.T) {
	flags := []struct {
		name     string
		defValue string
	}{
		{"limit", "50"},
		{"after", ""},
		{"before", ""},
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

func TestListCommandRequiresMailboxArg(t *testing.T) {
	// ExactArgs(1) means zero args should return an error.
	err := listCmd.Args(listCmd, []string{})
	if err == nil {
		t.Error("list command should require exactly one argument (mailbox name)")
	}
}

func TestListCommandRejectsExtraArgs(t *testing.T) {
	err := listCmd.Args(listCmd, []string{"INBOX", "extra"})
	if err == nil {
		t.Error("list command should reject more than one argument")
	}
}
