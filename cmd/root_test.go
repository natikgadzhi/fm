package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/natikgadzhi/fm/internal/auth"
)

func TestExitCodeSuccess(t *testing.T) {
	code := exitCode(nil)
	if code != ExitSuccess {
		t.Errorf("exitCode(nil): got %d, want %d", code, ExitSuccess)
	}
}

func TestExitCodeGeneralError(t *testing.T) {
	code := exitCode(errors.New("some error"))
	if code != ExitError {
		t.Errorf("exitCode(general error): got %d, want %d", code, ExitError)
	}
}

func TestExitCodeAuthError(t *testing.T) {
	err := &auth.AuthError{Message: "test auth error"}
	code := exitCode(err)
	if code != ExitAuthError {
		t.Errorf("exitCode(auth error): got %d, want %d", code, ExitAuthError)
	}
}

func TestExitCodeWrappedAuthError(t *testing.T) {
	inner := &auth.AuthError{Message: "inner auth error"}
	wrapped := errors.Join(errors.New("wrapper"), inner)
	code := exitCode(wrapped)
	if code != ExitAuthError {
		t.Errorf("exitCode(wrapped auth error): got %d, want %d", code, ExitAuthError)
	}
}

func TestExitCodeContextCanceled(t *testing.T) {
	code := exitCode(context.Canceled)
	if code != ExitError {
		t.Errorf("exitCode(context.Canceled): got %d, want %d", code, ExitError)
	}
}

func TestValidOutputFormats(t *testing.T) {
	valid := []string{"text", "json", "markdown"}
	for _, f := range valid {
		if !validOutputFormats[f] {
			t.Errorf("format %q should be valid", f)
		}
	}
}

func TestInvalidOutputFormat(t *testing.T) {
	invalid := []string{"xml", "csv", "yaml", ""}
	for _, f := range invalid {
		if validOutputFormats[f] {
			t.Errorf("format %q should be invalid", f)
		}
	}
}

func TestVerboseFlagRegistered(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("verbose")
	if flag == nil {
		t.Error("--verbose flag should be registered on root command")
	}
	if flag.DefValue != "false" {
		t.Errorf("--verbose default: got %q, want %q", flag.DefValue, "false")
	}
}
