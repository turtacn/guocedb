// Package constants 定义了 GuoceDB 系统的全局常量
// Package constants defines global constants for GuoceDB system
package constants

import (
	"time"
)

// ===== 版本信息 Version Information =====

const (
	// Version GuoceDB 版本号
	// Version GuoceDB version number
	Version = "0.1.0"

	// APIVersion API 版本号
	// APIVersion API version number
	APIVersion = "v1"

	// ProtocolVersion MySQL 协议版本
	// ProtocolVersion MySQL protocol version
	ProtocolVersion = "5.7.0"

	// BuildTime 构建时间（由构建脚本注入）
	// BuildTime build time (injected by build script)
	BuildTime = "unknown"

	// GitCommit Git 提交哈希（由构建脚本注入）
	// GitCommit Git commit hash (injected by build script)
	GitCommit = "unknown"
)

// ===== 系统限制 System Limits =====

const (
	// MaxDatabaseNameLength 数据库名最大长度
	// MaxDatabaseNameLength maximum database name length
	MaxDatabaseNameLength = 64

	// MaxTableNameLength 表名最大长度
	// MaxTableNameLength maximum table name length
	MaxTableNameLength = 64

	// MaxColumnNameLength 列名最大长度
	// MaxColumnNameLength maximum column name length
	MaxColumnNameLength = 64

	// MaxIndexNameLength 索引名最大长度
	// MaxIndexNameLength maximum index name length
	MaxIndexNameLength = 64

	// MaxColumnsPerTable 每表最大列数
	// MaxColumnsPerTable maximum columns per table
	MaxColumnsPerTable = 1024

	// MaxIndexesPerTable 每表最大索引数
	// MaxIndexesPerTable maximum indexes per table
	MaxIndexesPerTable = 64

	// MaxKeyLength 索引键最大长度
	// MaxKeyLength maximum index key length
	MaxKeyLength = 3072

	// MaxRowSize 行最大大小（字节）
	// MaxRowSize maximum row size (bytes)
	MaxRowSize = 65535

	// MaxQueryLength 查询语句最大长度
	// MaxQueryLength maximum query length
	MaxQueryLength = 16 * 1024 * 1024 // 16MB

	// MaxConnections 最大连接数
	// MaxConnections maximum number of connections
	MaxConnections = 10000

	// MaxPreparedStatements 最大预处理语句数
	// MaxPreparedStatements maximum prepared statements
	MaxPreparedStatements = 16382

	// MaxTransactionSize 事务最大大小
	// MaxTransactionSize maximum transaction size
	MaxTransactionSize = 100 * 1024 * 1024 // 100MB
)

// ===== 默认值 Default Values =====

const (
	// DefaultPort 默认服务端口
	// DefaultPort default server port
	DefaultPort = 3307

	// DefaultHost 默认监听地址
	// DefaultHost default host address
	DefaultHost = "0.0.0.0"

	// DefaultMaxConnections 默认最大连接数
	// DefaultMaxConnections default maximum connections
	DefaultMaxConnections = 100

	// DefaultConnectionTimeout 默认连接超时时间
	// DefaultConnectionTimeout default connection timeout
	DefaultConnectionTimeout = 10 * time.Second

	// DefaultQueryTimeout 默认查询超时时间
	// DefaultQueryTimeout default query timeout
	DefaultQueryTimeout = 30 * time.Second

	// DefaultIdleTimeout 默认空闲超时时间
	// DefaultIdleTimeout default idle timeout
	DefaultIdleTimeout = 8 * time.Hour

	// DefaultTransactionTimeout 默认事务超时时间
	// DefaultTransactionTimeout default transaction timeout
	DefaultTransactionTimeout = 60 * time.Second

	// DefaultCharset 默认字符集
	// DefaultCharset default character set
	DefaultCharset = "utf8mb4"

	// DefaultCollation 默认排序规则
	// DefaultCollation default collation
	DefaultCollation = "utf8mb4_general_ci"

	// DefaultPageSize 默认页大小
	// DefaultPageSize default page size
	DefaultPageSize = 16 * 1024 // 16KB

	// DefaultBufferPoolSize 默认缓冲池大小
	// DefaultBufferPoolSize default buffer pool size
	DefaultBufferPoolSize = 128 * 1024 * 1024 // 128MB

	// DefaultLogLevel 默认日志级别
	// DefaultLogLevel default log level
	DefaultLogLevel = "info"

	// DefaultDataDir 默认数据目录
	// DefaultDataDir default data directory
	DefaultDataDir = "./data"

	// DefaultLogDir 默认日志目录
	// DefaultLogDir default log directory
	DefaultLogDir = "./logs"

	// DefaultConfigFile 默认配置文件
	// DefaultConfigFile default configuration file
	DefaultConfigFile = "config.yaml"
)

// ===== 存储引擎相关 Storage Engine Related =====

const (
	// StorageEngineBadger Badger 存储引擎
	// StorageEngineBadger Badger storage engine
	StorageEngineBadger = "badger"

	// StorageEngineKVD KVD 存储引擎
	// StorageEngineKVD KVD storage engine
	StorageEngineKVD = "kvd"

	// StorageEngineMDD MDD 存储引擎
	// StorageEngineMDD MDD storage engine
	StorageEngineMDD = "mdd"

	// StorageEngineMDI MDI 存储引擎
	// StorageEngineMDI MDI storage engine
	StorageEngineMDI = "mdi"

	// DefaultStorageEngine 默认存储引擎
	// DefaultStorageEngine default storage engine
	DefaultStorageEngine = StorageEngineBadger
)

// ===== 系统表和数据库 System Tables and Databases =====

const (
	// SystemDatabase 系统数据库名
	// SystemDatabase system database name
	SystemDatabase = "guoce_system"

	// InformationSchema 信息模式数据库名
	// InformationSchema information schema database name
	InformationSchema = "information_schema"

	// PerformanceSchema 性能模式数据库名
	// PerformanceSchema performance schema database name
	PerformanceSchema = "performance_schema"

	// SystemTableUsers 用户表
	// SystemTableUsers users table
	SystemTableUsers = "users"

	// SystemTableDatabases 数据库表
	// SystemTableDatabases databases table
	SystemTableDatabases = "databases"

	// SystemTableTables 表信息表
	// SystemTableTables tables information table
	SystemTableTables = "tables"

	// SystemTableColumns 列信息表
	// SystemTableColumns columns information table
	SystemTableColumns = "columns"

	// SystemTableIndexes 索引信息表
	// SystemTableIndexes indexes information table
	SystemTableIndexes = "indexes"

	// SystemTablePrivileges 权限表
	// SystemTablePrivileges privileges table
	SystemTablePrivileges = "privileges"

	// SystemTableConfig 配置表
	// SystemTableConfig configuration table
	SystemTableConfig = "config"
)

// ===== 事务隔离级别 Transaction Isolation Levels =====

const (
	// IsolationLevelReadUncommitted 读未提交
	// IsolationLevelReadUncommitted read uncommitted
	IsolationLevelReadUncommitted = "READ_UNCOMMITTED"

	// IsolationLevelReadCommitted 读已提交
	// IsolationLevelReadCommitted read committed
	IsolationLevelReadCommitted = "READ_COMMITTED"

	// IsolationLevelRepeatableRead 可重复读
	// IsolationLevelRepeatableRead repeatable read
	IsolationLevelRepeatableRead = "REPEATABLE_READ"

	// IsolationLevelSerializable 串行化
	// IsolationLevelSerializable serializable
	IsolationLevelSerializable = "SERIALIZABLE"

	// DefaultIsolationLevel 默认隔离级别
	// DefaultIsolationLevel default isolation level
	DefaultIsolationLevel = IsolationLevelRepeatableRead
)

// ===== 权限类型 Privilege Types =====

const (
	// PrivilegeSelect SELECT 权限
	// PrivilegeSelect SELECT privilege
	PrivilegeSelect = "SELECT"

	// PrivilegeInsert INSERT 权限
	// PrivilegeInsert INSERT privilege
	PrivilegeInsert = "INSERT"

	// PrivilegeUpdate UPDATE 权限
	// PrivilegeUpdate UPDATE privilege
	PrivilegeUpdate = "UPDATE"

	// PrivilegeDelete DELETE 权限
	// PrivilegeDelete DELETE privilege
	PrivilegeDelete = "DELETE"

	// PrivilegeCreate CREATE 权限
	// PrivilegeCreate CREATE privilege
	PrivilegeCreate = "CREATE"

	// PrivilegeDrop DROP 权限
	// PrivilegeDrop DROP privilege
	PrivilegeDrop = "DROP"

	// PrivilegeAlter ALTER 权限
	// PrivilegeAlter ALTER privilege
	PrivilegeAlter = "ALTER"

	// PrivilegeIndex INDEX 权限
	// PrivilegeIndex INDEX privilege
	PrivilegeIndex = "INDEX"

	// PrivilegeGrant GRANT 权限
	// PrivilegeGrant GRANT privilege
	PrivilegeGrant = "GRANT"

	// PrivilegeSuper SUPER 权限
	// PrivilegeSuper SUPER privilege
	PrivilegeSuper = "SUPER"

	// PrivilegeAll 所有权限
	// PrivilegeAll all privileges
	PrivilegeAll = "ALL"
)

// ===== 网络协议相关 Network Protocol Related =====

const (
	// ProtocolTCP TCP 协议
	// ProtocolTCP TCP protocol
	ProtocolTCP = "tcp"

	// ProtocolUnix Unix Socket 协议
	// ProtocolUnix Unix socket protocol
	ProtocolUnix = "unix"

	// MySQLProtocolVersion MySQL 协议版本号
	// MySQLProtocolVersion MySQL protocol version number
	MySQLProtocolVersion = 10

	// MySQLCapabilityFlags MySQL 能力标志
	// MySQLCapabilityFlags MySQL capability flags
	MySQLCapabilityFlags = 0x807ff7df

	// MaxPacketSize 最大数据包大小
	// MaxPacketSize maximum packet size
	MaxPacketSize = 16 * 1024 * 1024 // 16MB

	// DefaultPacketSize 默认数据包大小
	// DefaultPacketSize default packet size
	DefaultPacketSize = 4 * 1024 // 4KB
)

// ===== 文件和路径相关 File and Path Related =====

const (
	// DataFileExtension 数据文件扩展名
	// DataFileExtension data file extension
	DataFileExtension = ".gdb"

	// IndexFileExtension 索引文件扩展名
	// IndexFileExtension index file extension
	IndexFileExtension = ".idx"

	// LogFileExtension 日志文件扩展名
	// LogFileExtension log file extension
	LogFileExtension = ".log"

	// MetadataFileName 元数据文件名
	// MetadataFileName metadata file name
	MetadataFileName = "metadata.json"

	// LockFileName 锁文件名
	// LockFileName lock file name
	LockFileName = ".lock"

	// WALFilePrefix WAL 文件前缀
	// WALFilePrefix WAL file prefix
	WALFilePrefix = "wal_"

	// BackupFilePrefix 备份文件前缀
	// BackupFilePrefix backup file prefix
	BackupFilePrefix = "backup_"
)

// ===== 缓存相关 Cache Related =====

const (
	// CacheTypeQuery 查询缓存类型
	// CacheTypeQuery query cache type
	CacheTypeQuery = "query"

	// CacheTypePlan 执行计划缓存类型
	// CacheTypePlan plan cache type
	CacheTypePlan = "plan"

	// CacheTypeResult 结果缓存类型
	// CacheTypeResult result cache type
	CacheTypeResult = "result"

	// DefaultCacheSize 默认缓存大小
	// DefaultCacheSize default cache size
	DefaultCacheSize = 64 * 1024 * 1024 // 64MB

	// DefaultCacheTTL 默认缓存过期时间
	// DefaultCacheTTL default cache TTL
	DefaultCacheTTL = 5 * time.Minute

	// MaxCacheKeyLength 缓存键最大长度
	// MaxCacheKeyLength maximum cache key length
	MaxCacheKeyLength = 250
)

// ===== 监控指标相关 Monitoring Metrics Related =====

const (
	// MetricNamespace 指标命名空间
	// MetricNamespace metrics namespace
	MetricNamespace = "guocedb"

	// MetricSubsystemServer 服务器子系统
	// MetricSubsystemServer server subsystem
	MetricSubsystemServer = "server"

	// MetricSubsystemQuery 查询子系统
	// MetricSubsystemQuery query subsystem
	MetricSubsystemQuery = "query"

	// MetricSubsystemStorage 存储子系统
	// MetricSubsystemStorage storage subsystem
	MetricSubsystemStorage = "storage"

	// MetricSubsystemTransaction 事务子系统
	// MetricSubsystemTransaction transaction subsystem
	MetricSubsystemTransaction = "transaction"

	// MetricSubsystemCache 缓存子系统
	// MetricSubsystemCache cache subsystem
	MetricSubsystemCache = "cache"

	// MetricsUpdateInterval 指标更新间隔
	// MetricsUpdateInterval metrics update interval
	MetricsUpdateInterval = 10 * time.Second
)

// ===== 日志相关 Logging Related =====

const (
	// LogFormatJSON JSON 日志格式
	// LogFormatJSON JSON log format
	LogFormatJSON = "json"

	// LogFormatText 文本日志格式
	// LogFormatText text log format
	LogFormatText = "text"

	// LogRotationDaily 每日日志轮换
	// LogRotationDaily daily log rotation
	LogRotationDaily = "daily"

	// LogRotationSize 按大小日志轮换
	// LogRotationSize size-based log rotation
	LogRotationSize = "size"

	// DefaultLogMaxSize 默认日志文件最大大小
	// DefaultLogMaxSize default log file max size
	DefaultLogMaxSize = 100 * 1024 * 1024 // 100MB

	// DefaultLogMaxBackups 默认日志备份数
	// DefaultLogMaxBackups default log backup count
	DefaultLogMaxBackups = 7

	// DefaultLogMaxAge 默认日志保留天数
	// DefaultLogMaxAge default log retention days
	DefaultLogMaxAge = 30
)

// ===== 编码相关 Encoding Related =====

const (
	// EncodingKeyPrefix 键编码前缀
	// EncodingKeyPrefix key encoding prefix
	EncodingKeyPrefix = byte(0x01)

	// EncodingTablePrefix 表编码前缀
	// EncodingTablePrefix table encoding prefix
	EncodingTablePrefix = byte(0x02)

	// EncodingIndexPrefix 索引编码前缀
	// EncodingIndexPrefix index encoding prefix
	EncodingIndexPrefix = byte(0x03)

	// EncodingMetaPrefix 元数据编码前缀
	// EncodingMetaPrefix metadata encoding prefix
	EncodingMetaPrefix = byte(0x04)

	// EncodingSequencePrefix 序列编码前缀
	// EncodingSequencePrefix sequence encoding prefix
	EncodingSequencePrefix = byte(0x05)

	// EncodingSeparator 编码分隔符
	// EncodingSeparator encoding separator
	EncodingSeparator = byte(0x00)

	// EncodingTerminator 编码终止符
	// EncodingTerminator encoding terminator
	EncodingTerminator = byte(0xFF)
)

// ===== 其他常量 Other Constants =====

const (
	// DefaultBatchSize 默认批处理大小
	// DefaultBatchSize default batch size
	DefaultBatchSize = 1000

	// MaxRetries 最大重试次数
	// MaxRetries maximum retry count
	MaxRetries = 3

	// RetryInterval 重试间隔
	// RetryInterval retry interval
	RetryInterval = 100 * time.Millisecond

	// HealthCheckInterval 健康检查间隔
	// HealthCheckInterval health check interval
	HealthCheckInterval = 5 * time.Second

	// StatisticsUpdateInterval 统计信息更新间隔
	// StatisticsUpdateInterval statistics update interval
	StatisticsUpdateInterval = 5 * time.Minute

	// CompactionInterval 压缩间隔
	// CompactionInterval compaction interval
	CompactionInterval = 1 * time.Hour

	// CheckpointInterval 检查点间隔
	// CheckpointInterval checkpoint interval
	CheckpointInterval = 5 * time.Minute
)

// ===== 魔数和签名 Magic Numbers and Signatures =====

const (
	// FileMagicNumber 文件魔数
	// FileMagicNumber file magic number
	FileMagicNumber = 0x47554F43 // "GUOC" in hex

	// FileVersion 文件版本
	// FileVersion file version
	FileVersion = 1

	// BlockMagicNumber 块魔数
	// BlockMagicNumber block magic number
	BlockMagicNumber = 0x424C4F43 // "BLOC" in hex

	// PageMagicNumber 页魔数
	// PageMagicNumber page magic number
	PageMagicNumber = 0x50414745 // "PAGE" in hex
)
