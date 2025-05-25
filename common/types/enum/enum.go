// Package enum defines global enumeration types used throughout the Guocedb project.
// These enums provide a set of named constants that represent predefined choices,
// enhancing type safety, readability, and maintainability.
//
// 此包定义了 Guocedb 项目中使用的全局枚举类型。
// 这些枚举提供了一组命名常量，代表预定义的选择，
// 从而增强了类型安全、可读性和可维护性。
package enum

import (
	"fmt"
	"strings"
)

// DBPresentationType represents the type of database presentation.
// This could be used for various purposes like how data is displayed or handled
// within different layers of the database system.
//
// DBPresentationType 表示数据库的呈现类型。
// 这可以用于数据库系统不同层中数据如何显示或处理的各种目的。
type DBPresentationType int

const (
	// DBPresentationType_Unspecified indicates an unspecified database presentation type.
	// 未指定数据库呈现类型。
	DBPresentationType_Unspecified DBPresentationType = iota
	// DBPresentationType_MySQL indicates MySQL compatible presentation.
	// MySQL 兼容的呈现类型。
	DBPresentationType_MySQL
	// DBPresentationType_JSON indicates JSON presentation.
	// JSON 呈现类型。
	DBPresentationType_JSON
	// DBPresentationType_GRPC indicates gRPC presentation (e.g., for API responses).
	// gRPC 呈现类型（例如，用于 API 响应）。
	DBPresentationType_GRPC
	// DBPresentationType_CSV indicates CSV presentation.
	// CSV 呈现类型。
	DBPresentationType_CSV
	// DBPresentationType_XML indicates XML presentation.
	// XML 呈现类型。
	DBPresentationType_XML
)

// String returns the string representation of a DBPresentationType.
// String 方法返回 DBPresentationType 的字符串表示。
func (t DBPresentationType) String() string {
	switch t {
	case DBPresentationType_Unspecified:
		return "UNSPECIFIED"
	case DBPresentationType_MySQL:
		return "MYSQL"
	case DBPresentationType_JSON:
		return "JSON"
	case DBPresentationType_GRPC:
		return "GRPC"
	case DBPentationType_CSV:
		return "CSV"
	case DBPresentationType_XML:
		return "XML"
	default:
		return fmt.Sprintf("UNKNOWN_DB_PRESENTATION_TYPE(%d)", t)
	}
}

// ParseDBPresentationType converts a string to a DBPresentationType.
// It is case-insensitive.
// ParseDBPresentationType 将字符串转换为 DBPresentationType。
// 它不区分大小写。
func ParseDBPresentationType(s string) (DBPresentationType, error) {
	switch strings.ToUpper(s) {
	case "UNSPECIFIED":
		return DBPresentationType_Unspecified, nil
	case "MYSQL":
		return DBPresentationType_MySQL, nil
	case "JSON":
		return DBPresentationType_JSON, nil
	case "GRPC":
		return DBPresentationType_GRPC, nil
	case "CSV":
		return DBPresentationType_CSV, nil
	case "XML":
		return DBPresentationType_XML, nil
	default:
		return DBPresentationType_Unspecified, fmt.Errorf("invalid DB presentation type: %s", s)
	}
}

// StorageEngineType represents the type of storage engine used by Guocedb.
// This enum allows easy identification and selection of different backend storage
// implementations, supporting the pluggable storage design.
//
// StorageEngineType 表示 Guocedb 使用的存储引擎类型。
// 此枚举方便识别和选择不同的后端存储实现，支持可插拔的存储设计。
type StorageEngineType int

const (
	// StorageEngineType_Unspecified indicates an unspecified storage engine type.
	// 未指定存储引擎类型。
	StorageEngineType_Unspecified StorageEngineType = iota
	// StorageEngineType_Badger indicates the Badger key-value store engine.
	// Badger 键值存储引擎。
	StorageEngineType_Badger
	// StorageEngineType_KVD indicates a generic Key-Value Database engine (placeholder).
	// 通用键值数据库引擎（占位符）。
	StorageEngineType_KVD
	// StorageEngineType_MDD indicates a Memory-Disk Database engine (placeholder).
	// 内存-磁盘数据库引擎（占位符）。
	StorageEngineType_MDD
	// StorageEngineType_MDI indicates a Memory-Disk Index engine (placeholder).
	// 内存-磁盘索引引擎（占位符）。
	StorageEngineType_MDI
)

// String returns the string representation of a StorageEngineType.
// String 方法返回 StorageEngineType 的字符串表示。
func (t StorageEngineType) String() string {
	switch t {
	case StorageEngineType_Unspecified:
		return "UNSPECIFIED"
	case StorageEngineType_Badger:
		return "BADGER"
	case StorageEngineType_KVD:
		return "KVD"
	case StorageEngineType_MDD:
		return "MDD"
	case StorageEngineType_MDI:
		return "MDI"
	default:
		return fmt.Sprintf("UNKNOWN_STORAGE_ENGINE_TYPE(%d)", t)
	}
}

// ParseStorageEngineType converts a string to a StorageEngineType.
// It is case-insensitive.
// ParseStorageEngineType 将字符串转换为 StorageEngineType。
// 它不区分大小写。
func ParseStorageEngineType(s string) (StorageEngineType, error) {
	switch strings.ToUpper(s) {
	case "UNSPECIFIED":
		return StorageEngineType_Unspecified, nil
	case "BADGER":
		return StorageEngineType_Badger, nil
	case "KVD":
		return StorageEngineType_KVD, nil
	case "MDD":
		return StorageEngineType_MDD, nil
	case "MDI":
		return StorageEngineType_MDI, nil
	default:
		return StorageEngineType_Unspecified, fmt.Errorf("invalid storage engine type: %s", s)
	}
}

// LogLevel represents the logging verbosity level.
// Used by the unified logging interface to control message output.
// LogLevel 表示日志的详细程度。
// 由统一日志接口使用，以控制消息输出。
type LogLevel int

const (
	// LogLevel_DEBUG is for verbose debugging messages.
	// 调试级别，用于详细的调试信息。
	LogLevel_DEBUG LogLevel = iota
	// LogLevel_INFO is for informational messages.
	// 信息级别，用于一般性信息。
	LogLevel_INFO
	// LogLevel_WARN is for warning messages.
	// 警告级别，用于潜在问题。
	LogLevel_WARN
	// LogLevel_ERROR is for error messages, indicating a problem.
	// 错误级别，表示一个问题。
	LogLevel_ERROR
	// LogLevel_FATAL is for critical errors that cause program termination.
	// 致命级别，表示导致程序终止的严重错误。
	LogLevel_FATAL
)

// String returns the string representation of a LogLevel.
// String 方法返回 LogLevel 的字符串表示。
func (l LogLevel) String() string {
	switch l {
	case LogLevel_DEBUG:
		return "DEBUG"
	case LogLevel_INFO:
		return "INFO"
	case LogLevel_WARN:
		return "WARN"
	case LogLevel_ERROR:
		return "ERROR"
	case LogLevel_FATAL:
		return "FATAL"
	default:
		return fmt.Sprintf("UNKNOWN_LOG_LEVEL(%d)", l)
	}
}

// ParseLogLevel converts a string to a LogLevel.
// It is case-insensitive.
// ParseLogLevel 将字符串转换为 LogLevel。
// 它不区分大小写。
func ParseLogLevel(s string) (LogLevel, error) {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LogLevel_DEBUG, nil
	case "INFO":
		return LogLevel_INFO, nil
	case "WARN":
		return LogLevel_WARN, nil
	case "ERROR":
		return LogLevel_ERROR, nil
	case "FATAL":
		return LogLevel_FATAL, nil
	default:
		return LogLevel_DEBUG, fmt.Errorf("invalid log level: %s", s)
	}
}

// AuthMethodType represents the authentication method used for client connections.
// 认证方式类型，用于客户端连接。
type AuthMethodType int

const (
	// AuthMethod_NativePassword indicates MySQL's native password authentication.
	// MySQL 原生密码认证。
	AuthMethod_NativePassword AuthMethodType = iota
	// AuthMethod_CleartextPassword indicates cleartext password authentication (less secure).
	// 明文密码认证（安全性较低）。
	AuthMethod_CleartextPassword
	// AuthMethod_CachingSha2Password indicates caching_sha2_password authentication (MySQL 8.0 default).
	// caching_sha2_password 认证（MySQL 8.0 默认）。
	AuthMethod_CachingSha2Password
	// AuthMethod_TLSClientCert indicates authentication via TLS client certificates.
	// 通过 TLS 客户端证书进行认证。
	AuthMethod_TLSClientCert
)

// String returns the string representation of an AuthMethodType.
// String 方法返回 AuthMethodType 的字符串表示。
func (a AuthMethodType) String() string {
	switch a {
	case AuthMethod_NativePassword:
		return "NativePassword"
	case AuthMethod_CleartextPassword:
		return "CleartextPassword"
	case AuthMethod_CachingSha2Password:
		return "CachingSha2Password"
	case AuthMethod_TLSClientCert:
		return "TLSClientCert"
	default:
		return fmt.Sprintf("UNKNOWN_AUTH_METHOD_TYPE(%d)", a)
	}
}

// ParseAuthMethodType converts a string to an AuthMethodType.
// It is case-insensitive.
// ParseAuthMethodType 将字符串转换为 AuthMethodType。
// 它不区分大小写。
func ParseAuthMethodType(s string) (AuthMethodType, error) {
	switch strings.ToLower(s) {
	case "nativepassword":
		return AuthMethod_NativePassword, nil
	case "cleartextpassword":
		return AuthMethod_CleartextPassword, nil
	case "cachingsha2password":
		return AuthMethod_CachingSha2Password, nil
	case "tlsclientcert":
		return AuthMethod_TLSClientCert, nil
	default:
		return AuthMethod_NativePassword, fmt.Errorf("invalid authentication method type: %s", s)
	}
}

// AuditEventType represents the type of an audit log event.
// AuditEventType 表示审计日志事件的类型。
type AuditEventType int

const (
	// AuditEvent_Unspecified indicates an unspecified audit event type.
	// 未指定审计事件类型。
	AuditEvent_Unspecified AuditEventType = iota
	// AuditEvent_Connection indicates a client connection event (e.g., connect, disconnect).
	// 连接事件（例如，连接、断开连接）。
	AuditEvent_Connection
	// AuditEvent_Authentication indicates an authentication attempt (success/failure).
	// 认证尝试事件（成功/失败）。
	AuditEvent_Authentication
	// AuditEvent_Query indicates a SQL query execution event.
	// SQL 查询执行事件。
	AuditEvent_Query
	// AuditEvent_DDL indicates a Data Definition Language (DDL) operation.
	// 数据定义语言（DDL）操作。
	AuditEvent_DDL
	// AuditEvent_DML indicates a Data Manipulation Language (DML) operation.
	// 数据操作语言（DML）操作。
	AuditEvent_DML
	// AuditEvent_Admin indicates an administrative operation (e.g., user management, configuration change).
	// 管理操作（例如，用户管理、配置更改）。
	AuditEvent_Admin
)

// String returns the string representation of an AuditEventType.
// String 方法返回 AuditEventType 的字符串表示。
func (e AuditEventType) String() string {
	switch e {
	case AuditEvent_Unspecified:
		return "UNSPECIFIED"
	case AuditEvent_Connection:
		return "CONNECTION"
	case AuditEvent_Authentication:
		return "AUTHENTICATION"
	case AuditEvent_Query:
		return "QUERY"
	case AuditEvent_DDL:
		return "DDL"
	case AuditEvent_DML:
		return "DML"
	case AuditEvent_Admin:
		return "ADMIN"
	default:
		return fmt.Sprintf("UNKNOWN_AUDIT_EVENT_TYPE(%d)", e)
	}
}

// ParseAuditEventType converts a string to an AuditEventType.
// It is case-insensitive.
// ParseAuditEventType 将字符串转换为 AuditEventType。
// 它不区分大小写。
func ParseAuditEventType(s string) (AuditEventType, error) {
	switch strings.ToUpper(s) {
	case "UNSPECIFIED":
		return AuditEvent_Unspecified, nil
	case "CONNECTION":
		return AuditEvent_Connection, nil
	case "AUTHENTICATION":
		return AuditEvent_Authentication, nil
	case "QUERY":
		return AuditEvent_Query, nil
	case "DDL":
		return AuditEvent_DDL, nil
	case "DML":
		return AuditEvent_DML, nil
	case "ADMIN":
		return AuditEvent_Admin, nil
	default:
		return AuditEvent_Unspecified, fmt.Errorf("invalid audit event type: %s", s)
	}
}
