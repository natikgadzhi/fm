package auth

// ExitCodeAuth is the exit code used for authentication errors.
const ExitCodeAuth = 2

// AuthError represents an authentication-related error that should cause
// the process to exit with ExitCodeAuth (2).
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
