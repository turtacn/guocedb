package errors

import (
	"fmt"
)

// Error represents a custom error with a code and message.
// It implements the standard 'error' interface.
type Error struct {
	Code    int
	Message string
	cause   error
}

// New creates a new custom error.
func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap creates a new error that wraps an existing error, adding more context.
func Wrap(err error, code int, message string) *Error {
	return &Error{Code: code, Message: message, cause: err}
}

// Error returns the string representation of the error.
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

// Unwrap returns the wrapped error, allowing it to work with the standard
// library's errors.Is and errors.As functions.
func (e *Error) Unwrap() error {
	return e.cause
}

// TODO: Add support for internationalization (i18n) of error messages.
// TODO: Add a function to map internal error codes to HTTP status codes for the REST API.
