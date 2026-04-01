package cmd

import (
	"context"
	"errors"
	"testing"

	clierrors "github.com/natikgadzhi/cli-kit/errors"
)

func TestExitCodeSuccess(t *testing.T) {
	code := exitCode(nil)
	if code != clierrors.ExitSuccess {
		t.Errorf("exitCode(nil): got %d, want %d", code, clierrors.ExitSuccess)
	}
}

func TestExitCodeGeneralError(t *testing.T) {
	code := exitCode(errors.New("some error"))
	if code != clierrors.ExitError {
		t.Errorf("exitCode(general error): got %d, want %d", code, clierrors.ExitError)
	}
}

func TestExitCodeAuthError(t *testing.T) {
	err := clierrors.WrapAuth(errors.New("test"), "test auth error", "")
	code := exitCode(err)
	if code != clierrors.ExitAuthError {
		t.Errorf("exitCode(auth error): got %d, want %d", code, clierrors.ExitAuthError)
	}
}

func TestExitCodeWrappedAuthError(t *testing.T) {
	inner := clierrors.WrapAuth(errors.New("test"), "inner auth error", "")
	wrapped := errors.Join(errors.New("wrapper"), inner)
	code := exitCode(wrapped)
	if code != clierrors.ExitAuthError {
		t.Errorf("exitCode(wrapped auth error): got %d, want %d", code, clierrors.ExitAuthError)
	}
}

func TestExitCodeContextCanceled(t *testing.T) {
	code := exitCode(context.Canceled)
	if code != clierrors.ExitError {
		t.Errorf("exitCode(context.Canceled): got %d, want %d", code, clierrors.ExitError)
	}
}

func TestExitCodeCLIError(t *testing.T) {
	err := clierrors.NewCLIError(clierrors.ExitAuthError, "access denied")
	code := exitCode(err)
	if code != clierrors.ExitAuthError {
		t.Errorf("exitCode(CLIError): got %d, want %d", code, clierrors.ExitAuthError)
	}
}

func TestDebugFlagRegistered(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("debug")
	if flag == nil {
		t.Error("--debug flag should be registered on root command")
	}
	if flag.DefValue != "false" {
		t.Errorf("--debug default: got %q, want %q", flag.DefValue, "false")
	}
}

func TestOutputFlagRegistered(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Error("-o/--output flag should be registered on root command")
	}
}

func TestDerivedFlagRegistered(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("derived")
	if flag == nil {
		t.Error("-d/--derived flag should be registered on root command")
	}
}
