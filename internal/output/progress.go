package output

import (
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/term"
)

// Progress is a simple progress counter that writes status updates to stderr.
// It only renders output when stderr is connected to a TTY.
//
// Usage:
//
//	p := NewProgress("Fetching emails", 50)
//	for i := range 50 {
//	    // do work
//	    p.Inc()
//	}
//	p.Done()
type Progress struct {
	mu      sync.Mutex
	label   string
	total   int
	current int
	w       io.Writer
	isTTY   bool
}

// NewProgress creates a new progress counter with the given label and total count.
// It automatically detects whether stderr is a TTY and only writes output if so.
func NewProgress(label string, total int) *Progress {
	return newProgress(label, total, os.Stderr, isStderrTTY())
}

// newProgress is the internal constructor, allowing injection of writer and TTY flag for testing.
func newProgress(label string, total int, w io.Writer, isTTY bool) *Progress {
	p := &Progress{
		label: label,
		total: total,
		w:     w,
		isTTY: isTTY,
	}
	if p.isTTY && total > 0 {
		// Print initial state.
		fmt.Fprintf(p.w, "\r%s... [0/%d]", p.label, p.total)
	}
	return p
}

// Inc increments the progress counter by one and updates the display.
func (p *Progress) Inc() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current++
	if !p.isTTY || p.total == 0 {
		return
	}

	// Overwrite the current line with updated progress.
	fmt.Fprintf(p.w, "\r%s... [%d/%d]", p.label, p.current, p.total)
}

// Done finalizes the progress display by clearing the line.
func (p *Progress) Done() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isTTY || p.total == 0 {
		return
	}

	// Clear the progress line so it doesn't pollute the output.
	// Use a carriage return followed by spaces to overwrite, then return again.
	clearLen := len(p.label) + 20 // generous padding for "... [999/999]"
	fmt.Fprintf(p.w, "\r%*s\r", clearLen, "")
}

// isStderrTTY checks whether stderr is connected to a terminal.
func isStderrTTY() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}
