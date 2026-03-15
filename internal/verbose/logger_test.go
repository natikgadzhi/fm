package verbose

import (
	"testing"
)

func TestEnabledDefaultFalse(t *testing.T) {
	// Reset state.
	mu.Lock()
	enabled = false
	mu.Unlock()

	if Enabled() {
		t.Error("verbose should be disabled by default")
	}
}

func TestEnableAndEnabled(t *testing.T) {
	// Reset state.
	mu.Lock()
	enabled = false
	mu.Unlock()

	Enable()
	if !Enabled() {
		t.Error("verbose should be enabled after Enable()")
	}

	// Reset for other tests.
	mu.Lock()
	enabled = false
	mu.Unlock()
}

func TestLogNoOpWhenDisabled(t *testing.T) {
	// Reset state.
	mu.Lock()
	enabled = false
	mu.Unlock()

	// This should not panic or produce output.
	Log("test message %s", "hello")
}
