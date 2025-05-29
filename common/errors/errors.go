// Package errors 定义GuoceDB项目的统一错误处理机制
// Package errors defines unified error handling mechanism for GuoceDB project
package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorCategory 错误分类枚举
// ErrorCategory enumeration for error categories
type ErrorCategory int

const (
	// ErrCategorySystem 系统错误
	// ErrCategorySystem system errors
	ErrCategorySystem ErrorCategory = iota
	// ErrCategoryBusiness 业务错误
	// ErrCategoryBusiness business errors
	ErrCategoryBusiness
	// ErrCategoryNetwork 网络错误
	// ErrCategoryNetwork network errors
	ErrCategoryNetwork
	// ErrCategoryStorage 存储错误
	// ErrCategoryStorage storage errors
	ErrCategoryStorage
	// ErrCategoryAuth 认证错误
	// ErrCategoryAuth authentication errors
	ErrCategoryAuth
	// ErrCategoryPermission 权限错误
	// ErrCategoryPermission permission errors
	ErrCategoryPermission
	// ErrCategoryValidation 验证错误
	// ErrCategoryValidation validation errors
	ErrCategoryValidation
	// ErrCategoryTimeout 超时错误
	// ErrCategoryTimeout timeout errors
	ErrCategoryTimeout
)

// String 返回错误分类的字符串表示
// String returns string representation of error category
func (c ErrorCategory) String() string {
	switch c {
	case ErrCategorySystem:
		return "SYSTEM"
	case ErrCategoryBusiness:
		return "BUSINESS"
	case ErrCategoryNetwork:
		return "NETWORK"
	case ErrCategoryStorage:
		return "STORAGE"
	case ErrCategoryAuth:
		return "AUTH"
	case ErrCategoryPermission:
		return "PERMISSION"
	case ErrCategoryValidation:
		return "VALIDATION"
	case ErrCategoryTimeout:
		return "TIMEOUT"
	default:
		return fmt.Sprintf("UNKNOWN_CATEGORY(%d)", int(c))
	}
}

// ErrorSeverity 错误严重程度枚举
// ErrorSeverity enumeration for error severity levels
type ErrorSeverity int

const (
	// SeverityLow 低严重性
	// SeverityLow low severity
	SeverityLow ErrorSeverity = iota
	// SeverityMedium 中等严重性
	// SeverityMedium medium severity
	SeverityMedium
	// SeverityHigh 高严重性
	// SeverityHigh high severity
	SeverityHigh
	// SeverityCritical 严重
	// SeverityCritical critical severity
	SeverityCritical
)

// String 返回错误严重程度的字符串表示
// String returns string representation of error severity
func (s ErrorSeverity) String() string {
	switch s {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return fmt.Sprintf("UNKNOWN_SEVERITY(%d)", int(s))
	}
}

// ErrorCode 错误码定义
// ErrorCode error code definitions
type ErrorCode int

const (
	// 系统错误码 (1000-1999)
	// System error codes (1000-1999)
	ErrCodeSystemFailure      ErrorCode = 1000 // 系统故障 System failure
	ErrCodeOutOfMemory        ErrorCode = 1001 // 内存不足 Out of memory
	ErrCodeInternalError      ErrorCode = 1002 // 内部错误 Internal error
	ErrCodeConfigError        ErrorCode = 1003 // 配置错误 Configuration error
	ErrCodeInitializationFail ErrorCode = 1004 // 初始化失败 Initialization failure

	// 业务错误码 (2000-2999)
	// Business error codes (2000-2999)
	ErrCodeInvalidOperation  ErrorCode = 2000 // 无效操作 Invalid operation
	ErrCodeResourceNotFound  ErrorCode = 2001 // 资源未找到 Resource not found
	ErrCodeResourceExists    ErrorCode = 2002 // 资源已存在 Resource already exists
	ErrCodeInvalidParameter  ErrorCode = 2003 // 无效参数 Invalid parameter
	ErrCodeOperationConflict ErrorCode = 2004 // 操作冲突 Operation conflict

	// 网络错误码 (3000-3999)
	// Network error codes (3000-3999)
	ErrCodeConnectionFailed   ErrorCode = 3000 // 连接失败 Connection failed
	ErrCodeConnectionTimeout  ErrorCode = 3001 // 连接超时 Connection timeout
	ErrCodeNetworkError       ErrorCode = 3002 // 网络错误 Network error
	ErrCodeProtocolError      ErrorCode = 3003 // 协议错误 Protocol error
	ErrCodeServiceUnavailable ErrorCode = 3004 // 服务不可用 Service unavailable

	// 存储错误码 (4000-4999)
	// Storage error codes (4000-4999)
	ErrCodeStorageFailure      ErrorCode = 4000 // 存储故障 Storage failure
	ErrCodeDatabaseNotFound    ErrorCode = 4001 // 数据库未找到 Database not found
	ErrCodeTableNotFound       ErrorCode = 4002 // 表未找到 Table not found
	ErrCodeColumnNotFound      ErrorCode = 4003 // 列未找到 Column not found
	ErrCodeDuplicateKey        ErrorCode = 4004 // 重复键 Duplicate key
	ErrCodeConstraintViolation ErrorCode = 4005 // 约束违反 Constraint violation
	ErrCodeTransactionConflict ErrorCode = 4006 // 事务冲突 Transaction conflict
	ErrCodeLockTimeout         ErrorCode = 4007 // 锁超时 Lock timeout
	ErrCodeDeadlock            ErrorCode = 4008 // 死锁 Deadlock
	ErrCodeFileNotFound        ErrorCode = 4009

	// 认证错误码 (5000-5999)
	// Authentication error codes (5000-5999)
	ErrCodeAuthenticationFailed ErrorCode = 5000 // 认证失败 Authentication failed
	ErrCodeInvalidCredentials   ErrorCode = 5001 // 无效凭据 Invalid credentials
	ErrCodeUserNotFound         ErrorCode = 5002 // 用户未找到 User not found
	ErrCodeAccountLocked        ErrorCode = 5003 // 账户锁定 Account locked
	ErrCodePasswordExpired      ErrorCode = 5004 // 密码过期 Password expired
	ErrCodeTokenInvalid         ErrorCode = 5005 // 令牌无效 Token invalid
	ErrCodeTokenExpired         ErrorCode = 5006 // 令牌过期 Token expired

	// 权限错误码 (6000-6999)
	// Permission error codes (6000-6999)
	ErrCodeAccessDenied          ErrorCode = 6000 // 拒绝访问 Access denied
	ErrCodeInsufficientPrivilege ErrorCode = 6001 // 权限不足 Insufficient privilege
	ErrCodeOperationNotAllowed   ErrorCode = 6002 // 操作不被允许 Operation not allowed
	ErrCodeResourceForbidden     ErrorCode = 6003 // 资源被禁止 Resource forbidden

	// 验证错误码 (7000-7999)
	// Validation error codes (7000-7999)
	ErrCodeValidationFailed ErrorCode = 7000 // 验证失败 Validation failed
	ErrCodeInvalidFormat    ErrorCode = 7001 // 无效格式 Invalid format
	ErrCodeValueOutOfRange  ErrorCode = 7002 // 值超出范围 Value out of range
	ErrCodeRequiredField    ErrorCode = 7003 // 必填字段 Required field

	// 超时错误码 (8000-8999)
	// Timeout error codes (8000-8999)
	ErrCodeOperationTimeout ErrorCode = 8000 // 操作超时 Operation timeout
	ErrCodeQueryTimeout     ErrorCode = 8001 // 查询超时 Query timeout
	ErrCodeRequestTimeout   ErrorCode = 8002 // 请求超时 Request timeout

	// TODO: others
	ErrCodeNotSupported     ErrorCode = 9001
	ErrCodeNotFound         ErrorCode = 9002
	ErrCodeAlreadyExists    ErrorCode = 9003
	ErrCodeInvalidState     ErrorCode = 9004
	ErrCodePermissionDenied ErrorCode = 9005
	ErrCodeVersionMismatch  ErrorCode = 9006
	ErrCodeQuotaExceeded    ErrorCode = 9007
	ErrCodeDataCorruption   ErrorCode = 9008
)

// String 返回错误码的字符串表示
// String returns string representation of error code
func (c ErrorCode) String() string {
	switch c {
	// System errors
	case ErrCodeSystemFailure:
		return "SYSTEM_FAILURE"
	case ErrCodeOutOfMemory:
		return "OUT_OF_MEMORY"
	case ErrCodeInternalError:
		return "INTERNAL_ERROR"
	case ErrCodeConfigError:
		return "CONFIG_ERROR"
	case ErrCodeInitializationFail:
		return "INITIALIZATION_FAIL"
	// Business errors
	case ErrCodeInvalidOperation:
		return "INVALID_OPERATION"
	case ErrCodeResourceNotFound:
		return "RESOURCE_NOT_FOUND"
	case ErrCodeResourceExists:
		return "RESOURCE_EXISTS"
	case ErrCodeInvalidParameter:
		return "INVALID_PARAMETER"
	case ErrCodeOperationConflict:
		return "OPERATION_CONFLICT"
	// Network errors
	case ErrCodeConnectionFailed:
		return "CONNECTION_FAILED"
	case ErrCodeConnectionTimeout:
		return "CONNECTION_TIMEOUT"
	case ErrCodeNetworkError:
		return "NETWORK_ERROR"
	case ErrCodeProtocolError:
		return "PROTOCOL_ERROR"
	case ErrCodeServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	// Storage errors
	case ErrCodeStorageFailure:
		return "STORAGE_FAILURE"
	case ErrCodeDatabaseNotFound:
		return "DATABASE_NOT_FOUND"
	case ErrCodeTableNotFound:
		return "TABLE_NOT_FOUND"
	case ErrCodeColumnNotFound:
		return "COLUMN_NOT_FOUND"
	case ErrCodeDuplicateKey:
		return "DUPLICATE_KEY"
	case ErrCodeConstraintViolation:
		return "CONSTRAINT_VIOLATION"
	case ErrCodeTransactionConflict:
		return "TRANSACTION_CONFLICT"
	case ErrCodeLockTimeout:
		return "LOCK_TIMEOUT"
	case ErrCodeDeadlock:
		return "DEADLOCK"
	// Authentication errors
	case ErrCodeAuthenticationFailed:
		return "AUTHENTICATION_FAILED"
	case ErrCodeInvalidCredentials:
		return "INVALID_CREDENTIALS"
	case ErrCodeUserNotFound:
		return "USER_NOT_FOUND"
	case ErrCodeAccountLocked:
		return "ACCOUNT_LOCKED"
	case ErrCodePasswordExpired:
		return "PASSWORD_EXPIRED"
	case ErrCodeTokenInvalid:
		return "TOKEN_INVALID"
	case ErrCodeTokenExpired:
		return "TOKEN_EXPIRED"
	// Permission errors
	case ErrCodeAccessDenied:
		return "ACCESS_DENIED"
	case ErrCodeInsufficientPrivilege:
		return "INSUFFICIENT_PRIVILEGE"
	case ErrCodeOperationNotAllowed:
		return "OPERATION_NOT_ALLOWED"
	case ErrCodeResourceForbidden:
		return "RESOURCE_FORBIDDEN"
	// Validation errors
	case ErrCodeValidationFailed:
		return "VALIDATION_FAILED"
	case ErrCodeInvalidFormat:
		return "INVALID_FORMAT"
	case ErrCodeValueOutOfRange:
		return "VALUE_OUT_OF_RANGE"
	case ErrCodeRequiredField:
		return "REQUIRED_FIELD"
	// Timeout errors
	case ErrCodeOperationTimeout:
		return "OPERATION_TIMEOUT"
	case ErrCodeQueryTimeout:
		return "QUERY_TIMEOUT"
	case ErrCodeRequestTimeout:
		return "REQUEST_TIMEOUT"
	default:
		return fmt.Sprintf("UNKNOWN_ERROR_CODE(%d)", int(c))
	}
}

// Category 返回错误码对应的错误分类
// Category returns error category for the error code
func (c ErrorCode) Category() ErrorCategory {
	switch {
	case c >= 1000 && c < 2000:
		return ErrCategorySystem
	case c >= 2000 && c < 3000:
		return ErrCategoryBusiness
	case c >= 3000 && c < 4000:
		return ErrCategoryNetwork
	case c >= 4000 && c < 5000:
		return ErrCategoryStorage
	case c >= 5000 && c < 6000:
		return ErrCategoryAuth
	case c >= 6000 && c < 7000:
		return ErrCategoryPermission
	case c >= 7000 && c < 8000:
		return ErrCategoryValidation
	case c >= 8000 && c < 9000:
		return ErrCategoryTimeout
	default:
		return ErrCategorySystem
	}
}

// Severity 返回错误码对应的严重程度
// Severity returns error severity for the error code
func (c ErrorCode) Severity() ErrorSeverity {
	switch c.Category() {
	case ErrCategorySystem:
		return SeverityCritical
	case ErrCategoryStorage:
		if c == ErrCodeStorageFailure || c == ErrCodeDeadlock {
			return SeverityHigh
		}
		return SeverityMedium
	case ErrCategoryNetwork:
		if c == ErrCodeServiceUnavailable {
			return SeverityHigh
		}
		return SeverityMedium
	case ErrCategoryAuth, ErrCategoryPermission:
		return SeverityMedium
	case ErrCategoryTimeout:
		return SeverityMedium
	case ErrCategoryValidation, ErrCategoryBusiness:
		return SeverityLow
	default:
		return SeverityMedium
	}
}

// GuoceError GuoceDB错误接口
// GuoceError interface for GuoceDB errors
type GuoceError interface {
	error
	// Code 返回错误码
	// Code returns error code
	Code() ErrorCode
	// Category 返回错误分类
	// Category returns error category
	Category() ErrorCategory
	// Severity 返回错误严重程度
	// Severity returns error severity
	Severity() ErrorSeverity
	// Message 返回错误消息
	// Message returns error message
	Message() string
	// Details 返回错误详情
	// Details returns error details
	Details() map[string]interface{}
	// Cause 返回原因错误
	// Cause returns cause error
	Cause() error
	// Stack 返回堆栈信息
	// Stack returns stack trace
	Stack() string
	// WithDetail 添加错误详情
	// WithDetail adds error detail
	WithDetail(key string, value interface{}) GuoceError
	// WithCause 设置原因错误
	// WithCause sets cause error
	WithCause(err error) GuoceError
}

// baseError 基础错误实现
// baseError base error implementation
type baseError struct {
	code    ErrorCode              // 错误码 Error code
	message string                 // 错误消息 Error message
	details map[string]interface{} // 错误详情 Error details
	cause   error                  // 原因错误 Cause error
	stack   string                 // 堆栈信息 Stack trace
}

// NewError 创建新的GuoceError
// NewError creates new GuoceError
func NewError(code ErrorCode, message string) GuoceError {
	return &baseError{
		code:    code,
		message: message,
		details: make(map[string]interface{}),
		stack:   getStackTrace(),
	}
}

// NewErrorf 创建带格式化消息的GuoceError
// NewErrorf creates GuoceError with formatted message
func NewErrorf(code ErrorCode, format string, args ...interface{}) GuoceError {
	return NewError(code, fmt.Sprintf(format, args...))
}

// WrapError 包装现有错误为GuoceError
// WrapError wraps existing error as GuoceError
func WrapError(code ErrorCode, message string, cause error) GuoceError {
	return &baseError{
		code:    code,
		message: message,
		details: make(map[string]interface{}),
		cause:   cause,
		stack:   getStackTrace(),
	}
}

// Error 实现error接口
// Error implements error interface
func (e *baseError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.code.String(), e.message, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.code.String(), e.message)
}

// Code 返回错误码
// Code returns error code
func (e *baseError) Code() ErrorCode {
	return e.code
}

// Category 返回错误分类
// Category returns error category
func (e *baseError) Category() ErrorCategory {
	return e.code.Category()
}

// Severity 返回错误严重程度
// Severity returns error severity
func (e *baseError) Severity() ErrorSeverity {
	return e.code.Severity()
}

// Message 返回错误消息
// Message returns error message
func (e *baseError) Message() string {
	return e.message
}

// Details 返回错误详情
// Details returns error details
func (e *baseError) Details() map[string]interface{} {
	return e.details
}

// Cause 返回原因错误
// Cause returns cause error
func (e *baseError) Cause() error {
	return e.cause
}

// Stack 返回堆栈信息
// Stack returns stack trace
func (e *baseError) Stack() string {
	return e.stack
}

// WithDetail 添加错误详情
// WithDetail adds error detail
func (e *baseError) WithDetail(key string, value interface{}) GuoceError {
	e.details[key] = value
	return e
}

// WithCause 设置原因错误
// WithCause sets cause error
func (e *baseError) WithCause(err error) GuoceError {
	e.cause = err
	return e
}

// getStackTrace 获取堆栈跟踪信息
// getStackTrace gets stack trace information
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var str strings.Builder

	for i := 0; i < n; i++ {
		pc := pcs[i]
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line := fn.FileLine(pc)
		if i > 0 {
			str.WriteString("\n")
		}
		str.WriteString(fmt.Sprintf("\t%s:%d %s", file, line, fn.Name()))
	}

	return str.String()
}

// IsGuoceError 检查错误是否为GuoceError
// IsGuoceError checks if error is GuoceError
func IsGuoceError(err error) bool {
	_, ok := err.(GuoceError)
	return ok
}

// GetErrorCode 获取错误码，如果不是GuoceError则返回内部错误码
// GetErrorCode gets error code, returns internal error code if not GuoceError
func GetErrorCode(err error) ErrorCode {
	if gErr, ok := err.(GuoceError); ok {
		return gErr.Code()
	}
	return ErrCodeInternalError
}

// GetErrorCategory 获取错误分类
// GetErrorCategory gets error category
func GetErrorCategory(err error) ErrorCategory {
	if gErr, ok := err.(GuoceError); ok {
		return gErr.Category()
	}
	return ErrCategorySystem
}

// GetErrorSeverity 获取错误严重程度
// GetErrorSeverity gets error severity
func GetErrorSeverity(err error) ErrorSeverity {
	if gErr, ok := err.(GuoceError); ok {
		return gErr.Severity()
	}
	return SeverityMedium
}

// 预定义的常用错误
// Predefined common errors
var (
	// ErrSystemFailure 系统故障错误
	// ErrSystemFailure system failure error
	ErrSystemFailure = NewError(ErrCodeSystemFailure, "System failure occurred")

	// ErrInternalError 内部错误
	// ErrInternalError internal error
	ErrInternalError = NewError(ErrCodeInternalError, "Internal error occurred")

	// ErrResourceNotFound 资源未找到错误
	// ErrResourceNotFound resource not found error
	ErrResourceNotFound = NewError(ErrCodeResourceNotFound, "Resource not found")

	// ErrInvalidParameter 无效参数错误
	// ErrInvalidParameter invalid parameter error
	ErrInvalidParameter = NewError(ErrCodeInvalidParameter, "Invalid parameter")

	// ErrDatabaseNotFound 数据库未找到错误
	// ErrDatabaseNotFound database not found error
	ErrDatabaseNotFound = NewError(ErrCodeDatabaseNotFound, "Database not found")

	// ErrTableNotFound 表未找到错误
	// ErrTableNotFound table not found error
	ErrTableNotFound = NewError(ErrCodeTableNotFound, "Table not found")

	// ErrAuthenticationFailed 认证失败错误
	// ErrAuthenticationFailed authentication failed error
	ErrAuthenticationFailed = NewError(ErrCodeAuthenticationFailed, "Authentication failed")

	// ErrAccessDenied 拒绝访问错误
	// ErrAccessDenied access denied error
	ErrAccessDenied = NewError(ErrCodeAccessDenied, "Access denied")

	// ErrOperationTimeout 操作超时错误
	// ErrOperationTimeout operation timeout error
	ErrOperationTimeout = NewError(ErrCodeOperationTimeout, "Operation timeout")
)
