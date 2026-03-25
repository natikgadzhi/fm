package cmd

import (
	"testing"
)

func TestVersionCommandRegistered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("version subcommand should be registered on root command")
	}
}

func TestVersionDefaultValues(t *testing.T) {
	// Verify the default build-time variables are set.
	if Version == "" {
		t.Error("Version should have a default value")
	}
	if Commit == "" {
		t.Error("Commit should have a default value")
	}
	if Date == "" {
		t.Error("Date should have a default value")
	}
}
