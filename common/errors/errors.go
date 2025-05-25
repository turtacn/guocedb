// Package errors defines unified error types and provides utility functions
// for error handling throughout the Guocedb project.
//
// 统一错误定义包，旨在提供 Guocedb 项目中所有错误类型及其处理工具，
// 提升错误信息的标准化和可追溯性。
package errors

import (
	"fmt"
	"net/http" // Potentially for mapping errors to HTTP status codes in API layer
)

// Error represents a Guocedb specific error.
// It embeds the built-in error interface and adds a Code for programmatic handling
// and a PublicMessage for user-facing error messages.
// Error 结构体表示 Guocedb 特定的错误。
// 它嵌入了 Go 内置的 error 接口，并添加了用于程序化处理的错误码（Code）
// 和用于用户界面显示的消息（PublicMessage）。
type Error struct {
	// Code is an internal error code that can be used for programmatic error handling
	// and localization.
	// 错误码，内部使用，便于程序化处理和国际化。
	Code ErrorCode
	// PublicMessage is a user-friendly message that can be displayed to the end-user.
	// 用户友好的消息，可以直接显示给最终用户。
	PublicMessage string
	// Inner (optional) is the wrapped error, allowing for error chaining.
	// 内部错误（可选），用于错误链，保留原始错误信息。
	Inner error
}

// Error implements the built-in error interface.
// It returns a detailed error string, prioritizing the inner error if present,
// otherwise falling back to the public message and code.
// Error 方法实现了 Go 内置的 error 接口。
// 它返回一个详细的错误字符串，如果存在内部错误，则优先显示内部错误，
// 否则显示公共消息和错误码。
func (e *Error) Error() string {
	if e.Inner != nil {
		return fmt.Sprintf("code %d (%s): %s, inner: %s", e.Code, e.PublicMessage, e.Inner.Error(), e.Code.String())
	}
	return fmt.Sprintf("code %d (%s): %s", e.Code, e.PublicMessage, e.Code.String())
}

// Unwrap returns the wrapped error, if any, allowing for error chain traversal.
// Unwrap 方法返回包装的内部错误（如果存在），支持错误链的遍历。
func (e *Error) Unwrap() error {
	return e.Inner
}

// New creates a new Guocedb error with a specific error code and an optional inner error.
// The public message is derived from the error code's default string.
// New 函数创建一个新的 Guocedb 错误，包含指定的错误码和可选的内部错误。
// 公共消息将从错误码的默认字符串中获取。
func New(code ErrorCode, inner ...error) *Error {
	var innerErr error
	if len(inner) > 0 {
		innerErr = inner[0]
	}
	return &Error{
		Code:          code,
		PublicMessage: code.String(), // Default public message is the code's string representation
		Inner:         innerErr,
	}
}

// Newf creates a new Guocedb error with a specific error code and a formatted public message.
// It also accepts an optional inner error.
// Newf 函数创建一个新的 Guocedb 错误，包含指定的错误码和格式化的公共消息。
// 它也接受一个可选的内部错误。
func Newf(code ErrorCode, format string, args ...interface{}) *Error {
	// Check if the last argument is an error to unwrap it
	var innerErr error
	if len(args) > 0 {
		if err, ok := args[len(args)-1].(error); ok {
			innerErr = err
			args = args[:len(args)-1] // Remove the error from args for fmt.Sprintf
		}
	}

	return &Error{
		Code:          code,
		PublicMessage: fmt.Sprintf(format, args...),
		Inner:         innerErr,
	}
}

// IsGuocedbError checks if an error is a Guocedb error and optionally matches a specific code.
// If targetCode is NoError, it only checks if it's a Guocedb error.
// IsGuocedbError 函数检查一个错误是否为 Guocedb 错误，并可选地检查是否与指定的错误码匹配。
// 如果 targetCode 是 NoError，则仅检查是否为 Guocedb 错误。
func IsGuocedbError(err error, targetCode ...ErrorCode) bool {
	if err == nil {
		return false
	}
	ge, ok := err.(*Error)
	if !ok {
		return false
	}
	if len(targetCode) > 0 && targetCode[0] != NoError {
		return ge.Code == targetCode[0]
	}
	return true
}

// GetGuocedbErrorCode extracts the ErrorCode from a Guocedb error.
// Returns NoError if the error is not a Guocedb error or nil.
// GetGuocedbErrorCode 函数从 Guocedb 错误中提取错误码。
// 如果错误不是 Guocedb 错误或为 nil，则返回 NoError。
func GetGuocedbErrorCode(err error) ErrorCode {
	if err == nil {
		return NoError
	}
	if ge, ok := err.(*Error); ok {
		return ge.Code
	}
	return NoError
}

// MapToHTTPStatusCode provides a mapping from Guocedb ErrorCode to HTTP status codes.
// This is useful when exposing errors via a RESTful API.
// MapToHTTPStatusCode 函数提供 Guocedb 错误码到 HTTP 状态码的映射。
// 在通过 RESTful API 暴露错误时非常有用。
func MapToHTTPStatusCode(code ErrorCode) int {
	switch code {
	case ErrInvalidArgument, ErrSyntax:
		return http.StatusBadRequest // 400 Bad Request
	case ErrNotFound:
		return http.StatusNotFound // 404 Not Found
	case ErrUnauthorized:
		return http.StatusUnauthorized // 401 Unauthorized
	case ErrPermissionDenied:
		return http.StatusForbidden // 403 Forbidden
	case ErrAlreadyExists, ErrDuplicateEntry:
		return http.StatusConflict // 409 Conflict
	case ErrInternal, ErrStorageFailure, ErrTransactionFailed, ErrUnknown:
		return http.StatusInternalServerError // 500 Internal Server Error
	case ErrServiceUnavailable:
		return http.StatusServiceUnavailable // 503 Service Unavailable
	default:
		return http.StatusInternalServerError // Default to 500 for unhandled errors
	}
}
