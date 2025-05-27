package errors

import (
	"fmt"
	"runtime"
	"strings"

	goerrors "errors" // Alias standard errors package
)

// ErrorCode defines the type for GuoceDB specific error codes.
// ErrorCode defines the type for GuoceDB specific error codes.
type ErrorCode int

// Predefined GuoceDB error codes.
// Predefined GuoceDB error codes.
const (
	// ErrCodeUnknown Default/unknown error.
	// ErrCodeUnknown Default/unknown error.
	ErrCodeUnknown ErrorCode = iota

	// --- Common Errors (1-99) ---
	// --- Common Errors (1-99) ---
	// ErrCodeInternal Internal error, indicates a bug or unexpected state.
	// ErrCodeInternal Internal error, indicates a bug or unexpected state.
	ErrCodeInternal
	// ErrCodeNotImplemented Feature or functionality not yet implemented.
	// ErrCodeNotImplemented Feature or functionality not yet implemented.
	ErrCodeNotImplemented
	// ErrCodeInvalidArgument Invalid argument provided to a function or operation.
	// ErrCodeInvalidArgument Invalid argument provided to a function or operation.
	ErrCodeInvalidArgument
	// ErrCodeNotFound Resource (database, table, key, etc.) not found.
	// ErrCodeNotFound Resource (database, table, key, etc.) not found.
	ErrCodeNotFound
	// ErrCodeAlreadyExists Resource already exists.
	// ErrCodeAlreadyExists Resource already exists.
	ErrCodeAlreadyExists
	// ErrCodePermissionDenied Operation denied due to insufficient permissions.
	// ErrCodePermissionDenied Operation denied due to insufficient permissions.
	ErrCodePermissionDenied
	// ErrCodeIO Error during Input/Output operation (e.g., disk read/write).
	// ErrCodeIO Error during Input/Output operation (e.g., disk read/write).
	ErrCodeIO
	// ErrCodeSerialization Error during data serialization or deserialization.
	// ErrCodeSerialization Error during data serialization or deserialization.
	ErrCodeSerialization
	// ErrCodeTimeout Operation timed out.
	// ErrCodeTimeout Operation timed out.
	ErrCodeTimeout
	// ErrCodeCancelled Operation was cancelled.
	// ErrCodeCancelled Operation was cancelled.
	ErrCodeCancelled

	// --- Type System Errors (100-199) ---
	// --- Type System Errors (100-199) ---
	// ErrCodeTypeMismatch Incompatible data types used in an operation.
	// ErrCodeTypeMismatch Incompatible data types used in an operation.
	ErrCodeTypeMismatch
	// ErrCodeConversion Failed to convert value from one type to another.
	// ErrCodeConversion Failed to convert value from one type to another.
	ErrCodeConversion
	// ErrCodeComparison Failed to compare two values.
	// ErrCodeComparison Failed to compare two values.
	ErrCodeComparison
	// ErrCodeOverflow Numeric value overflowed its type's capacity.
	// ErrCodeOverflow Numeric value overflowed its type's capacity.
	ErrCodeOverflow
	// ErrCodeDivideByZero Division by zero attempted.
	// ErrCodeDivideByZero Division by zero attempted.
	ErrCodeDivideByZero

	// --- Parser & Analyzer Errors (200-299) ---
	// --- Parser & Analyzer Errors (200-299) ---
	// ErrCodeSyntaxError SQL syntax error during parsing.
	// ErrCodeSyntaxError SQL syntax error during parsing.
	ErrCodeSyntaxError
	// ErrCodeSemanticError SQL semantic error during analysis (e.g., unknown column).
	// ErrCodeSemanticError SQL semantic error during analysis (e.g., unknown column).
	ErrCodeSemanticError
	// ErrCodeUnsupportedFeature Unsupported SQL feature encountered.
	// ErrCodeUnsupportedFeature Unsupported SQL feature encountered.
	ErrCodeUnsupportedFeature

	// --- Execution Errors (300-399) ---
	// --- Execution Errors (300-399) ---
	// ErrCodeTxnConflict Transaction conflict detected (e.g., lock contention, write conflict).
	// ErrCodeTxnConflict Transaction conflict detected (e.g., lock contention, write conflict).
	ErrCodeTxnConflict
	// ErrCodeTxnAborted Transaction was aborted.
	// ErrCodeTxnAborted Transaction was aborted.
	ErrCodeTxnAborted
	// ErrCodeConstraintViolation Constraint violation (e.g., UNIQUE, NOT NULL, FOREIGN KEY).
	// ErrCodeConstraintViolation Constraint violation (e.g., UNIQUE, NOT NULL, FOREIGN KEY).
	ErrCodeConstraintViolation
	// ErrCodeExecutionError General error during query execution.
	// ErrCodeExecutionError General error during query execution.
	ErrCodeExecutionError

	// --- Storage Errors (400-499) ---
	// --- Storage Errors (400-499) ---
	// ErrCodeStorageEngine Storage engine specific error.
	// ErrCodeStorageEngine Storage engine specific error.
	ErrCodeStorageEngine
	// ErrCodeDataCorrupted Data corruption detected in storage.
	// ErrCodeDataCorrupted Data corruption detected in storage.
	ErrCodeDataCorrupted

	// --- Network/Protocol Errors (500-599) ---
	// --- Network/Protocol Errors (500-599) ---
	// ErrCodeNetwork Network communication error.
	// ErrCodeNetwork Network communication error.
	ErrCodeNetwork
	// ErrCodeProtocol Protocol violation or error during client/server communication.
	// ErrCodeProtocol Protocol violation or error during client/server communication.
	ErrCodeProtocol
)

// Error represents a GuoceDB error, encapsulating a code, message,
// an optional cause, and stack trace.
// Error represents a GuoceDB error, encapsulating a code, message,
// an optional cause, and stack trace.
type Error struct {
	code  ErrorCode
	msg   string
	cause error
	stack []uintptr // Program counters for stack trace
}

const maxStackDepth = 32

// New creates a new GuoceDB error with the given code and message.
// It captures the stack trace at the point of creation.
// New creates a new GuoceDB error with the given code and message.
// It captures the stack trace at the point of creation.
func New(code ErrorCode, msg string) *Error {
	return newError(code, msg, nil, 2) // Skip 2 frames: New and newError
}

// Newf creates a new GuoceDB error with the given code and formatted message.
// It captures the stack trace at the point of creation.
// Newf creates a new GuoceDB error with the given code and formatted message.
// It captures the stack trace at the point of creation.
func Newf(code ErrorCode, format string, args ...interface{}) *Error {
	return newError(code, fmt.Sprintf(format, args...), nil, 2) // Skip 2 frames: Newf and newError
}

// Wrap wraps an existing error with a new GuoceDB error context (code and message).
// If the original error is nil, Wrap returns nil.
// The original error is preserved as the 'cause'.
// It captures the stack trace at the point of wrapping.
// Wrap wraps an existing error with a new GuoceDB error context (code and message).
// If the original error is nil, Wrap returns nil.
// The original error is preserved as the 'cause'.
// It captures the stack trace at the point of wrapping.
func Wrap(err error, code ErrorCode, msg string) *Error {
	if err == nil {
		return nil
	}
	// If err is already a *Error, potentially reuse its stack?
	// For now, always capture a new stack at the wrapping point.
	// If err is already a *Error, potentially reuse its stack?
	// For now, always capture a new stack at the wrapping point.
	return newError(code, msg, err, 2) // Skip 2 frames: Wrap and newError
}

// Wrapf wraps an existing error with a new GuoceDB error context (code and formatted message).
// If the original error is nil, Wrapf returns nil.
// The original error is preserved as the 'cause'.
// It captures the stack trace at the point of wrapping.
// Wrapf wraps an existing error with a new GuoceDB error context (code and formatted message).
// If the original error is nil, Wrapf returns nil.
// The original error is preserved as the 'cause'.
// It captures the stack trace at the point of wrapping.
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return newError(code, fmt.Sprintf(format, args...), err, 2) // Skip 2 frames: Wrapf and newError
}

// newError is the internal constructor. Skip indicates how many frames to skip for stack trace.
// newError is the internal constructor. Skip indicates how many frames to skip for stack trace.
func newError(code ErrorCode, msg string, cause error, skip int) *Error {
	pcs := make([]uintptr, maxStackDepth)
	// Add 1 to skip to account for runtime.Callers itself
	// Add 1 to skip to account for runtime.Callers itself
	n := runtime.Callers(skip+1, pcs)
	e := &Error{
		code:  code,
		msg:   msg,
		cause: cause,
		stack: pcs[:n], // Keep only the frames recorded
	}
	return e
}

// Error implements the standard error interface.
// It provides a user-friendly message including the code, message, and the causal chain.
// Error implements the standard error interface.
// It provides a user-friendly message including the code, message, and the causal chain.
func (e *Error) Error() string {
	if e.cause != nil {
		// Include cause's message for clarity
		// Include cause's message for clarity
		return fmt.Sprintf("[%s] %s: %v", CodeToString(e.code), e.msg, e.cause)
	}
	return fmt.Sprintf("[%s] %s", CodeToString(e.code), e.msg)
}

// Message returns the specific message of this error, excluding the code and cause.
// Message returns the specific message of this error, excluding the code and cause.
func (e *Error) Message() string {
	return e.msg
}

// Code returns the GuoceDB error code associated with this error.
// Code returns the GuoceDB error code associated with this error.
func (e *Error) Code() ErrorCode {
	return e.code
}

// Unwrap provides compatibility with Go 1.13+ error wrapping (errors.Unwrap).
// Unwrap provides compatibility with Go 1.13+ error wrapping (errors.Unwrap).
func (e *Error) Unwrap() error {
	return e.cause
}

// StackTrace returns the captured stack trace formatted as a string.
// StackTrace returns the captured stack trace formatted as a string.
func (e *Error) StackTrace() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Error: %s\n", e.Error())) // Start with the error message itself
	frames := runtime.CallersFrames(e.stack)
	for {
		frame, more := frames.Next()
		// Example format: main.main /path/to/file.go:123
		// Example format: main.main /path/to/file.go:123
		builder.WriteString(fmt.Sprintf("- %s %s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	// Include stack trace of the cause if it's also a *Error
	// Include stack trace of the cause if it's also a *Error
	var causeErr *Error
	if goerrors.As(e.cause, &causeErr) {
		builder.WriteString("Caused by:\n")
		builder.WriteString(causeErr.StackTrace()) // Recursive call
	} else if e.cause != nil {
		// If cause is not *Error, just append its standard message
		// If cause is not *Error, just append its standard message
		builder.WriteString(fmt.Sprintf("Caused by: %v\n", e.cause))
	}
	return builder.String()
}

// Is checks if any error in the chain matches the target ErrorCode.
// It unwraps errors using errors.Unwrap.
// Is checks if any error in the chain matches the target ErrorCode.
// It unwraps errors using errors.Unwrap.
func Is(err error, targetCode ErrorCode) bool {
	if err == nil {
		return false
	}
	var e *Error
	if goerrors.As(err, &e) {
		if e.code == targetCode {
			return true
		}
		// Check the cause recursively
		// Check the cause recursively
		return Is(e.cause, targetCode)
	}
	// If it's not a *Error, it can't match our codes directly.
	// We could potentially check wrapped standard errors here if needed,
	// but the primary use case is checking our custom codes.
	// If it's not a *Error, it can't match our codes directly.
	// We could potentially check wrapped standard errors here if needed,
	// but the primary use case is checking our custom codes.

	// Check the cause if the current error is not *Error but wraps something
	// Check the cause if the current error is not *Error but wraps something
	cause := goerrors.Unwrap(err)
	if cause != nil {
		return Is(cause, targetCode)
	}

	return false
}

// GetCode extracts the GuoceDB ErrorCode from an error.
// It traverses the error chain and returns the code of the first *Error found.
// If no *Error is found, it returns ErrCodeUnknown.
// GetCode extracts the GuoceDB ErrorCode from an error.
// It traverses the error chain and returns the code of the first *Error found.
// If no *Error is found, it returns ErrCodeUnknown.
func GetCode(err error) ErrorCode {
	if err == nil {
		return ErrCodeUnknown // Or perhaps a specific "NoError" code if needed
	}
	var e *Error
	if goerrors.As(err, &e) {
		return e.code
	}
	// If the top-level error isn't a *Error, maybe its cause is?
	// This is less common if Wrap is used consistently, but possible.
	// cause := goerrors.Unwrap(err)
	// if cause != nil {
	// 	return GetCode(cause) // Recursive call
	// }
	return ErrCodeUnknown // Not a GuoceDB error or no code found in chain
}

// CodeToString converts an ErrorCode to its string representation.
// CodeToString converts an ErrorCode to its string representation.
func CodeToString(code ErrorCode) string {
	switch code {
	case ErrCodeInternal:
		return "INTERNAL_ERROR"
	case ErrCodeNotImplemented:
		return "NOT_IMPLEMENTED"
	case ErrCodeInvalidArgument:
		return "INVALID_ARGUMENT"
	case ErrCodeNotFound:
		return "NOT_FOUND"
	case ErrCodeAlreadyExists:
		return "ALREADY_EXISTS"
	case ErrCodePermissionDenied:
		return "PERMISSION_DENIED"
	case ErrCodeIO:
		return "IO_ERROR"
	case ErrCodeSerialization:
		return "SERIALIZATION_ERROR"
	case ErrCodeTimeout:
		return "TIMEOUT"
	case ErrCodeCancelled:
		return "CANCELLED"
	case ErrCodeTypeMismatch:
		return "TYPE_MISMATCH"
	case ErrCodeConversion:
		return "CONVERSION_ERROR"
	case ErrCodeComparison:
		return "COMPARISON_ERROR"
	case ErrCodeOverflow:
		return "OVERFLOW"
	case ErrCodeDivideByZero:
		return "DIVIDE_BY_ZERO"
	case ErrCodeSyntaxError:
		return "SYNTAX_ERROR"
	case ErrCodeSemanticError:
		return "SEMANTIC_ERROR"
	case ErrCodeUnsupportedFeature:
		return "UNSUPPORTED_FEATURE"
	case ErrCodeTxnConflict:
		return "TRANSACTION_CONFLICT"
	case ErrCodeTxnAborted:
		return "TRANSACTION_ABORTED"
	case ErrCodeConstraintViolation:
		return "CONSTRAINT_VIOLATION"
	case ErrCodeExecutionError:
		return "EXECUTION_ERROR"
	case ErrCodeStorageEngine:
		return "STORAGE_ENGINE_ERROR"
	case ErrCodeDataCorrupted:
		return "DATA_CORRUPTED"
	case ErrCodeNetwork:
		return "NETWORK_ERROR"
	case ErrCodeProtocol:
		return "PROTOCOL_ERROR"
	case ErrCodeUnknown:
		fallthrough
	default:
		return "UNKNOWN_ERROR"
	}
}

// --- Standard Error Compatibility ---
// --- Standard Error Compatibility ---

// Cause returns the underlying cause of the error, compatible with pkg/errors.
// Deprecated: Use errors.Unwrap directly from the standard library.
// func Cause(err error) error {
//    type causer interface {
//        Cause() error
//    }
//    for err != nil {
//        cause, ok := err.(causer)
//        if !ok {
//            break
//        }
//        c := cause.Cause()
//        if c == nil {
//            break // Should not happen with our *Error type, but defensive
//        }
//        err = c
//    }
//    // If we are using Go 1.13+, errors.Unwrap is preferred.
//    // This Cause function primarily exists for compatibility if older patterns
//    // or the pkg/errors library were being used.
//    // Standard library errors.Unwrap should be the primary way to traverse.
//    return err // Returns the root cause
// }

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
// This simply delegates to the standard library's errors.As.
// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
// This simply delegates to the standard library's errors.As.
func As(err error, target interface{}) bool {
	return goerrors.As(err, target)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
// This simply delegates to the standard library's errors.Unwrap.
// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
// This simply delegates to the standard library's errors.Unwrap.
func Unwrap(err error) error {
	return goerrors.Unwrap(err)
}
