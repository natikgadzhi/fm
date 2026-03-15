package jmap

import "fmt"

// PartialResultError indicates that a batch operation completed partially.
// It wraps the successfully fetched emails and the underlying error that
// caused the batch to stop. Callers should check for this error type using
// errors.As and display the partial results along with a warning.
type PartialResultError struct {
	// Emails contains the successfully fetched emails before the error occurred.
	Emails []Email

	// Fetched is the number of emails successfully retrieved.
	Fetched int

	// Total is the total number of emails that were requested.
	Total int

	// Err is the underlying error that caused the batch to stop.
	Err error
}

// Error implements the error interface.
func (e *PartialResultError) Error() string {
	return fmt.Sprintf("partial result: fetched %d of %d emails: %v", e.Fetched, e.Total, e.Err)
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *PartialResultError) Unwrap() error {
	return e.Err
}
