package enum

// DatabaseStatus 定义数据库的运行状态
// DatabaseStatus defines the operational status of a database.
type DatabaseStatus int

const (
	// DatabaseStatusUnknown 未知状态
	// DatabaseStatusUnknown Unknown status.
	DatabaseStatusUnknown DatabaseStatus = iota
	// DatabaseStatusOnline 数据库在线可用
	// DatabaseStatusOnline Database is online and available.
	DatabaseStatusOnline
	// DatabaseStatusOffline 数据库离线
	// DatabaseStatusOffline Database is offline.
	DatabaseStatusOffline
	// DatabaseStatusMaintenance 数据库处于维护模式
	// DatabaseStatusMaintenance Database is in maintenance mode.
	DatabaseStatusMaintenance
	// DatabaseStatusStarting 数据库正在启动中
	// DatabaseStatusStarting Database is starting up.
	DatabaseStatusStarting
	// DatabaseStatusStopping 数据库正在停止中
	// DatabaseStatusStopping Database is shutting down.
	DatabaseStatusStopping
)

// String 返回数据库状态的可读字符串表示
// String returns the readable string representation of the database status.
func (s DatabaseStatus) String() string {
	switch s {
	case DatabaseStatusOnline:
		return "Online"
	case DatabaseStatusOffline:
		return "Offline"
	case DatabaseStatusMaintenance:
		return "Maintenance"
	case DatabaseStatusStarting:
		return "Starting"
	case DatabaseStatusStopping:
		return "Stopping"
	default:
		return "Unknown"
	}
}

// SQLDataType 定义内部支持的 SQL 数据类型代码
// SQLDataType defines the codes for supported SQL data types internally.
// 注意：这些类型可能需要映射到 go-mysql-server 的 sql.Type 或其他具体实现
// Note: These types might need mapping to go-mysql-server's sql.Type or other specific implementations.
type SQLDataType int

const (
	// SQLDataTypeUnknown 未知类型
	// SQLDataTypeUnknown Unknown type.
	SQLDataTypeUnknown SQLDataType = iota
	// SQLDataTypeNull NULL 类型
	// SQLDataTypeNull NULL type.
	SQLDataTypeNull
	// SQLDataTypeInt8 TINYINT 类型
	// SQLDataTypeInt8 TINYINT type.
	SQLDataTypeInt8
	// SQLDataTypeInt16 SMALLINT 类型
	// SQLDataTypeInt16 SMALLINT type.
	SQLDataTypeInt16
	// SQLDataTypeInt32 INT 类型
	// SQLDataTypeInt32 INT type.
	SQLDataTypeInt32
	// SQLDataTypeInt64 BIGINT 类型
	// SQLDataTypeInt64 BIGINT type.
	SQLDataTypeInt64
	// SQLDataTypeUint8 UNSIGNED TINYINT 类型
	// SQLDataTypeUint8 UNSIGNED TINYINT type.
	SQLDataTypeUint8
	// SQLDataTypeUint16 UNSIGNED SMALLINT 类型
	// SQLDataTypeUint16 UNSIGNED SMALLINT type.
	SQLDataTypeUint16
	// SQLDataTypeUint32 UNSIGNED INT 类型
	// SQLDataTypeUint32 UNSIGNED INT type.
	SQLDataTypeUint32
	// SQLDataTypeUint64 UNSIGNED BIGINT 类型
	// SQLDataTypeUint64 UNSIGNED BIGINT type.
	SQLDataTypeUint64
	// SQLDataTypeFloat32 FLOAT 类型
	// SQLDataTypeFloat32 FLOAT type.
	SQLDataTypeFloat32
	// SQLDataTypeFloat64 DOUBLE 类型
	// SQLDataTypeFloat64 DOUBLE type.
	SQLDataTypeFloat64
	// SQLDataTypeDecimal DECIMAL/NUMERIC 类型
	// SQLDataTypeDecimal DECIMAL/NUMERIC type.
	SQLDataTypeDecimal
	// SQLDataTypeBoolean BOOLEAN 类型
	// SQLDataTypeBoolean BOOLEAN type.
	SQLDataTypeBoolean
	// SQLDataTypeVarchar VARCHAR 类型
	// SQLDataTypeVarchar VARCHAR type.
	SQLDataTypeVarchar
	// SQLDataTypeChar CHAR 类型
	// SQLDataTypeChar CHAR type.
	SQLDataTypeChar
	// SQLDataTypeText TEXT 类型
	// SQLDataTypeText TEXT type.
	SQLDataTypeText
	// SQLDataTypeBlob BLOB 类型
	// SQLDataTypeBlob BLOB type.
	SQLDataTypeBlob
	// SQLDataTypeDate DATE 类型
	// SQLDataTypeDate DATE type.
	SQLDataTypeDate
	// SQLDataTypeTime TIME 类型
	// SQLDataTypeTime TIME type.
	SQLDataTypeTime
	// SQLDataTypeTimestamp TIMESTAMP 类型
	// SQLDataTypeTimestamp TIMESTAMP type.
	SQLDataTypeTimestamp
	// SQLDataTypeDateTime DATETIME 类型
	// SQLDataTypeDateTime DATETIME type.
	SQLDataTypeDateTime
	// SQLDataTypeJSON JSON 类型
	// SQLDataTypeJSON JSON type.
	SQLDataTypeJSON
	// SQLDataTypeEnum ENUM 类型 (字符串表示)
	// SQLDataTypeEnum ENUM type (represented as string).
	SQLDataTypeEnum
	// SQLDataTypeSet SET 类型 (字符串表示)
	// SQLDataTypeSet SET type (represented as string).
	SQLDataTypeSet
	// ... 其他未来可能支持的类型
	// ... other types that might be supported in the future
)

// String 返回 SQL 数据类型的可读字符串表示
// String returns the readable string representation of the SQL data type.
func (t SQLDataType) String() string {
	// 这里可以根据需要实现完整的映射
	// A full mapping can be implemented here as needed.
	switch t {
	case SQLDataTypeNull:
		return "NULL"
	case SQLDataTypeInt8:
		return "TINYINT"
	case SQLDataTypeInt16:
		return "SMALLINT"
	case SQLDataTypeInt32:
		return "INT"
	case SQLDataTypeInt64:
		return "BIGINT"
	case SQLDataTypeUint8:
		return "TINYINT UNSIGNED"
	case SQLDataTypeUint16:
		return "SMALLINT UNSIGNED"
	case SQLDataTypeUint32:
		return "INT UNSIGNED"
	case SQLDataTypeUint64:
		return "BIGINT UNSIGNED"
	case SQLDataTypeFloat32:
		return "FLOAT"
	case SQLDataTypeFloat64:
		return "DOUBLE"
	case SQLDataTypeDecimal:
		return "DECIMAL"
	case SQLDataTypeBoolean:
		return "BOOLEAN"
	case SQLDataTypeVarchar:
		return "VARCHAR"
	case SQLDataTypeChar:
		return "CHAR"
	case SQLDataTypeText:
		return "TEXT"
	case SQLDataTypeBlob:
		return "BLOB"
	case SQLDataTypeDate:
		return "DATE"
	case SQLDataTypeTime:
		return "TIME"
	case SQLDataTypeTimestamp:
		return "TIMESTAMP"
	case SQLDataTypeDateTime:
		return "DATETIME"
	case SQLDataTypeJSON:
		return "JSON"
	case SQLDataTypeEnum:
		return "ENUM"
	case SQLDataTypeSet:
		return "SET"
	default:
		return "Unknown"
	}
}

// TxnIsolationLevel 定义事务隔离级别
// TxnIsolationLevel defines the transaction isolation levels.
type TxnIsolationLevel int

const (
	// TxnIsolationLevelReadUncommitted 读未提交
	// TxnIsolationLevelReadUncommitted Read Uncommitted.
	TxnIsolationLevelReadUncommitted TxnIsolationLevel = iota
	// TxnIsolationLevelReadCommitted 读已提交
	// TxnIsolationLevelReadCommitted Read Committed.
	TxnIsolationLevelReadCommitted
	// TxnIsolationLevelRepeatableRead 可重复读
	// TxnIsolationLevelRepeatableRead Repeatable Read.
	TxnIsolationLevelRepeatableRead
	// TxnIsolationLevelSerializable 串行化
	// TxnIsolationLevelSerializable Serializable.
	TxnIsolationLevelSerializable
)

// String 返回事务隔离级别的可读字符串表示
// String returns the readable string representation of the transaction isolation level.
func (l TxnIsolationLevel) String() string {
	switch l {
	case TxnIsolationLevelReadUncommitted:
		return "READ UNCOMMITTED"
	case TxnIsolationLevelReadCommitted:
		return "READ COMMITTED"
	case TxnIsolationLevelRepeatableRead:
		return "REPEATABLE READ"
	case TxnIsolationLevelSerializable:
		return "SERIALIZABLE"
	default:
		return "UNKNOWN"
	}
}

// StorageEngineType 定义支持的存储引擎类型
// StorageEngineType defines the supported storage engine types.
type StorageEngineType int

const (
	// StorageEngineTypeUnknown 未知引擎
	// StorageEngineTypeUnknown Unknown engine.
	StorageEngineTypeUnknown StorageEngineType = iota
	// StorageEngineTypeBadger Badger KV 存储引擎
	// StorageEngineTypeBadger Badger KV storage engine.
	StorageEngineTypeBadger
	// StorageEngineTypeKVD KVD 存储引擎 (占位符)
	// StorageEngineTypeKVD KVD storage engine (placeholder).
	StorageEngineTypeKVD
	// StorageEngineTypeMDD MDD 存储引擎 (占位符)
	// StorageEngineTypeMDD MDD storage engine (placeholder).
	StorageEngineTypeMDD
	// StorageEngineTypeMDI MDI 存储引擎 (占位符)
	// StorageEngineTypeMDI MDI storage engine (placeholder).
	StorageEngineTypeMDI
	// StorageEngineTypeMemory 内存存储引擎 (主要用于测试或元数据)
	// StorageEngineTypeMemory In-memory storage engine (mainly for testing or metadata).
	StorageEngineTypeMemory
)

// String 返回存储引擎类型的可读字符串表示
// String returns the readable string representation of the storage engine type.
func (t StorageEngineType) String() string {
	switch t {
	case StorageEngineTypeBadger:
		return "Badger"
	case StorageEngineTypeKVD:
		return "KVD"
	case StorageEngineTypeMDD:
		return "MDD"
	case StorageEngineTypeMDI:
		return "MDI"
	case StorageEngineTypeMemory:
		return "Memory"
	default:
		return "Unknown"
	}
}

// OperationStatus 定义操作的执行状态
// OperationStatus defines the execution status of an operation.
type OperationStatus int

const (
	// OperationStatusUnknown 未知状态
	// OperationStatusUnknown Unknown status.
	OperationStatusUnknown OperationStatus = iota
	// OperationStatusSuccess 操作成功
	// OperationStatusSuccess Operation succeeded.
	OperationStatusSuccess
	// OperationStatusFailure 操作失败
	// OperationStatusFailure Operation failed.
	OperationStatusFailure
	// OperationStatusPending 操作待处理
	// OperationStatusPending Operation is pending.
	OperationStatusPending
	// OperationStatusRunning 操作正在进行中
	// OperationStatusRunning Operation is in progress.
	OperationStatusRunning
	// OperationStatusCancelled 操作已取消
	// OperationStatusCancelled Operation was cancelled.
	OperationStatusCancelled
)

// String 返回操作状态的可读字符串表示
// String returns the readable string representation of the operation status.
func (s OperationStatus) String() string {
	switch s {
	case OperationStatusSuccess:
		return "Success"
	case OperationStatusFailure:
		return "Failure"
	case OperationStatusPending:
		return "Pending"
	case OperationStatusRunning:
		return "Running"
	case OperationStatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// PermissionType 定义权限类型 (可以使用位掩码)
// PermissionType defines permission types (can use bitmask).
type PermissionType uint64

const (
	// PermNone 无权限
	// PermNone No permission.
	PermNone PermissionType = 0
	// PermSelect SELECT 权限
	// PermSelect SELECT permission.
	PermSelect PermissionType = 1 << 0
	// PermInsert INSERT 权限
	// PermInsert INSERT permission.
	PermInsert PermissionType = 1 << 1
	// PermUpdate UPDATE 权限
	// PermUpdate UPDATE permission.
	PermUpdate PermissionType = 1 << 2
	// PermDelete DELETE 权限
	// PermDelete DELETE permission.
	PermDelete PermissionType = 1 << 3
	// PermCreate CREATE 权限 (数据库, 表, 索引, 用户等)
	// PermCreate CREATE permission (database, table, index, user, etc.).
	PermCreate PermissionType = 1 << 4
	// PermDrop DROP 权限 (数据库, 表, 索引, 用户等)
	// PermDrop DROP permission (database, table, index, user, etc.).
	PermDrop PermissionType = 1 << 5
	// PermAlter ALTER 权限 (表, 数据库等)
	// PermAlter ALTER permission (table, database, etc.).
	PermAlter PermissionType = 1 << 6
	// PermGrant GRANT OPTION 权限 (授予权限)
	// PermGrant GRANT OPTION permission (to grant permissions).
	PermGrant PermissionType = 1 << 7
	// PermShowDB SHOW DATABASES 权限
	// PermShowDB SHOW DATABASES permission.
	PermShowDB PermissionType = 1 << 8
	// PermProcess PROCESS 权限 (查看进程)
	// PermProcess PROCESS permission (view processes).
	PermProcess PermissionType = 1 << 9
	// PermSuper SUPER 权限 (管理权限)
	// PermSuper SUPER permission (administrative privileges).
	PermSuper PermissionType = 1 << 10
	// PermAll 所有权限 (通常用于管理员)
	// PermAll All privileges (typically for administrators).
	PermAll PermissionType = PermSelect | PermInsert | PermUpdate | PermDelete | PermCreate | PermDrop | PermAlter | PermGrant | PermShowDB | PermProcess | PermSuper // 合并所有基础权限 Combine all basic permissions
)

// Has 返回给定权限位掩码是否包含特定权限
// Has returns whether the given permission bitmask contains a specific permission.
func (p PermissionType) Has(perm PermissionType) bool {
	return p&perm == perm
}

// String 返回权限类型的可读字符串表示（可能包含多个权限）
// String returns a readable string representation of the permission type (may include multiple permissions).
func (p PermissionType) String() string {
	if p == PermNone {
		return "NONE"
	}
	if p == PermAll {
		return "ALL PRIVILEGES" // 或者根据 MySQL 习惯返回 GRANT OPTION ?
		// Or return GRANT OPTION according to MySQL convention?
	}

	var permissions []string
	if p.Has(PermSelect) {
		permissions = append(permissions, "SELECT")
	}
	if p.Has(PermInsert) {
		permissions = append(permissions, "INSERT")
	}
	if p.Has(PermUpdate) {
		permissions = append(permissions, "UPDATE")
	}
	if p.Has(PermDelete) {
		permissions = append(permissions, "DELETE")
	}
	if p.Has(PermCreate) {
		permissions = append(permissions, "CREATE")
	}
	if p.Has(PermDrop) {
		permissions = append(permissions, "DROP")
	}
	if p.Has(PermAlter) {
		permissions = append(permissions, "ALTER")
	}
	if p.Has(PermGrant) {
		permissions = append(permissions, "GRANT OPTION")
	}
	if p.Has(PermShowDB) {
		permissions = append(permissions, "SHOW DATABASES")
	}
	if p.Has(PermProcess) {
		permissions = append(permissions, "PROCESS")
	}
	if p.Has(PermSuper) {
		permissions = append(permissions, "SUPER")
	}

	// 使用 Go 标准库 strings.Join 可能会引入不必要的依赖，这里简单实现
	// Using strings.Join from the Go standard library might introduce unnecessary dependencies, simple implementation here.
	result := ""
	for i, permStr := range permissions {
		if i > 0 {
			result += ", "
		}
		result += permStr
	}
	return result
}
