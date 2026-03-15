package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestProgressTTYOutput(t *testing.T) {
	var buf bytes.Buffer

	// Simulate TTY mode with total of 3.
	p := newProgress("Fetching emails", 3, &buf, true)

	// Should have printed initial state.
	out := buf.String()
	if !strings.Contains(out, "Fetching emails... [0/3]") {
		t.Errorf("expected initial progress, got: %q", out)
	}

	p.Inc()
	out = buf.String()
	if !strings.Contains(out, "[1/3]") {
		t.Errorf("expected [1/3] in output, got: %q", out)
	}

	p.Inc()
	out = buf.String()
	if !strings.Contains(out, "[2/3]") {
		t.Errorf("expected [2/3] in output, got: %q", out)
	}

	p.Inc()
	out = buf.String()
	if !strings.Contains(out, "[3/3]") {
		t.Errorf("expected [3/3] in output, got: %q", out)
	}

	p.Done()
	// After Done(), the line should be cleared.
	out = buf.String()
	// The last thing written should end with \r (clearing the line).
	if !strings.HasSuffix(out, "\r") {
		t.Errorf("expected output to end with \\r after Done(), got: %q", out)
	}
}

func TestProgressNonTTYOutput(t *testing.T) {
	var buf bytes.Buffer

	// Simulate non-TTY mode.
	p := newProgress("Fetching emails", 3, &buf, false)

	p.Inc()
	p.Inc()
	p.Inc()
	p.Done()

	// In non-TTY mode, nothing should be written.
	if buf.Len() != 0 {
		t.Errorf("expected no output in non-TTY mode, got: %q", buf.String())
	}
}

func TestProgressZeroTotal(t *testing.T) {
	var buf bytes.Buffer

	// Zero total should produce no output even in TTY mode.
	p := newProgress("Fetching", 0, &buf, true)
	p.Inc()
	p.Done()

	if buf.Len() != 0 {
		t.Errorf("expected no output for zero total, got: %q", buf.String())
	}
}

func TestProgressConcurrentInc(t *testing.T) {
	var buf bytes.Buffer

	p := newProgress("Loading", 100, &buf, true)

	// Run concurrent increments — this should not panic or race.
	done := make(chan struct{})
	for range 10 {
		go func() {
			for range 10 {
				p.Inc()
			}
			done <- struct{}{}
		}()
	}

	for range 10 {
		<-done
	}

	p.Done()

	// Verify counter reached 100.
	out := buf.String()
	if !strings.Contains(out, "[100/100]") {
		t.Errorf("expected [100/100] in concurrent output, got: %q", out)
	}
}
