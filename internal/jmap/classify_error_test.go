package jmap

import (
	"errors"
	"testing"

	"github.com/natikgadzhi/fm/internal/auth"
)

func TestClassifyErrorNil(t *testing.T) {
	if classifyError(nil) != nil {
		t.Error("classifyError(nil) should return nil")
	}
}

func TestClassifyError401(t *testing.T) {
	err := classifyError(errors.New("HTTP 401 Unauthorized"))
	var authErr *auth.AuthError
	if !errors.As(err, &authErr) {
		t.Fatal("401 error should be classified as AuthError")
	}
	if authErr.Message == "" {
		t.Error("AuthError message should not be empty")
	}
}

func TestClassifyErrorUnauthorized(t *testing.T) {
	err := classifyError(errors.New("server returned Unauthorized"))
	var authErr *auth.AuthError
	if !errors.As(err, &authErr) {
		t.Fatal("Unauthorized error should be classified as AuthError")
	}
}

func TestClassifyErrorNetworkDNS(t *testing.T) {
	err := classifyError(errors.New("dial tcp: lookup api.fastmail.com: no such host"))
	if err == nil {
		t.Fatal("should not return nil")
	}
	errStr := err.Error()
	if errStr == "dial tcp: lookup api.fastmail.com: no such host" {
		t.Error("error should be wrapped with a helpful message")
	}
}

func TestClassifyErrorConnectionRefused(t *testing.T) {
	err := classifyError(errors.New("dial tcp 127.0.0.1:443: connection refused"))
	if err == nil {
		t.Fatal("should not return nil")
	}
	errStr := err.Error()
	if errStr == "dial tcp 127.0.0.1:443: connection refused" {
		t.Error("error should be wrapped with a helpful message")
	}
}

func TestClassifyErrorGeneric(t *testing.T) {
	original := errors.New("some other error")
	classified := classifyError(original)
	if classified != original {
		t.Error("generic errors should be returned as-is")
	}
}

func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"no such host", true},
		{"connection refused", true},
		{"connection reset by peer", true},
		{"network is unreachable", true},
		{"i/o timeout", true},
		{"dial tcp 127.0.0.1:443: connection refused", true},
		{"TLS handshake timeout", true},
		{"some other error", false},
		{"HTTP 500 Internal Server Error", false},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			got := isNetworkError(errors.New(tt.msg))
			if got != tt.want {
				t.Errorf("isNetworkError(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestIsNetworkErrorNil(t *testing.T) {
	if isNetworkError(nil) {
		t.Error("isNetworkError(nil) should return false")
	}
}
