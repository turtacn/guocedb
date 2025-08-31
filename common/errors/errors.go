// Package errors provides a unified error handling mechanism for guocedb.
package errors

import (
	"fmt"
	"net/http"

	"github.com/turtacn/guocedb/common/constants"
)

// Error represents a custom error with a code, message, and an underlying error.
type Error struct {
	Code    int
	Message string
	Err     error
}

// New creates a new custom error.
func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Newf creates a new custom error with a formatted message.
func Newf(code int, format string, args ...interface{}) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Wrap wraps an existing error with a new custom error.
func Wrap(err error, code int, message string) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

// Wrapf wraps an existing error with a new custom error and a formatted message.
func Wrapf(err error, code int, format string, args ...interface{}) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), Err: err}
}

// Error returns the string representation of the error.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error for error chaining.
func (e *Error) Unwrap() error {
	return e.Err
}

// ToHTTPStatusCode maps an error code to an HTTP status code.
// This is a simplified mapping.
func ToHTTPStatusCode(code int) int {
	switch code {
	case constants.ErrCodeSyntax:
		return http.StatusBadRequest
	case constants.ErrCodeRuntime:
		return http.StatusInternalServerError
	case constants.ErrCodeSystem:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// GetLocalizationKey would return a key for i18n, placeholder for now.
func (e *Error) GetLocalizationKey() string {
	// In a real implementation, this would return a key like "error.syntax.1001"
	// to be used with a localization library.
	return fmt.Sprintf("error.%d", e.Code)
}

// Predefined errors
var (
	ErrSyntax      = New(constants.ErrCodeSyntax, "Syntax error")
	ErrRuntime     = New(constants.ErrCodeRuntime, "Runtime error")
	ErrSystem      = New(constants.ErrCodeSystem, "System error")
	ErrNotImplemented = New(constants.ErrCodeSystem, "Not implemented")
)
