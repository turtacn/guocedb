// Package errors 提供了 GuoceDB 的统一错误处理机制
// Package errors provides unified error handling mechanism for GuoceDB
package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ===== 错误级别 Error Levels =====

// Level 错误级别
// Level error level
type Level int

const (
	// LevelWarning 警告级别
	// LevelWarning warning level
	LevelWarning Level = iota
	// LevelError 错误级别
	// LevelError error level
	LevelError
	// LevelFatal 致命错误级别
	// LevelFatal fatal error level
	LevelFatal
)

// String 返回错误级别的字符串表示
// String returns string representation of error level
func (l Level) String() string {
	switch l {
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ===== 错误分类 Error Categories =====

// Category 错误分类
// Category error category
type Category string

const (
	// CategorySyntax SQL语法错误
	// CategorySyntax SQL syntax error
	CategorySyntax Category = "SYNTAX"

	// CategoryRuntime 运行时错误
	// CategoryRuntime runtime error
	CategoryRuntime Category = "RUNTIME"

	// CategorySystem 系统错误
	// CategorySystem system error
	CategorySystem Category = "SYSTEM"

	// CategoryStorage 存储错误
	// CategoryStorage storage error
	CategoryStorage Category = "STORAGE"

	// CategoryNetwork 网络错误
	// CategoryNetwork network error
	CategoryNetwork Category = "NETWORK"

	// CategoryAuth 认证授权错误
	// CategoryAuth authentication/authorization error
	CategoryAuth Category = "AUTH"

	// CategoryTransaction 事务错误
	// CategoryTransaction transaction error
	CategoryTransaction Category = "TRANSACTION"

	// CategoryConstraint 约束违反错误
	// CategoryConstraint constraint violation error
	CategoryConstraint Category = "CONSTRAINT"

	// CategoryData 数据错误
	// CategoryData data error
	CategoryData Category = "DATA"

	// CategoryConfig 配置错误
	// CategoryConfig configuration error
	CategoryConfig Category = "CONFIG"
)

// ===== 错误码定义 Error Code Definitions =====

// 错误码范围分配：
// 1000-1999: SQL语法错误
// 2000-2999: 运行时错误
// 3000-3999: 系统错误
// 4000-4999: 存储错误
// 5000-5999: 网络错误
// 6000-6999: 认证授权错误
// 7000-7999: 事务错误
// 8000-8999: 约束错误
// 9000-9999: 数据错误

const (
	// ===== SQL语法错误 SQL Syntax Errors (1000-1999) =====

	// ErrSyntaxError 通用语法错误
	// ErrSyntaxError general syntax error
	ErrSyntaxError = 1000

	// ErrUnknownDatabase 未知数据库
	// ErrUnknownDatabase unknown database
	ErrUnknownDatabase = 1001

	// ErrUnknownTable 未知表
	// ErrUnknownTable unknown table
	ErrUnknownTable = 1002

	// ErrUnknownColumn 未知列
	// ErrUnknownColumn unknown column
	ErrUnknownColumn = 1003

	// ErrDuplicateDatabase 数据库已存在
	// ErrDuplicateDatabase database already exists
	ErrDuplicateDatabase = 1004

	// ErrDuplicateTable 表已存在
	// ErrDuplicateTable table already exists
	ErrDuplicateTable = 1005

	// ErrDuplicateColumn 列已存在
	// ErrDuplicateColumn column already exists
	ErrDuplicateColumn = 1006

	// ErrInvalidIdentifier 无效标识符
	// ErrInvalidIdentifier invalid identifier
	ErrInvalidIdentifier = 1007

	// ErrInvalidExpression 无效表达式
	// ErrInvalidExpression invalid expression
	ErrInvalidExpression = 1008

	// ErrInvalidFunction 无效函数
	// ErrInvalidFunction invalid function
	ErrInvalidFunction = 1009

	// ErrWrongNumberOfArguments 参数数量错误
	// ErrWrongNumberOfArguments wrong number of arguments
	ErrWrongNumberOfArguments = 1010

	// ===== 运行时错误 Runtime Errors (2000-2999) =====

	// ErrOutOfMemory 内存不足
	// ErrOutOfMemory out of memory
	ErrOutOfMemory = 2000

	// ErrTimeout 操作超时
	// ErrTimeout operation timeout
	ErrTimeout = 2001

	// ErrCancelled 操作被取消
	// ErrCancelled operation cancelled
	ErrCancelled = 2002

	// ErrTooManyConnections 连接数过多
	// ErrTooManyConnections too many connections
	ErrTooManyConnections = 2003

	// ErrConnectionLost 连接丢失
	// ErrConnectionLost connection lost
	ErrConnectionLost = 2004

	// ErrQueryTooLarge 查询过大
	// ErrQueryTooLarge query too large
	ErrQueryTooLarge = 2005

	// ErrResultSetTooLarge 结果集过大
	// ErrResultSetTooLarge result set too large
	ErrResultSetTooLarge = 2006

	// ErrDivisionByZero 除零错误
	// ErrDivisionByZero division by zero
	ErrDivisionByZero = 2007

	// ErrOverflow 数值溢出
	// ErrOverflow numeric overflow
	ErrOverflow = 2008

	// ErrInvalidOperation 无效操作
	// ErrInvalidOperation invalid operation
	ErrInvalidOperation = 2009

	// ===== 系统错误 System Errors (3000-3999) =====

	// ErrInternal 内部错误
	// ErrInternal internal error
	ErrInternal = 3000

	// ErrNotImplemented 功能未实现
	// ErrNotImplemented not implemented
	ErrNotImplemented = 3001

	// ErrUnsupported 不支持的功能
	// ErrUnsupported unsupported feature
	ErrUnsupported = 3002

	// ErrSystemShutdown 系统关闭中
	// ErrSystemShutdown system shutting down
	ErrSystemShutdown = 3003

	// ErrInitializationFailed 初始化失败
	// ErrInitializationFailed initialization failed
	ErrInitializationFailed = 3004

	// ErrConfigInvalid 配置无效
	// ErrConfigInvalid invalid configuration
	ErrConfigInvalid = 3005

	// ===== 存储错误 Storage Errors (4000-4999) =====

	// ErrStorageError 通用存储错误
	// ErrStorageError general storage error
	ErrStorageError = 4000

	// ErrDiskFull 磁盘空间不足
	// ErrDiskFull disk full
	ErrDiskFull = 4001

	// ErrCorruptedData 数据损坏
	// ErrCorruptedData corrupted data
	ErrCorruptedData = 4002

	// ErrIOError IO错误
	// ErrIOError IO error
	ErrIOError = 4003

	// ErrFileNotFound 文件未找到
	// ErrFileNotFound file not found
	ErrFileNotFound = 4004

	// ErrFileExists 文件已存在
	// ErrFileExists file already exists
	ErrFileExists = 4005

	// ErrStorageEngineError 存储引擎错误
	// ErrStorageEngineError storage engine error
	ErrStorageEngineError = 4006

	// ===== 网络错误 Network Errors (5000-5999) =====

	// ErrNetworkError 通用网络错误
	// ErrNetworkError general network error
	ErrNetworkError = 5000

	// ErrConnectionRefused 连接被拒绝
	// ErrConnectionRefused connection refused
	ErrConnectionRefused = 5001

	// ErrHostUnreachable 主机不可达
	// ErrHostUnreachable host unreachable
	ErrHostUnreachable = 5002

	// ErrProtocolError 协议错误
	// ErrProtocolError protocol error
	ErrProtocolError = 5003

	// ===== 认证授权错误 Auth Errors (6000-6999) =====

	// ErrAccessDenied 访问被拒绝
	// ErrAccessDenied access denied
	ErrAccessDenied = 6000

	// ErrInvalidCredentials 无效凭证
	// ErrInvalidCredentials invalid credentials
	ErrInvalidCredentials = 6001

	// ErrUserNotFound 用户不存在
	// ErrUserNotFound user not found
	ErrUserNotFound = 6002

	// ErrInsufficientPrivileges 权限不足
	// ErrInsufficientPrivileges insufficient privileges
	ErrInsufficientPrivileges = 6003

	// ErrPasswordExpired 密码已过期
	// ErrPasswordExpired password expired
	ErrPasswordExpired = 6004

	// ErrAccountLocked 账户已锁定
	// ErrAccountLocked account locked
	ErrAccountLocked = 6005

	// ===== 事务错误 Transaction Errors (7000-7999) =====

	// ErrTransactionError 通用事务错误
	// ErrTransactionError general transaction error
	ErrTransactionError = 7000

	// ErrDeadlock 死锁
	// ErrDeadlock deadlock
	ErrDeadlock = 7001

	// ErrLockTimeout 锁超时
	// ErrLockTimeout lock timeout
	ErrLockTimeout = 7002

	// ErrTransactionRollback 事务回滚
	// ErrTransactionRollback transaction rollback
	ErrTransactionRollback = 7003

	// ErrTransactionTooLarge 事务过大
	// ErrTransactionTooLarge transaction too large
	ErrTransactionTooLarge = 7004

	// ErrIsolationLevel 隔离级别错误
	// ErrIsolationLevel isolation level error
	ErrIsolationLevel = 7005

	// ErrConcurrentUpdate 并发更新冲突
	// ErrConcurrentUpdate concurrent update conflict
	ErrConcurrentUpdate = 7006

	// ===== 约束错误 Constraint Errors (8000-8999) =====

	// ErrConstraintViolation 通用约束违反
	// ErrConstraintViolation general constraint violation
	ErrConstraintViolation = 8000

	// ErrDuplicateKey 主键/唯一键重复
	// ErrDuplicateKey duplicate key
	ErrDuplicateKey = 8001

	// ErrForeignKeyViolation 外键约束违反
	// ErrForeignKeyViolation foreign key violation
	ErrForeignKeyViolation = 8002

	// ErrNotNullViolation 非空约束违反
	// ErrNotNullViolation not null violation
	ErrNotNullViolation = 8003

	// ErrCheckViolation 检查约束违反
	// ErrCheckViolation check constraint violation
	ErrCheckViolation = 8004

	// ===== 数据错误 Data Errors (9000-9999) =====

	// ErrDataError 通用数据错误
	// ErrDataError general data error
	ErrDataError = 9000

	// ErrInvalidDataType 无效数据类型
	// ErrInvalidDataType invalid data type
	ErrInvalidDataType = 9001

	// ErrDataTruncated 数据被截断
	// ErrDataTruncated data truncated
	ErrDataTruncated = 9002

	// ErrInvalidDatetime 无效日期时间
	// ErrInvalidDatetime invalid datetime
	ErrInvalidDatetime = 9003

	// ErrInvalidCharset 无效字符集
	// ErrInvalidCharset invalid charset
	ErrInvalidCharset = 9004

	// ErrDataOutOfRange 数据超出范围
	// ErrDataOutOfRange data out of range
	ErrDataOutOfRange = 9005
)

// ===== 错误消息模板 Error Message Templates =====

var errorMessages = map[int]string{
	// SQL语法错误
	ErrSyntaxError:            "You have an error in your SQL syntax",
	ErrUnknownDatabase:        "Unknown database '%s'",
	ErrUnknownTable:           "Unknown table '%s'",
	ErrUnknownColumn:          "Unknown column '%s' in '%s'",
	ErrDuplicateDatabase:      "Database '%s' already exists",
	ErrDuplicateTable:         "Table '%s' already exists",
	ErrDuplicateColumn:        "Duplicate column name '%s'",
	ErrInvalidIdentifier:      "Invalid identifier '%s'",
	ErrInvalidExpression:      "Invalid expression: %s",
	ErrInvalidFunction:        "Unknown function '%s'",
	ErrWrongNumberOfArguments: "Incorrect number of arguments for function %s",

	// 运行时错误
	ErrOutOfMemory:        "Out of memory",
	ErrTimeout:            "Operation timed out",
	ErrCancelled:          "Operation was cancelled",
	ErrTooManyConnections: "Too many connections",
	ErrConnectionLost:     "Lost connection to server",
	ErrQueryTooLarge:      "Query too large",
	ErrResultSetTooLarge:  "Result set too large",
	ErrDivisionByZero:     "Division by zero",
	ErrOverflow:           "Numeric value overflow",
	ErrInvalidOperation:   "Invalid operation: %s",

	// 系统错误
	ErrInternal:             "Internal error: %s",
	ErrNotImplemented:       "Feature not implemented: %s",
	ErrUnsupported:          "Unsupported feature: %s",
	ErrSystemShutdown:       "Server is shutting down",
	ErrInitializationFailed: "Initialization failed: %s",
	ErrConfigInvalid:        "Invalid configuration: %s",

	// 存储错误
	ErrStorageError:       "Storage error: %s",
	ErrDiskFull:           "Disk full",
	ErrCorruptedData:      "Data corruption detected",
	ErrIOError:            "IO error: %s",
	ErrFileNotFound:       "File not found: %s",
	ErrFileExists:         "File already exists: %s",
	ErrStorageEngineError: "Storage engine error: %s",

	// 网络错误
	ErrNetworkError:      "Network error: %s",
	ErrConnectionRefused: "Connection refused",
	ErrHostUnreachable:   "Host unreachable: %s",
	ErrProtocolError:     "Protocol error: %s",

	// 认证授权错误
	ErrAccessDenied:           "Access denied for user '%s'",
	ErrInvalidCredentials:     "Invalid username or password",
	ErrUserNotFound:           "User '%s' not found",
	ErrInsufficientPrivileges: "Insufficient privileges for operation",
	ErrPasswordExpired:        "Password has expired",
	ErrAccountLocked:          "Account is locked",

	// 事务错误
	ErrTransactionError:    "Transaction error: %s",
	ErrDeadlock:            "Deadlock found when trying to get lock",
	ErrLockTimeout:         "Lock wait timeout exceeded",
	ErrTransactionRollback: "Transaction has been rolled back",
	ErrTransactionTooLarge: "Transaction too large",
	ErrIsolationLevel:      "Invalid transaction isolation level",
	ErrConcurrentUpdate:    "Concurrent update conflict",

	// 约束错误
	ErrConstraintViolation: "Constraint violation: %s",
	ErrDuplicateKey:        "Duplicate entry '%s' for key '%s'",
	ErrForeignKeyViolation: "Foreign key constraint fails",
	ErrNotNullViolation:    "Column '%s' cannot be null",
	ErrCheckViolation:      "Check constraint '%s' is violated",

	// 数据错误
	ErrDataError:       "Data error: %s",
	ErrInvalidDataType: "Invalid data type for column '%s'",
	ErrDataTruncated:   "Data truncated for column '%s'",
	ErrInvalidDatetime: "Invalid datetime value: '%s'",
	ErrInvalidCharset:  "Invalid character set: '%s'",
	ErrDataOutOfRange:  "Value out of range for column '%s'",
}

// ===== 错误结构体 Error Structure =====

// Error GuoceDB错误结构体
// Error GuoceDB error structure
type Error struct {
	Code     int                    // 错误码 Error code
	Level    Level                  // 错误级别 Error level
	Category Category               // 错误分类 Error category
	Message  string                 // 错误消息 Error message
	Details  string                 // 详细信息 Detailed information
	Context  map[string]interface{} // 上下文信息 Context information
	Cause    error                  // 原因错误 Cause error
	Stack    []StackFrame           // 调用栈 Call stack
}

// StackFrame 调用栈帧
// StackFrame call stack frame
type StackFrame struct {
	File     string // 文件名 File name
	Line     int    // 行号 Line number
	Function string // 函数名 Function name
}

// Error 实现 error 接口
// Error implements error interface
func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回包装的错误
// Unwrap returns wrapped error
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithContext 添加上下文信息
// WithContext adds context information
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithDetails 添加详细信息
// WithDetails adds detailed information
func (e *Error) WithDetails(details string) *Error {
	e.Details = details
	return e
}

// WithCause 设置原因错误
// WithCause sets cause error
func (e *Error) WithCause(cause error) *Error {
	e.Cause = cause
	return e
}

// Is 判断是否为特定错误码
// Is checks if error has specific code
func (e *Error) Is(code int) bool {
	return e.Code == code
}

// HasCategory 判断是否属于特定分类
// HasCategory checks if error belongs to specific category
func (e *Error) HasCategory(category Category) bool {
	return e.Category == category
}

// ===== 错误创建函数 Error Creation Functions =====

// New 创建新错误
// New creates new error
func New(code int, args ...interface{}) *Error {
	message, ok := errorMessages[code]
	if !ok {
		message = "Unknown error"
	}

	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	e := &Error{
		Code:     code,
		Level:    getDefaultLevel(code),
		Category: getDefaultCategory(code),
		Message:  message,
		Context:  make(map[string]interface{}),
	}

	// 捕获调用栈
	e.captureStack()

	return e
}

// Newf 创建格式化错误
// Newf creates formatted error
func Newf(code int, format string, args ...interface{}) *Error {
	e := &Error{
		Code:     code,
		Level:    getDefaultLevel(code),
		Category: getDefaultCategory(code),
		Message:  fmt.Sprintf(format, args...),
		Context:  make(map[string]interface{}),
	}

	e.captureStack()
	return e
}

// Wrap 包装已有错误
// Wrap wraps existing error
func Wrap(err error, code int, args ...interface{}) *Error {
	if err == nil {
		return nil
	}

	e := New(code, args...)
	e.Cause = err

	// 如果原错误也是 Error 类型，继承一些信息
	if guoceErr, ok := err.(*Error); ok {
		// 继承上下文
		for k, v := range guoceErr.Context {
			e.Context[k] = v
		}
	}

	return e
}

// ===== 辅助函数 Helper Functions =====

// getDefaultLevel 获取默认错误级别
// getDefaultLevel gets default error level
func getDefaultLevel(code int) Level {
	switch {
	case code >= 3000 && code < 4000: // 系统错误
		return LevelFatal
	case code >= 7000 && code < 8000: // 事务错误
		return LevelError
	default:
		return LevelError
	}
}

// getDefaultCategory 获取默认错误分类
// getDefaultCategory gets default error category
func getDefaultCategory(code int) Category {
	switch {
	case code >= 1000 && code < 2000:
		return CategorySyntax
	case code >= 2000 && code < 3000:
		return CategoryRuntime
	case code >= 3000 && code < 4000:
		return CategorySystem
	case code >= 4000 && code < 5000:
		return CategoryStorage
	case code >= 5000 && code < 6000:
		return CategoryNetwork
	case code >= 6000 && code < 7000:
		return CategoryAuth
	case code >= 7000 && code < 8000:
		return CategoryTransaction
	case code >= 8000 && code < 9000:
		return CategoryConstraint
	case code >= 9000 && code < 10000:
		return CategoryData
	default:
		return CategorySystem
	}
}

// captureStack 捕获调用栈
// captureStack captures call stack
func (e *Error) captureStack() {
	const maxStackDepth = 32
	pcs := make([]uintptr, maxStackDepth)
	n := runtime.Callers(3, pcs) // 跳过前3层

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()

		// 跳过runtime和错误包内部的帧
		if strings.Contains(frame.File, "runtime/") ||
			strings.Contains(frame.File, "common/errors") {
			if !more {
				break
			}
			continue
		}

		e.Stack = append(e.Stack, StackFrame{
			File:     frame.File,
			Line:     frame.Line,
			Function: frame.Function,
		})

		if !more {
			break
		}
	}
}

// ===== 错误判断函数 Error Check Functions =====

// IsNotFound 判断是否为未找到错误
// IsNotFound checks if error is not found error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*Error); ok {
		return e.Code == ErrUnknownDatabase ||
			e.Code == ErrUnknownTable ||
			e.Code == ErrUnknownColumn ||
			e.Code == ErrUserNotFound ||
			e.Code == ErrFileNotFound
	}
	return false
}

// IsDuplicate 判断是否为重复错误
// IsDuplicate checks if error is duplicate error
func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*Error); ok {
		return e.Code == ErrDuplicateDatabase ||
			e.Code == ErrDuplicateTable ||
			e.Code == ErrDuplicateColumn ||
			e.Code == ErrDuplicateKey ||
			e.Code == ErrFileExists
	}
	return false
}

// IsTimeout 判断是否为超时错误
// IsTimeout checks if error is timeout error
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*Error); ok {
		return e.Code == ErrTimeout || e.Code == ErrLockTimeout
	}
	return false
}

// IsConstraintViolation 判断是否为约束违反错误
// IsConstraintViolation checks if error is constraint violation
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*Error); ok {
		return e.Category == CategoryConstraint
	}
	return false
}

// IsRetryable 判断错误是否可重试
// IsRetryable checks if error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*Error); ok {
		return e.Code == ErrTimeout ||
			e.Code == ErrDeadlock ||
			e.Code == ErrLockTimeout ||
			e.Code == ErrConcurrentUpdate ||
			e.Code == ErrConnectionLost
	}
	return false
}

// ===== 错误格式化 Error Formatting =====

// FormatError 格式化错误信息
// FormatError formats error information
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	e, ok := err.(*Error)
	if !ok {
		return err.Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] Error %d: %s\n", e.Level, e.Code, e.Message))

	if e.Details != "" {
		sb.WriteString(fmt.Sprintf("Details: %s\n", e.Details))
	}

	if len(e.Context) > 0 {
		sb.WriteString("Context:\n")
		for k, v := range e.Context {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}

	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf("Caused by: %v\n", e.Cause))
	}

	if len(e.Stack) > 0 {
		sb.WriteString("Stack trace:\n")
		for i, frame := range e.Stack {
			sb.WriteString(fmt.Sprintf("  %d. %s\n     at %s:%d\n",
				i+1, frame.Function, frame.File, frame.Line))
		}
	}

	return sb.String()
}

// ===== 批量错误处理 Batch Error Handling =====

// ErrorList 错误列表
// ErrorList error list
type ErrorList struct {
	Errors []*Error
}

// Add 添加错误
// Add adds error
func (el *ErrorList) Add(err error) {
	if err == nil {
		return
	}

	if e, ok := err.(*Error); ok {
		el.Errors = append(el.Errors, e)
	} else {
		// 将普通错误包装为 Error
		el.Errors = append(el.Errors, Wrap(err, ErrInternal, err.Error()))
	}
}

// HasErrors 判断是否有错误
// HasErrors checks if there are errors
func (el *ErrorList) HasErrors() bool {
	return len(el.Errors) > 0
}

// Error 实现 error 接口
// Error implements error interface
func (el *ErrorList) Error() string {
	if len(el.Errors) == 0 {
		return ""
	}

	var messages []string
	for _, err := range el.Errors {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// First 返回第一个错误
// First returns first error
func (el *ErrorList) First() *Error {
	if len(el.Errors) > 0 {
		return el.Errors[0]
	}
	return nil
}

// ===== 全局错误变量 Global Error Variables =====

var (
	// ErrNil 空错误，用于占位
	// ErrNil nil error for placeholder
	ErrNil = (*Error)(nil)
)
