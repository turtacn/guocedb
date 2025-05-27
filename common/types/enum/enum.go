// Package enum 定义GuoceDB项目使用的全局枚举类型
// Package enum defines global enumeration types used by GuoceDB project
package enum

import "fmt"

// StorageEngineType 存储引擎类型枚举
// StorageEngineType enumeration for storage engine types
type StorageEngineType int

const (
	// BadgerEngine Badger KV存储引擎
	// BadgerEngine Badger KV storage engine
	BadgerEngine StorageEngineType = iota
	// KVDEngine KVD存储引擎
	// KVDEngine KVD storage engine
	KVDEngine
	// MDDEngine MDD存储引擎
	// MDDEngine MDD storage engine
	MDDEngine
	// MDIEngine MDI存储引擎
	// MDIEngine MDI storage engine
	MDIEngine
)

// String 返回存储引擎类型的字符串表示
// String returns string representation of storage engine type
func (s StorageEngineType) String() string {
	switch s {
	case BadgerEngine:
		return "BadgerEngine"
	case KVDEngine:
		return "KVDEngine"
	case MDDEngine:
		return "MDDEngine"
	case MDIEngine:
		return "MDIEngine"
	default:
		return fmt.Sprintf("Unknown StorageEngineType(%d)", int(s))
	}
}

// IsValid 检查存储引擎类型是否有效
// IsValid checks if storage engine type is valid
func (s StorageEngineType) IsValid() bool {
	return s >= BadgerEngine && s <= MDIEngine
}

// TransactionStatus 事务状态枚举
// TransactionStatus enumeration for transaction states
type TransactionStatus int

const (
	// TxActive 活跃事务状态
	// TxActive active transaction state
	TxActive TransactionStatus = iota
	// TxPreparing 准备提交状态
	// TxPreparing preparing to commit state
	TxPreparing
	// TxCommitted 已提交状态
	// TxCommitted committed state
	TxCommitted
	// TxAborted 已中止状态
	// TxAborted aborted state
	TxAborted
)

// String 返回事务状态的字符串表示
// String returns string representation of transaction status
func (t TransactionStatus) String() string {
	switch t {
	case TxActive:
		return "Active"
	case TxPreparing:
		return "Preparing"
	case TxCommitted:
		return "Committed"
	case TxAborted:
		return "Aborted"
	default:
		return fmt.Sprintf("Unknown TransactionStatus(%d)", int(t))
	}
}

// IsValid 检查事务状态是否有效
// IsValid checks if transaction status is valid
func (t TransactionStatus) IsValid() bool {
	return t >= TxActive && t <= TxAborted
}

// IsFinal 检查事务状态是否为最终状态
// IsFinal checks if transaction status is final
func (t TransactionStatus) IsFinal() bool {
	return t == TxCommitted || t == TxAborted
}

// LogLevel 日志级别枚举
// LogLevel enumeration for log levels
type LogLevel int

const (
	// LogDebug 调试级别
	// LogDebug debug level
	LogDebug LogLevel = iota
	// LogInfo 信息级别
	// LogInfo info level
	LogInfo
	// LogWarn 警告级别
	// LogWarn warning level
	LogWarn
	// LogError 错误级别
	// LogError error level
	LogError
	// LogFatal 致命错误级别
	// LogFatal fatal error level
	LogFatal
)

// String 返回日志级别的字符串表示
// String returns string representation of log level
func (l LogLevel) String() string {
	switch l {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	case LogWarn:
		return "WARN"
	case LogError:
		return "ERROR"
	case LogFatal:
		return "FATAL"
	default:
		return fmt.Sprintf("Unknown LogLevel(%d)", int(l))
	}
}

// IsValid 检查日志级别是否有效
// IsValid checks if log level is valid
func (l LogLevel) IsValid() bool {
	return l >= LogDebug && l <= LogFatal
}

// Priority 返回日志级别的优先级（数值越大优先级越高）
// Priority returns the priority of log level (higher number means higher priority)
func (l LogLevel) Priority() int {
	return int(l)
}

// AuthMethod 认证方式枚举
// AuthMethod enumeration for authentication methods
type AuthMethod int

const (
	// AuthNative 本地认证
	// AuthNative native authentication
	AuthNative AuthMethod = iota
	// AuthLDAP LDAP认证
	// AuthLDAP LDAP authentication
	AuthLDAP
	// AuthOAuth OAuth认证
	// AuthOAuth OAuth authentication
	AuthOAuth
)

// String 返回认证方式的字符串表示
// String returns string representation of authentication method
func (a AuthMethod) String() string {
	switch a {
	case AuthNative:
		return "Native"
	case AuthLDAP:
		return "LDAP"
	case AuthOAuth:
		return "OAuth"
	default:
		return fmt.Sprintf("Unknown AuthMethod(%d)", int(a))
	}
}

// IsValid 检查认证方式是否有效
// IsValid checks if authentication method is valid
func (a AuthMethod) IsValid() bool {
	return a >= AuthNative && a <= AuthOAuth
}

// PermissionType 权限类型枚举
// PermissionType enumeration for permission types
type PermissionType int

const (
	// PermSelect SELECT权限
	// PermSelect SELECT permission
	PermSelect PermissionType = iota
	// PermInsert INSERT权限
	// PermInsert INSERT permission
	PermInsert
	// PermUpdate UPDATE权限
	// PermUpdate UPDATE permission
	PermUpdate
	// PermDelete DELETE权限
	// PermDelete DELETE permission
	PermDelete
	// PermCreate CREATE权限
	// PermCreate CREATE permission
	PermCreate
	// PermDrop DROP权限
	// PermDrop DROP permission
	PermDrop
	// PermAlter ALTER权限
	// PermAlter ALTER permission
	PermAlter
	// PermGrant GRANT权限
	// PermGrant GRANT permission
	PermGrant
	// PermSuper 超级用户权限
	// PermSuper super user permission
	PermSuper
)

// String 返回权限类型的字符串表示
// String returns string representation of permission type
func (p PermissionType) String() string {
	switch p {
	case PermSelect:
		return "SELECT"
	case PermInsert:
		return "INSERT"
	case PermUpdate:
		return "UPDATE"
	case PermDelete:
		return "DELETE"
	case PermCreate:
		return "CREATE"
	case PermDrop:
		return "DROP"
	case PermAlter:
		return "ALTER"
	case PermGrant:
		return "GRANT"
	case PermSuper:
		return "SUPER"
	default:
		return fmt.Sprintf("Unknown PermissionType(%d)", int(p))
	}
}

// IsValid 检查权限类型是否有效
// IsValid checks if permission type is valid
func (p PermissionType) IsValid() bool {
	return p >= PermSelect && p <= PermSuper
}

// IsDDL 检查权限是否为DDL类型
// IsDDL checks if permission is DDL type
func (p PermissionType) IsDDL() bool {
	return p == PermCreate || p == PermDrop || p == PermAlter
}

// IsDML 检查权限是否为DML类型
// IsDML checks if permission is DML type
func (p PermissionType) IsDML() bool {
	return p == PermSelect || p == PermInsert || p == PermUpdate || p == PermDelete
}

// IsAdmin 检查权限是否为管理类型
// IsAdmin checks if permission is admin type
func (p PermissionType) IsAdmin() bool {
	return p == PermGrant || p == PermSuper
}

// ConnectionStatus 连接状态枚举
// ConnectionStatus enumeration for connection states
type ConnectionStatus int

const (
	// ConnDisconnected 断开连接状态
	// ConnDisconnected disconnected state
	ConnDisconnected ConnectionStatus = iota
	// ConnConnecting 连接中状态
	// ConnConnecting connecting state
	ConnConnecting
	// ConnAuthenticating 认证中状态
	// ConnAuthenticating authenticating state
	ConnAuthenticating
	// ConnConnected 已连接状态
	// ConnConnected connected state
	ConnConnected
	// ConnError 连接错误状态
	// ConnError connection error state
	ConnError
)

// String 返回连接状态的字符串表示
// String returns string representation of connection status
func (c ConnectionStatus) String() string {
	switch c {
	case ConnDisconnected:
		return "Disconnected"
	case ConnConnecting:
		return "Connecting"
	case ConnAuthenticating:
		return "Authenticating"
	case ConnConnected:
		return "Connected"
	case ConnError:
		return "Error"
	default:
		return fmt.Sprintf("Unknown ConnectionStatus(%d)", int(c))
	}
}

// IsValid 检查连接状态是否有效
// IsValid checks if connection status is valid
func (c ConnectionStatus) IsValid() bool {
	return c >= ConnDisconnected && c <= ConnError
}

// IsActive 检查连接是否处于活跃状态
// IsActive checks if connection is in active state
func (c ConnectionStatus) IsActive() bool {
	return c == ConnConnected
}

// QueryType 查询类型枚举
// QueryType enumeration for query types
type QueryType int

const (
	// QuerySelect SELECT查询
	// QuerySelect SELECT query
	QuerySelect QueryType = iota
	// QueryInsert INSERT查询
	// QueryInsert INSERT query
	QueryInsert
	// QueryUpdate UPDATE查询
	// QueryUpdate UPDATE query
	QueryUpdate
	// QueryDelete DELETE查询
	// QueryDelete DELETE query
	QueryDelete
	// QueryDDL DDL查询（CREATE/DROP/ALTER等）
	// QueryDDL DDL query (CREATE/DROP/ALTER etc.)
	QueryDDL
	// QueryDCL DCL查询（GRANT/REVOKE等）
	// QueryDCL DCL query (GRANT/REVOKE etc.)
	QueryDCL
	// QueryTCL TCL查询（BEGIN/COMMIT/ROLLBACK等）
	// QueryTCL TCL query (BEGIN/COMMIT/ROLLBACK etc.)
	QueryTCL
	// QueryUtility 工具查询（SHOW/DESCRIBE等）
	// QueryUtility utility query (SHOW/DESCRIBE etc.)
	QueryUtility
)

// String 返回查询类型的字符串表示
// String returns string representation of query type
func (q QueryType) String() string {
	switch q {
	case QuerySelect:
		return "SELECT"
	case QueryInsert:
		return "INSERT"
	case QueryUpdate:
		return "UPDATE"
	case QueryDelete:
		return "DELETE"
	case QueryDDL:
		return "DDL"
	case QueryDCL:
		return "DCL"
	case QueryTCL:
		return "TCL"
	case QueryUtility:
		return "UTILITY"
	default:
		return fmt.Sprintf("Unknown QueryType(%d)", int(q))
	}
}

// IsValid 检查查询类型是否有效
// IsValid checks if query type is valid
func (q QueryType) IsValid() bool {
	return q >= QuerySelect && q <= QueryUtility
}

// IsDML 检查查询是否为DML类型
// IsDML checks if query is DML type
func (q QueryType) IsDML() bool {
	return q >= QuerySelect && q <= QueryDelete
}

// RequiresTransaction 检查查询是否需要事务
// RequiresTransaction checks if query requires transaction
func (q QueryType) RequiresTransaction() bool {
	return q == QueryInsert || q == QueryUpdate || q == QueryDelete || q == QueryDDL
}

// IsReadOnly 检查查询是否为只读
// IsReadOnly checks if query is read-only
func (q QueryType) IsReadOnly() bool {
	return q == QuerySelect || q == QueryUtility
}

// EngineStatus 引擎状态枚举
// EngineStatus enumeration for engine states
type EngineStatus int

const (
	// EngineInit 引擎初始化状态
	// EngineInit engine initialization state
	EngineInit EngineStatus = iota
	// EngineStarting 引擎启动中状态
	// EngineStarting engine starting state
	EngineStarting
	// EngineRunning 引擎运行状态
	// EngineRunning engine running state
	EngineRunning
	// EngineStopping 引擎停止中状态
	// EngineStopping engine stopping state
	EngineStopping
	// EngineStopped 引擎已停止状态
	// EngineStopped engine stopped state
	EngineStopped
	// EngineError 引擎错误状态
	// EngineError engine error state
	EngineError
)

// String 返回引擎状态的字符串表示
// String returns string representation of engine status
func (e EngineStatus) String() string {
	switch e {
	case EngineInit:
		return "Init"
	case EngineStarting:
		return "Starting"
	case EngineRunning:
		return "Running"
	case EngineStopping:
		return "Stopping"
	case EngineStopped:
		return "Stopped"
	case EngineError:
		return "Error"
	default:
		return fmt.Sprintf("Unknown EngineStatus(%d)", int(e))
	}
}

// IsValid 检查引擎状态是否有效
// IsValid checks if engine status is valid
func (e EngineStatus) IsValid() bool {
	return e >= EngineInit && e <= EngineError
}

// IsHealthy 检查引擎是否健康
// IsHealthy checks if engine is healthy
func (e EngineStatus) IsHealthy() bool {
	return e == EngineRunning
}

// CanAcceptRequests 检查引擎是否可以接受请求
// CanAcceptRequests checks if engine can accept requests
func (e EngineStatus) CanAcceptRequests() bool {
	return e == EngineRunning
}
