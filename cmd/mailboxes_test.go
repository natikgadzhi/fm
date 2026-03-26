package cmd

import (
	"testing"
)

func TestMailboxesCommandRegistered(t *testing.T) {
	// Verify the mailboxes command is registered on the email command.
	found := false
	for _, cmd := range emailCmd.Commands() {
		if cmd.Name() == "mailboxes" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'mailboxes' subcommand to be registered on email command")
	}
}

func TestMailboxesCommandNoArgs(t *testing.T) {
	// The mailboxes command should accept no positional arguments.
	cmd := mailboxesCmd
	if cmd.Args == nil {
		t.Fatal("expected Args validator to be set")
	}
	// Passing args should be rejected by cobra.NoArgs.
	err := cmd.Args(cmd, []string{"extra"})
	if err == nil {
		t.Error("expected error when passing positional arguments to mailboxes command")
	}
}

func TestMailboxesCommandMetadata(t *testing.T) {
	if mailboxesCmd.Use != "mailboxes" {
		t.Errorf("expected Use to be 'mailboxes', got %q", mailboxesCmd.Use)
	}
	if mailboxesCmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if mailboxesCmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}
