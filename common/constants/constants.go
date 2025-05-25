// Package constants defines global constants used throughout the Guocedb project.
//
// 统一全局常量定义，旨在提高代码的可读性、可维护性和避免魔法数字/字符串。
package constants

// Version defines the current version of the Guocedb application.
// This should be updated for each release.
// 版本号，用于标识 Guocedb 应用的当前版本。
// 每次发布时应更新此常量。
const Version = "0.1.0-alpha"

// DefaultServerPort defines the default port for the Guocedb MySQL server.
// If no specific port is configured, the server will listen on this port.
// Guocedb MySQL 服务器的默认端口。
// 如果未配置特定端口，服务器将监听此端口。
const DefaultServerPort = 3306

// DefaultConfigPath defines the default path for the Guocedb configuration file.
// If the config file is not specified, the system will look for it at this path.
// Guocedb 配置文件的默认路径。
// 如果未指定配置文件，系统将在此路径查找。
const DefaultConfigPath = "./configs/config.yaml"

// DefaultLogFilePath defines the default path for the Guocedb log file.
// All application logs will be written to this file if not configured otherwise.
// Guocedb 日志文件的默认路径。
// 如果未另行配置，所有应用程序日志将写入此文件。
const DefaultLogFilePath = "./logs/guocedb.log"

// DefaultDataDirPath defines the default directory path for storing Guocedb data.
// This includes Badger KV store data and persistent catalog data.
// Guocedb 数据存储的默认目录路径。
// 这包括 Badger KV 存储数据和持久化 Catalog 数据。
const DefaultDataDirPath = "./data"

// MySQLCharsetUTF8mb4 defines the MySQL character set for UTF8mb4.
// This is the recommended charset for full Unicode support.
// MySQL 的 UTF8mb4 字符集。
// 建议使用此字符集以获得完整的 Unicode 支持。
const MySQLCharsetUTF8mb4 = "utf8mb4"

// MySQLCollationUTF8mb4Bin defines the MySQL collation for UTF8mb4_bin.
// This collation provides binary string comparison.
// MySQL 的 UTF8mb4_bin 排序规则。
// 此排序规则提供二进制字符串比较。
const MySQLCollationUTF8mb4Bin = "utf8mb4_bin"

// DefaultMaxConnections defines the default maximum number of concurrent client connections.
// 默认最大并发客户端连接数。
const DefaultMaxConnections = 1000

// ContextKeyTxn defines the context key for transaction objects.
// Used to store and retrieve the active transaction from a Go context.
// 事务对象的上下文键。
// 用于从 Go 上下文中存储和检索活动事务。
const ContextKeyTxn = "guocedb_transaction"

// ContextKeyLogger defines the context key for the logger instance.
// Used to retrieve the logger from a Go context for request-scoped logging.
// 日志器实例的上下文键。
// 用于从 Go 上下文中检索日志器，以便进行请求范围的日志记录。
const ContextKeyLogger = "guocedb_logger"

// DefaultBadgerValueLogFileSize defines the default size of the value log file in Badger (in bytes).
// This impacts how often Badger performs garbage collection.
// Badger 中值日志文件的默认大小（字节）。
// 这会影响 Badger 执行垃圾回收的频率。
const DefaultBadgerValueLogFileSize = 1 << 30 // 1GB

// DefaultBadgerSyncWrites defines whether Badger should sync writes to disk by default.
// Setting this to true ensures durability but can impact write performance.
// Badger 是否默认同步写入磁盘。
// 设置为 true 可确保数据持久性，但可能影响写入性能。
const DefaultBadgerSyncWrites = true

// ProtoBufAPIVersion defines the API version for protobuf definitions.
// protobuf 定义的 API 版本。
const ProtoBufAPIVersion = "v1"

// TableMetadataPrefix defines the prefix for table metadata keys in the KV store.
// KV 存储中表元数据键的前缀。
const TableMetadataPrefix = "meta_table_"

// IndexMetadataPrefix defines the prefix for index metadata keys in the KV store.
// KV 存储中索引元数据键的前缀。
const IndexMetadataPrefix = "meta_index_"

// DataPrefix defines the prefix for actual data keys in the KV store.
// KV 存储中实际数据键的前缀。
const DataPrefix = "data_"

// SystemDatabaseName defines the name of the system database (e.g., for metadata, users).
// 系统数据库的名称（例如，用于元数据、用户）。
const SystemDatabaseName = "guocedb_system"

// DefaultUser defines the default database user.
// 默认数据库用户。
const DefaultUser = "root"

// DefaultPassword defines the default password for the default user.
// 默认用户的默认密码。
const DefaultPassword = "" // For development/testing, production should use strong passwords.

// DefaultAuditLogPath defines the default path for the audit log file.
// 审计日志文件的默认路径。
const DefaultAuditLogPath = "./logs/guocedb_audit.log"
