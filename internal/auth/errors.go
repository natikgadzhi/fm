package auth

// AuthError represents an authentication-related error that should cause
// the process to exit with a dedicated exit code (see cmd/root.go).
type AuthError struct {
	Message string
	Err     error
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *AuthError) Unwrap() error {
	return e.Err
}
