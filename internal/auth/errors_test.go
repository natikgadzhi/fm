package auth

import (
	"errors"
	"testing"
)

func TestAuthErrorMessage(t *testing.T) {
	err := &AuthError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test error")
	}
}

func TestAuthErrorWithWrapped(t *testing.T) {
	inner := errors.New("inner error")
	err := &AuthError{Message: "outer", Err: inner}
	if err.Error() != "outer: inner error" {
		t.Errorf("Error() = %q, want %q", err.Error(), "outer: inner error")
	}
}

func TestAuthErrorUnwrap(t *testing.T) {
	inner := errors.New("inner error")
	err := &AuthError{Message: "outer", Err: inner}
	if !errors.Is(err, inner) {
		t.Error("errors.Is should find the inner error")
	}
}

func TestAuthErrorAs(t *testing.T) {
	err := &AuthError{Message: "auth failed"}
	wrapped := errors.Join(errors.New("context"), err)

	var authErr *AuthError
	if !errors.As(wrapped, &authErr) {
		t.Error("errors.As should find AuthError")
	}
	if authErr.Message != "auth failed" {
		t.Errorf("Message = %q, want %q", authErr.Message, "auth failed")
	}
}

func TestErrNoTokenIsAuthError(t *testing.T) {
	var authErr *AuthError
	if !errors.As(ErrNoToken, &authErr) {
		t.Error("ErrNoToken should be an *AuthError")
	}
}
