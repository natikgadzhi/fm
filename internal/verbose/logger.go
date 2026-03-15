// Package verbose provides a simple debug logger that writes to stderr
// when verbose mode is enabled via the --verbose flag.
package verbose

import (
	"fmt"
	"os"
	"sync"
)

// logger is the package-level Logger instance.
var (
	mu      sync.RWMutex
	enabled bool
)

// Enable turns on verbose logging.
func Enable() {
	mu.Lock()
	defer mu.Unlock()
	enabled = true
}

// Enabled reports whether verbose logging is on.
func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

// Log writes a formatted message to stderr if verbose mode is enabled.
// The message is prefixed with "[debug] ".
func Log(format string, args ...any) {
	if !Enabled() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[debug] %s\n", msg)
}
