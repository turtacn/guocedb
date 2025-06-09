// Package errors defines unified error types and handling mechanisms for the Guocedb project.
// This file serves as the central hub for error management, ensuring that all errors
// across modules are consistently structured, logged, and handled.
package errors

import (
	"fmt"     // Import fmt for error formatting.
	"runtime"
	"strings" // Import strings for potential string manipulation in error messages.

	"github.com/turtacn/guocedb/common/types/enum" // Import enum for error categorization.
)

// GuocedbError is the base error type for all errors within the Guocedb project.
// It encapsulates the error type, an internal error code, and the original error message.
type GuocedbError struct {
	Type     enum.GuocedbErrorType // Type categorizes the error (e.g., ErrConfiguration, ErrStorage).
	Code     int                   // Code provides a more specific internal error code for detailed tracking.
	Message  string                // Message is a human-readable error description.
	Cause    error                 // Cause holds the underlying error, if any, allowing for error chaining.
	Severity enum.LogLevel         // Severity indicates the recommended logging level for this error.
	stack    string                // Internal: Stack trace at the point the error was created

}

// NewGuocedbError creates a new GuocedbError instance.
// It takes an error type, an optional error code, a message, and an optional cause error.
// The severity is automatically set based on the error type, but can be overridden.
func NewGuocedbError(errType enum.GuocedbErrorType, code int, msg string, cause error) *GuocedbError {
	// Determine default severity based on error type.
	severity := enum.LogLevelError
	switch errType {
	case enum.ErrConfiguration, enum.ErrNetwork, enum.ErrStorage, enum.ErrCompute, enum.ErrTransaction, enum.ErrSecurity, enum.ErrProtocol:
		severity = enum.LogLevelFatal // These errors typically indicate critical failures.
	case enum.ErrInvalidArgument, enum.ErrNotFound, enum.ErrAlreadyExists, enum.ErrPermissionDenied, enum.ErrNotSupported:
		severity = enum.LogLevelWarn // These errors are often due to client-side issues or expected conditions.
	}

	return &GuocedbError{
		Type:     errType,
		Code:     code,
		Message:  msg,
		Cause:    cause,
		Severity: severity,
		stack: getStackTrace(),
	}
}

// Error implements the error interface for GuocedbError.
// It provides a formatted string representation of the error, including its type, code, and message.
func (e *GuocedbError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] ErrorCode:%d - %s", e.Type, e.Code, e.Message))
	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf(": %v", e.Cause)) // Append the cause if it exists.
	}
	return sb.String()
}

// Unwrap returns the underlying error (cause) of the GuocedbError.
// This method is crucial for error chaining and is used by `errors.Is` and `errors.As`.
func (e *GuocedbError) Unwrap() error {
	return e.Cause
}

// Is checks if the target error is of a specific GuocedbErrorType.
// This allows for type-based error handling without checking the exact instance.
func Is(err error, targetType enum.GuocedbErrorType) bool {
	if ge, ok := err.(*GuocedbError); ok {
		return ge.Type == targetType
	}
	return false
}

// As finds the first error in the chain that matches *GuocedbError and assigns it to target.
// This function helps to unwrap and inspect GuocedbError details from a complex error chain.
func As(err error, target **GuocedbError) bool {
	if err == nil || target == nil {
		return false
	}
	for err != nil {
		if ge, ok := err.(*GuocedbError); ok {
			*target = ge
			return true
		}
		err = Unwrap(err) // Custom Unwrap function (or use standard library errors.Unwrap in Go 1.13+)
	}
	return false
}

// Unwrap is a helper function to unwrap errors.
// This is a simplified version; for Go 1.13+, consider using errors.Unwrap directly.
func Unwrap(err error) error {
	if uw, ok := err.(interface{ Unwrap() error }); ok {
		return uw.Unwrap()
	}
	return nil
}

// Wrap wraps an existing error with a GuocedbError, adding a context message.
// If the original error is already a GuocedbError, it creates a new GuocedbError
// that wraps the original one and includes the new context.
// If the original error is a standard error, it creates a new GuocedbError with
// the standard error as the underlying error.
func Wrap(err error, cause string) error {
	if err == nil {
		return nil
	}

	// Capture the stack trace at the point of wrapping
	currentStack := getStackTrace()

	if gErr, ok := err.(*GuocedbError); ok {
		// If the original error is already a GuocedbError, we enrich it.
		// We inherit Type, Code, and Severity from the original error,
		// and set the original gErr as the Cause for the new error.
		return &GuocedbError{
			Type:     gErr.Type,
			Code:     gErr.Code,
			Message:  fmt.Sprintf("%s: %s", cause, gErr.Message), // Prepend new context
			Cause:    gErr, // The original GuocedbError is the cause
			Severity: gErr.Severity,
			stack:    currentStack, // New stack for the wrapping point
		}
	}

	// If the original error is a standard `error`, we create a new GuocedbError
	// with a default internal type/code/severity.
	return &GuocedbError{
		Type:     enum.ErrUnknown,    // Default to internal error type
		Code:     CodeInternalError,         // Default to unknown sub-code
		Message:  cause,               // The new message is the provided cause
		Cause:    err,                 // The original standard error is the cause
		Severity: enum.LogLevelError,  // Default severity for wrapped standard errors
		stack:    currentStack,        // New stack for the wrapping point
	}
}

// getStackTrace retrieves the stack trace at the point of the function call.
func getStackTrace() string {
	var sb strings.Builder
	// Skip 2 frames: getStackTrace itself and the caller of getStackTrace (e.g., NewGuocedbError or Wrap)
	for i := 2; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Only capture relevant project specific stack frames
		// This helps filter out noisy runtime/library calls from the stack trace
		if strings.Contains(file, "guocedb/") {
			funcName := runtime.FuncForPC(pc).Name()
			// Remove package path prefix for cleaner output
			if idx := strings.LastIndex(funcName, "/"); idx != -1 {
				funcName = funcName[idx+1:]
			}
			fmt.Fprintf(&sb, "\t%s:%d %s\n", file, line, funcName)
		}
	}
	return sb.String()
}

// RegisterErrorCode can be used to register a specific error code with a corresponding message.
// This can be expanded to a global error registry for consistent error message retrieval
// based on error codes. (Placeholder for future expansion)
func RegisterErrorCode(code int, message string) {
	// TODO: Implement a global error code registry for consistent error messages.
	// This would allow retrieving error messages based on a given error code,
	// rather than hardcoding them at the NewGuocedbError call site.
}

// Predefined error codes for common scenarios.
const (
	// Common error codes
	CodeInternalError         = 1000 // Generic internal server error.
	CodeInvalidInput          = 1001 // Input provided is invalid.
	CodeResourceNotFound      = 1002 // Requested resource does not exist.
	CodeResourceAlreadyExists = 1003 // Resource attempting to be created already exists.
	CodePermissionDenied      = 1004 // Operation denied due to insufficient permissions.
	CodeFeatureNotSupported   = 1005 // The requested feature is not yet supported.

	// Configuration error codes
	CodeConfigLoadFailed   = 2000 // Failed to load configuration.
	CodeConfigInvalidValue = 2001 // Invalid value found in configuration.

	// Network error codes
	CodeNetworkConnectionFailed = 3000 // Failed to establish a network connection.
	CodeNetworkReadFailed       = 3001 // Failed to read from network.
	CodeNetworkWriteFailed      = 3002 // Failed to write to network.

	// Storage error codes
	CodeStorageIOError        = 4000 // Generic storage I/O error.
	CodeStorageCorruption     = 4001 // Data corruption detected in storage.
	CodeStorageEngineInitFail = 4002 // Failed to initialize storage engine.

	// Compute error codes
	CodeSyntaxError    = 5000 // SQL syntax error.
	CodeSemanticError  = 5001 // SQL semantic error (e.g., table not found).
	CodeOptimizerError = 5002 // Query optimization failed.
	CodeExecutionError = 5003 // Query execution failed.

	// Security error codes
	CodeAuthenticationFailed = 6000 // User authentication failed.
	CodeAuthorizationFailed  = 6001 // User authorization failed.
	CodeEncryptionFailed     = 6002 // Data encryption/decryption failed.

	// Catalog error codes
	CodeCatalogNotFound     = 7000 // Metadata catalog not found.
	CodeCatalogUpdateFailed = 7001 // Failed to update metadata catalog.
)
