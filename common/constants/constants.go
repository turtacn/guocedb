// Package constants 定义GuoceDB项目的全局常量
// Package constants defines global constants for GuoceDB project
package constants

import "time"

// 版本信息常量
// Version information constants
const (
	// DatabaseName 数据库名称
	// DatabaseName database name
	DatabaseName = "GuoceDB"

	// MajorVersion 主版本号
	// MajorVersion major version number
	MajorVersion = 1

	// MinorVersion 次版本号
	// MinorVersion minor version number
	MinorVersion = 0

	// PatchVersion 补丁版本号
	// PatchVersion patch version number
	PatchVersion = 0

	// BuildVersion 构建版本号
	// BuildVersion build version number
	BuildVersion = "alpha"

	// FullVersion 完整版本号
	// FullVersion full version string
	FullVersion = "1.0.0-alpha"

	// ServerVersion MySQL兼容版本号
	// ServerVersion MySQL compatible version
	ServerVersion = "8.0.33-GuoceDB-1.0.0"
)

// 默认网络配置常量
// Default network configuration constants
const (
	// DefaultMySQLPort 默认MySQL协议端口
	// DefaultMySQLPort default MySQL protocol port
	DefaultMySQLPort = 3306

	// DefaultManagementPort 默认管理API端口
	// DefaultManagementPort default management API port
	DefaultManagementPort = 8080

	// DefaultGRPCPort 默认gRPC端口
	// DefaultGRPCPort default gRPC port
	DefaultGRPCPort = 9090

	// DefaultMetricsPort 默认指标端口
	// DefaultMetricsPort default metrics port
	DefaultMetricsPort = 9091

	// DefaultBindAddress 默认绑定地址
	// DefaultBindAddress default bind address
	DefaultBindAddress = "0.0.0.0"

	// DefaultUnixSocket 默认Unix Socket路径
	// DefaultUnixSocket default Unix socket path
	DefaultUnixSocket = "/tmp/guocedb.sock"
)

// 超时配置常量
// Timeout configuration constants
const (
	// DefaultConnectTimeout 默认连接超时时间
	// DefaultConnectTimeout default connection timeout
	DefaultConnectTimeout = 30 * time.Second

	// DefaultReadTimeout 默认读取超时时间
	// DefaultReadTimeout default read timeout
	DefaultReadTimeout = 30 * time.Second

	// DefaultWriteTimeout 默认写入超时时间
	// DefaultWriteTimeout default write timeout
	DefaultWriteTimeout = 30 * time.Second

	// DefaultQueryTimeout 默认查询超时时间
	// DefaultQueryTimeout default query timeout
	DefaultQueryTimeout = 300 * time.Second

	// DefaultTransactionTimeout 默认事务超时时间
	// DefaultTransactionTimeout default transaction timeout
	DefaultTransactionTimeout = 600 * time.Second

	// DefaultLockTimeout 默认锁超时时间
	// DefaultLockTimeout default lock timeout
	DefaultLockTimeout = 60 * time.Second

	// DefaultKeepAliveTimeout 默认保活超时时间
	// DefaultKeepAliveTimeout default keep-alive timeout
	DefaultKeepAliveTimeout = 300 * time.Second
)

// 缓存配置常量
// Cache configuration constants
const (
	// DefaultCacheSize 默认缓存大小（字节）
	// DefaultCacheSize default cache size in bytes
	DefaultCacheSize = 256 * 1024 * 1024 // 256MB

	// DefaultMetadataCacheSize 默认元数据缓存大小
	// DefaultMetadataCacheSize default metadata cache size
	DefaultMetadataCacheSize = 64 * 1024 * 1024 // 64MB

	// DefaultQueryCacheSize 默认查询缓存大小
	// DefaultQueryCacheSize default query cache size
	DefaultQueryCacheSize = 128 * 1024 * 1024 // 128MB

	// DefaultCacheTTL 默认缓存TTL
	// DefaultCacheTTL default cache TTL
	DefaultCacheTTL = 1 * time.Hour

	// DefaultCacheCleanupInterval 默认缓存清理间隔
	// DefaultCacheCleanupInterval default cache cleanup interval
	DefaultCacheCleanupInterval = 10 * time.Minute
)

// MySQL协议常量
// MySQL protocol constants
const (
	// MySQLProtocolVersion MySQL协议版本
	// MySQLProtocolVersion MySQL protocol version
	MySQLProtocolVersion = 10

	// MySQLServerCapabilities MySQL服务器能力标志
	// MySQLServerCapabilities MySQL server capability flags
	MySQLServerCapabilities = 0x807ff7ff

	// MySQLCharsetCollation 默认字符集排序规则
	// MySQLCharsetCollation default charset collation
	MySQLCharsetCollation = 33 // utf8_general_ci

	// MySQLServerStatus 默认服务器状态
	// MySQLServerStatus default server status
	MySQLServerStatus = 0x0002 // SERVER_STATUS_AUTOCOMMIT

	// MySQLMaxPacketSize 最大包大小
	// MySQLMaxPacketSize maximum packet size
	MySQLMaxPacketSize = 16777216 // 16MB

	// MySQLDefaultAuthPlugin 默认认证插件
	// MySQLDefaultAuthPlugin default authentication plugin
	MySQLDefaultAuthPlugin = "mysql_native_password"
)

// MySQL命令类型常量
// MySQL command type constants
const (
	// MySQLComQuit 退出命令
	// MySQLComQuit quit command
	MySQLComQuit = 0x01

	// MySQLComInitDB 初始化数据库命令
	// MySQLComInitDB init database command
	MySQLComInitDB = 0x02

	// MySQLComQuery 查询命令
	// MySQLComQuery query command
	MySQLComQuery = 0x03

	// MySQLComFieldList 字段列表命令
	// MySQLComFieldList field list command
	MySQLComFieldList = 0x04

	// MySQLComCreateDB 创建数据库命令
	// MySQLComCreateDB create database command
	MySQLComCreateDB = 0x05

	// MySQLComDropDB 删除数据库命令
	// MySQLComDropDB drop database command
	MySQLComDropDB = 0x06

	// MySQLComPing Ping命令
	// MySQLComPing ping command
	MySQLComPing = 0x0e

	// MySQLComStmtPrepare 预处理语句命令
	// MySQLComStmtPrepare statement prepare command
	MySQLComStmtPrepare = 0x16

	// MySQLComStmtExecute 执行预处理语句命令
	// MySQLComStmtExecute statement execute command
	MySQLComStmtExecute = 0x17

	// MySQLComStmtClose 关闭预处理语句命令
	// MySQLComStmtClose statement close command
	MySQLComStmtClose = 0x19
)

// 存储引擎配置常量
// Storage engine configuration constants
const (
	// DefaultStorageEngine 默认存储引擎
	// DefaultStorageEngine default storage engine
	DefaultStorageEngine = "badger"

	// BadgerDefaultDir Badger默认数据目录
	// BadgerDefaultDir Badger default data directory
	BadgerDefaultDir = "./data/badger"

	// BadgerDefaultMemTableSize Badger默认内存表大小
	// BadgerDefaultMemTableSize Badger default memtable size
	BadgerDefaultMemTableSize = 64 * 1024 * 1024 // 64MB

	// BadgerDefaultValueLogFileSize Badger默认值日志文件大小
	// BadgerDefaultValueLogFileSize Badger default value log file size
	BadgerDefaultValueLogFileSize = 128 * 1024 * 1024 // 128MB

	// BadgerDefaultGCDiscardRatio Badger默认GC丢弃比率
	// BadgerDefaultGCDiscardRatio Badger default GC discard ratio
	BadgerDefaultGCDiscardRatio = 0.5

	// BadgerDefaultSyncWrites Badger默认同步写入
	// BadgerDefaultSyncWrites Badger default sync writes
	BadgerDefaultSyncWrites = true
)

// 系统限制常量
// System limit constants
const (
	// MaxConnections 最大连接数
	// MaxConnections maximum number of connections
	MaxConnections = 1000

	// MaxDatabases 最大数据库数量
	// MaxDatabases maximum number of databases
	MaxDatabases = 1000

	// MaxTablesPerDatabase 每个数据库最大表数量
	// MaxTablesPerDatabase maximum tables per database
	MaxTablesPerDatabase = 1000

	// MaxColumnsPerTable 每个表最大列数
	// MaxColumnsPerTable maximum columns per table
	MaxColumnsPerTable = 1000

	// MaxIndexesPerTable 每个表最大索引数
	// MaxIndexesPerTable maximum indexes per table
	MaxIndexesPerTable = 100

	// MaxKeyLength 最大键长度
	// MaxKeyLength maximum key length
	MaxKeyLength = 1024

	// MaxValueLength 最大值长度
	// MaxValueLength maximum value length
	MaxValueLength = 64 * 1024 * 1024 // 64MB

	// MaxRowSize 最大行大小
	// MaxRowSize maximum row size
	MaxRowSize = 64 * 1024 * 1024 // 64MB

	// MaxQueryLength 最大查询长度
	// MaxQueryLength maximum query length
	MaxQueryLength = 16 * 1024 * 1024 // 16MB

	// MaxTransactionSize 最大事务大小
	// MaxTransactionSize maximum transaction size
	MaxTransactionSize = 1024 * 1024 * 1024 // 1GB

	// MaxConcurrentTransactions 最大并发事务数
	// MaxConcurrentTransactions maximum concurrent transactions
	MaxConcurrentTransactions = 1000
)

// 字符串长度限制常量
// String length limit constants
const (
	// MaxDatabaseNameLength 最大数据库名长度
	// MaxDatabaseNameLength maximum database name length
	MaxDatabaseNameLength = 64

	// MaxTableNameLength 最大表名长度
	// MaxTableNameLength maximum table name length
	MaxTableNameLength = 64

	// MaxColumnNameLength 最大列名长度
	// MaxColumnNameLength maximum column name length
	MaxColumnNameLength = 64

	// MaxIndexNameLength 最大索引名长度
	// MaxIndexNameLength maximum index name length
	MaxIndexNameLength = 64

	// MaxUserNameLength 最大用户名长度
	// MaxUserNameLength maximum username length
	MaxUserNameLength = 32

	// MaxPasswordLength 最大密码长度
	// MaxPasswordLength maximum password length
	MaxPasswordLength = 256

	// MaxCommentLength 最大注释长度
	// MaxCommentLength maximum comment length
	MaxCommentLength = 1024
)

// 日志配置常量
// Logging configuration constants
const (
	// DefaultLogLevel 默认日志级别
	// DefaultLogLevel default log level
	DefaultLogLevel = "INFO"

	// DefaultLogFormat 默认日志格式
	// DefaultLogFormat default log format
	DefaultLogFormat = "json"

	// DefaultLogFile 默认日志文件
	// DefaultLogFile default log file
	DefaultLogFile = "./logs/guocedb.log"

	// DefaultLogMaxSize 默认日志文件最大大小（MB）
	// DefaultLogMaxSize default log file max size in MB
	DefaultLogMaxSize = 100

	// DefaultLogMaxBackups 默认日志备份文件数量
	// DefaultLogMaxBackups default number of log backup files
	DefaultLogMaxBackups = 10

	// DefaultLogMaxAge 默认日志文件保留天数
	// DefaultLogMaxAge default log file retention days
	DefaultLogMaxAge = 7

	// DefaultLogCompress 默认是否压缩日志
	// DefaultLogCompress default log compression
	DefaultLogCompress = true
)

// 安全配置常量
// Security configuration constants
const (
	// DefaultTLSEnabled 默认是否启用TLS
	// DefaultTLSEnabled default TLS enabled
	DefaultTLSEnabled = false

	// DefaultTLSCertFile 默认TLS证书文件
	// DefaultTLSCertFile default TLS certificate file
	DefaultTLSCertFile = "./certs/server.crt"

	// DefaultTLSKeyFile 默认TLS私钥文件
	// DefaultTLSKeyFile default TLS key file
	DefaultTLSKeyFile = "./certs/server.key"

	// DefaultAuthEnabled 默认是否启用认证
	// DefaultAuthEnabled default authentication enabled
	DefaultAuthEnabled = true

	// DefaultAuthMethod 默认认证方式
	// DefaultAuthMethod default authentication method
	DefaultAuthMethod = "native"

	// DefaultPasswordMinLength 默认密码最小长度
	// DefaultPasswordMinLength default minimum password length
	DefaultPasswordMinLength = 8

	// DefaultPasswordHashRounds 默认密码哈希轮数
	// DefaultPasswordHashRounds default password hash rounds
	DefaultPasswordHashRounds = 12

	// DefaultSessionTimeout 默认会话超时时间
	// DefaultSessionTimeout default session timeout
	DefaultSessionTimeout = 8 * time.Hour

	// DefaultTokenTTL 默认令牌TTL
	// DefaultTokenTTL default token TTL
	DefaultTokenTTL = 1 * time.Hour
)

// 性能配置常量
// Performance configuration constants
const (
	// DefaultMaxWorkers 默认最大工作线程数
	// DefaultMaxWorkers default maximum worker threads
	DefaultMaxWorkers = 100

	// DefaultQueueSize 默认队列大小
	// DefaultQueueSize default queue size
	DefaultQueueSize = 10000

	// DefaultBatchSize 默认批处理大小
	// DefaultBatchSize default batch size
	DefaultBatchSize = 1000

	// DefaultFlushInterval 默认刷新间隔
	// DefaultFlushInterval default flush interval
	DefaultFlushInterval = 5 * time.Second

	// DefaultStatsInterval 默认统计间隔
	// DefaultStatsInterval default statistics interval
	DefaultStatsInterval = 10 * time.Second

	// DefaultHealthCheckInterval 默认健康检查间隔
	// DefaultHealthCheckInterval default health check interval
	DefaultHealthCheckInterval = 30 * time.Second
)

// 数据类型常量
// Data type constants
const (
	// DataTypeUnknown 未知类型
	// DataTypeUnknown unknown type
	DataTypeUnknown = "UNKNOWN"

	// DataTypeNull NULL类型
	// DataTypeNull NULL type
	DataTypeNull = "NULL"

	// DataTypeBool 布尔类型
	// DataTypeBool boolean type
	DataTypeBool = "BOOL"

	// DataTypeInt8 8位整数类型
	// DataTypeInt8 8-bit integer type
	DataTypeInt8 = "TINYINT"

	// DataTypeInt16 16位整数类型
	// DataTypeInt16 16-bit integer type
	DataTypeInt16 = "SMALLINT"

	// DataTypeInt32 32位整数类型
	// DataTypeInt32 32-bit integer type
	DataTypeInt32 = "INT"

	// DataTypeInt64 64位整数类型
	// DataTypeInt64 64-bit integer type
	DataTypeInt64 = "BIGINT"

	// DataTypeFloat32 32位浮点类型
	// DataTypeFloat32 32-bit float type
	DataTypeFloat32 = "FLOAT"

	// DataTypeFloat64 64位浮点类型
	// DataTypeFloat64 64-bit float type
	DataTypeFloat64 = "DOUBLE"

	// DataTypeString 字符串类型
	// DataTypeString string type
	DataTypeString = "VARCHAR"

	// DataTypeText 文本类型
	// DataTypeText text type
	DataTypeText = "TEXT"

	// DataTypeBlob 二进制类型
	// DataTypeBlob binary type
	DataTypeBlob = "BLOB"

	// DataTypeDate 日期类型
	// DataTypeDate date type
	DataTypeDate = "DATE"

	// DataTypeTime 时间类型
	// DataTypeTime time type
	DataTypeTime = "TIME"

	// DataTypeDateTime 日期时间类型
	// DataTypeDateTime datetime type
	DataTypeDateTime = "DATETIME"

	// DataTypeTimestamp 时间戳类型
	// DataTypeTimestamp timestamp type
	DataTypeTimestamp = "TIMESTAMP"
)

// 系统数据库和表常量
// System database and table constants
const (
	// SystemDatabase 系统数据库名
	// SystemDatabase system database name
	SystemDatabase = "information_schema"

	// MetadataDatabase 元数据数据库名
	// MetadataDatabase metadata database name
	MetadataDatabase = "guocedb_meta"

	// UsersTable 用户表名
	// UsersTable users table name
	UsersTable = "users"

	// RolesTable 角色表名
	// RolesTable roles table name
	RolesTable = "roles"

	// PermissionsTable 权限表名
	// PermissionsTable permissions table name
	PermissionsTable = "permissions"

	// DatabasesTable 数据库表名
	// DatabasesTable databases table name
	DatabasesTable = "databases"

	// TablesTable 表信息表名
	// TablesTable tables table name
	TablesTable = "tables"

	// ColumnsTable 列信息表名
	// ColumnsTable columns table name
	ColumnsTable = "columns"

	// IndexesTable 索引信息表名
	// IndexesTable indexes table name
	IndexesTable = "indexes"
)

// HTTP状态码映射
// HTTP status code mapping
var HTTPStatusMapping = map[int]string{
	200: "OK",
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	409: "Conflict",
	429: "Too Many Requests",
	500: "Internal Server Error",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
}

// MySQL错误码映射
// MySQL error code mapping
var MySQLErrorMapping = map[int]string{
	1005: "Can't create table (errno: %d)",
	1006: "Can't create database '%s'",
	1007: "Can't create database '%s'; database exists",
	1008: "Can't drop database '%s'; database doesn't exist",
	1044: "Access denied for user '%s'@'%s' to database '%s'",
	1045: "Access denied for user '%s'@'%s' (using password: %s)",
	1046: "No database selected",
	1049: "Unknown database '%s'",
	1050: "Table '%s' already exists",
	1051: "Unknown table '%s'",
	1054: "Unknown column '%s' in '%s'",
	1062: "Duplicate entry '%s' for key '%s'",
	1064: "You have an error in your SQL syntax",
	1146: "Table '%s.%s' doesn't exist",
	1205: "Lock wait timeout exceeded; try restarting transaction",
	1213: "Deadlock found when trying to get lock; try restarting transaction",
}

// 配置文件路径常量
// Configuration file path constants
const (
	// DefaultConfigFile 默认配置文件路径
	// DefaultConfigFile default configuration file path
	DefaultConfigFile = "./configs/guocedb.yaml"

	// DefaultConfigDir 默认配置目录
	// DefaultConfigDir default configuration directory
	DefaultConfigDir = "./configs"

	// DefaultDataDir 默认数据目录
	// DefaultDataDir default data directory
	DefaultDataDir = "./data"

	// DefaultLogDir 默认日志目录
	// DefaultLogDir default log directory
	DefaultLogDir = "./logs"

	// DefaultCertDir 默认证书目录
	// DefaultCertDir default certificate directory
	DefaultCertDir = "./certs"
)

// 环境变量前缀常量
// Environment variable prefix constants
const (
	// EnvPrefix 环境变量前缀
	// EnvPrefix environment variable prefix
	EnvPrefix = "GUOCEDB_"

	// EnvConfigFile 配置文件环境变量
	// EnvConfigFile configuration file environment variable
	EnvConfigFile = "GUOCEDB_CONFIG_FILE"

	// EnvDataDir 数据目录环境变量
	// EnvDataDir data directory environment variable
	EnvDataDir = "GUOCEDB_DATA_DIR"

	// EnvLogLevel 日志级别环境变量
	// EnvLogLevel log level environment variable
	EnvLogLevel = "GUOCEDB_LOG_LEVEL"

	// EnvPort 端口环境变量
	// EnvPort port environment variable
	EnvPort = "GUOCEDB_PORT"
)
