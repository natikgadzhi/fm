package jmap

import (
	"errors"
	"fmt"
	"testing"
)

func TestPartialResultErrorMessage(t *testing.T) {
	inner := fmt.Errorf("rate limited")
	err := &PartialResultError{
		Emails:  []Email{{Id: "e1"}, {Id: "e2"}},
		Fetched: 2,
		Total:   5,
		Err:     inner,
	}

	msg := err.Error()
	if msg != "partial result: fetched 2 of 5 emails: rate limited" {
		t.Errorf("unexpected error message: %s", msg)
	}
}

func TestPartialResultErrorUnwrap(t *testing.T) {
	inner := fmt.Errorf("connection reset")
	err := &PartialResultError{
		Emails:  []Email{{Id: "e1"}},
		Fetched: 1,
		Total:   3,
		Err:     inner,
	}

	if !errors.Is(err, inner) {
		t.Error("errors.Is should match the wrapped error")
	}
}

func TestPartialResultErrorAs(t *testing.T) {
	inner := fmt.Errorf("server error")
	original := &PartialResultError{
		Emails:  []Email{{Id: "e1"}, {Id: "e2"}, {Id: "e3"}},
		Fetched: 3,
		Total:   10,
		Err:     inner,
	}

	// Wrap it in another error.
	wrapped := fmt.Errorf("fetch failed: %w", original)

	var partialErr *PartialResultError
	if !errors.As(wrapped, &partialErr) {
		t.Fatal("errors.As should find PartialResultError in wrapped error")
	}

	if partialErr.Fetched != 3 {
		t.Errorf("expected Fetched=3, got %d", partialErr.Fetched)
	}
	if partialErr.Total != 10 {
		t.Errorf("expected Total=10, got %d", partialErr.Total)
	}
	if len(partialErr.Emails) != 3 {
		t.Errorf("expected 3 emails, got %d", len(partialErr.Emails))
	}
}

func TestPartialResultErrorZeroEmails(t *testing.T) {
	err := &PartialResultError{
		Emails:  nil,
		Fetched: 0,
		Total:   5,
		Err:     fmt.Errorf("total failure"),
	}

	msg := err.Error()
	if msg != "partial result: fetched 0 of 5 emails: total failure" {
		t.Errorf("unexpected error message: %s", msg)
	}
}
