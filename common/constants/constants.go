package constants

// --- Version Information ---
// --- 版本信息 ---
const (
	// ProjectName is the official name of the database.
	// ProjectName 是数据库的官方名称。
	ProjectName = "GuoceDB"
	// Version represents the current version of GuoceDB.
	// Follow semantic versioning (MAJOR.MINOR.PATCH).
	// Version 代表 GuoceDB 的当前版本。
	// 遵循语义化版本控制 (主版本号.次版本号.修订号)。
	Version = "0.1.0-alpha"
)

// --- Network Defaults ---
// --- 网络默认值 ---
const (
	// DefaultPort is the default TCP port GuoceDB listens on for client connections (MySQL protocol).
	// DefaultPort 是 GuoceDB 监听客户端连接（MySQL 协议）的默认 TCP 端口。
	DefaultPort = 3307 // Using a different port than standard MySQL 3306 to avoid conflicts. // 使用与标准 MySQL 3306 不同的端口以避免冲突。
	// DefaultHost is the default host address GuoceDB listens on.
	// DefaultHost 是 GuoceDB 监听的默认主机地址。
	DefaultHost = "127.0.0.1"
	// DefaultUnixSocket is the default path for the Unix domain socket, if enabled.
	// DefaultUnixSocket 是 Unix 域套接字的默认路径（如果启用）。
	DefaultUnixSocket = "/tmp/guocedb.sock" // Example path, might vary based on OS/config. // 示例路径，可能因操作系统/配置而异。
)

// --- System Identifiers ---
// --- 系统标识符 ---
const (
	// SystemDatabaseName is the name of the internal system database holding metadata.
	// SystemDatabaseName 是存储元数据的内部系统数据库的名称。
	SystemDatabaseName = "guocedb_sys" // Internal system db name // 内部系统数据库名称
	// InformationSchemaName is the name of the database implementing the INFORMATION_SCHEMA standard.
	// InformationSchemaName 是实现 INFORMATION_SCHEMA 标准的数据库的名称。
	InformationSchemaName = "information_schema" // Standard SQL name // 标准 SQL 名称
	// PerformanceSchemaName is the name of the database for performance monitoring (placeholder).
	// PerformanceSchemaName 是用于性能监控的数据库的名称（占位符）。
	PerformanceSchemaName = "performance_schema" // MySQL standard name // MySQL 标准名称

	// SysTableUsers stores user account information within SystemDatabaseName.
	// SysTableUsers 在 SystemDatabaseName 中存储用户账户信息。
	SysTableUsers = "users"
	// SysTableDatabases stores database metadata within SystemDatabaseName.
	// SysTableDatabases 在 SystemDatabaseName 中存储数据库元数据。
	SysTableDatabases = "databases"
	// SysTableTables stores table metadata within SystemDatabaseName.
	// SysTableTables 在 SystemDatabaseName 中存储表元数据。
	SysTableTables = "tables"
	// SysTableColumns stores column metadata within SystemDatabaseName.
	// SysTableColumns 在 SystemDatabaseName 中存储列元数据。
	SysTableColumns = "columns"
	// SysTableIndexes stores index metadata within SystemDatabaseName.
	// SysTableIndexes 在 SystemDatabaseName 中存储索引元数据。
	SysTableIndexes = "indexes"
	// SysTableSettings stores system-wide settings within SystemDatabaseName (if applicable).
	// SysTableSettings 在 SystemDatabaseName 中存储系统范围的设置（如果适用）。
	SysTableSettings = "settings"
	// Add more system table names as needed...
	// 根据需要添加更多系统表名...
)

// --- Configuration Keys ---
// These keys might be used in config files (e.g., TOML, YAML) or environment variables.
// --- 配置键 ---
// 这些键可能用于配置文件（例如 TOML、YAML）或环境变量。
const (
	// ConfigKeyDataDir specifies the root directory for database data files.
	// ConfigKeyDataDir 指定数据库数据文件的根目录。
	ConfigKeyDataDir = "data_dir"
	// ConfigKeyHost specifies the listening host address.
	// ConfigKeyHost 指定监听的主机地址。
	ConfigKeyHost = "host"
	// ConfigKeyPort specifies the listening TCP port.
	// ConfigKeyPort 指定监听的 TCP 端口。
	ConfigKeyPort = "port"
	// ConfigKeyUnixSocket specifies the path for the Unix domain socket.
	// ConfigKeyUnixSocket 指定 Unix 域套接字的路径。
	ConfigKeyUnixSocket = "unix_socket"
	// ConfigKeyLogLevel specifies the logging level (e.g., "debug", "info", "warn", "error").
	// ConfigKeyLogLevel 指定日志级别（例如 "debug", "info", "warn", "error"）。
	ConfigKeyLogLevel = "log_level"
	// ConfigKeyLogFile specifies the path to the log file. Use "" or "stdout" for console logging.
	// ConfigKeyLogFile 指定日志文件的路径。使用 "" 或 "stdout" 进行控制台日志记录。
	ConfigKeyLogFile = "log_file"
	// ConfigKeyMaxConnections specifies the maximum number of concurrent client connections.
	// ConfigKeyMaxConnections 指定并发客户端连接的最大数量。
	ConfigKeyMaxConnections = "max_connections"
	// ConfigKeyConnectTimeout specifies the connection timeout in seconds.
	// ConfigKeyConnectTimeout 指定连接超时时间（秒）。
	ConfigKeyConnectTimeout = "connect_timeout"
	// ConfigKeyReadTimeout specifies the network read timeout in seconds.
	// ConfigKeyReadTimeout 指定网络读取超时时间（秒）。
	ConfigKeyReadTimeout = "read_timeout"
	// ConfigKeyWriteTimeout specifies the network write timeout in seconds.
	// ConfigKeyWriteTimeout 指定网络写入超时时间（秒）。
	ConfigKeyWriteTimeout = "write_timeout"
	// ConfigKeyDefaultIsolationLevel specifies the default transaction isolation level.
	// ConfigKeyDefaultIsolationLevel 指定默认事务隔离级别。
	ConfigKeyDefaultIsolationLevel = "default_isolation_level"
	// ConfigKeyDefaultStorageEngine specifies the default storage engine to use for new tables.
	// ConfigKeyDefaultStorageEngine 指定用于新表的默认存储引擎。
	ConfigKeyDefaultStorageEngine = "default_storage_engine"
	// ConfigKeyAuthPlugin specifies the default authentication plugin (e.g., "native_password").
	// ConfigKeyAuthPlugin 指定默认身份验证插件（例如 "native_password"）。
	ConfigKeyAuthPlugin = "auth_plugin"
	// ConfigKeyTLSEnabled enables or disables TLS for client connections.
	// ConfigKeyTLSEnabled 启用或禁用客户端连接的 TLS。
	ConfigKeyTLSEnabled = "tls_enabled"
	// ConfigKeyTLSCertFile specifies the path to the TLS certificate file.
	// ConfigKeyTLSCertFile 指定 TLS 证书文件的路径。
	ConfigKeyTLSCertFile = "tls_cert_file"
	// ConfigKeyTLSKeyFile specifies the path to the TLS key file.
	// ConfigKeyTLSKeyFile 指定 TLS 密钥文件的路径。
	ConfigKeyTLSKeyFile = "tls_key_file"
	// ConfigKeyTLSCAFile specifies the path to the TLS CA file (for client verification).
	// ConfigKeyTLSCAFile 指定 TLS CA 文件的路径（用于客户端验证）。
	ConfigKeyTLSCAFile = "tls_ca_file"
	// Add more config keys as needed...
	// 根据需要添加更多配置键...
)

// --- Default Limits and Values ---
// --- 默认限制和值 ---
const (
	// DefaultMaxConnections is the default limit for concurrent connections if not specified in config.
	// DefaultMaxConnections 是未在配置中指定时的默认并发连接数限制。
	DefaultMaxConnections = 100
	// DefaultMaxQuerySize is the default maximum size (in bytes) for an incoming SQL query.
	// DefaultMaxQuerySize 是传入 SQL 查询的默认最大大小（字节）。
	DefaultMaxQuerySize = 16 * 1024 * 1024 // 16 MiB // 16 MiB
	// DefaultConnectTimeout is the default timeout (in seconds) for establishing a client connection.
	// DefaultConnectTimeout 是建立客户端连接的默认超时时间（秒）。
	DefaultConnectTimeout = 10
	// DefaultReadTimeout is the default network read timeout (in seconds) for client connections. 0 means no timeout.
	// DefaultReadTimeout 是客户端连接的默认网络读取超时时间（秒）。0 表示无超时。
	DefaultReadTimeout = 300
	// DefaultWriteTimeout is the default network write timeout (in seconds) for client connections. 0 means no timeout.
	// DefaultWriteTimeout 是客户端连接的默认网络写入超时时间（秒）。0 表示无超时。
	DefaultWriteTimeout = 300
	// DefaultLockWaitTimeout is the default time (in seconds) a transaction waits for a lock.
	// DefaultLockWaitTimeout 是事务等待锁的默认时间（秒）。
	DefaultLockWaitTimeout = 50 // Similar to MySQL default // 与 MySQL 默认值类似

	// DefaultCharacterSetStr is the default character set for new databases/tables.
	// DefaultCharacterSetStr 是新数据库/表的默认字符集。
	DefaultCharacterSetStr = "utf8mb4"
	// DefaultCollationStr is the default collation for the default character set.
	// DefaultCollationStr 是默认字符集的默认排序规则。
	DefaultCollationStr = "utf8mb4_general_ci" // Common default, consider utf8mb4_0900_ai_ci for newer MySQL compat // 常用默认值，考虑使用 utf8mb4_0900_ai_ci 以兼容较新的 MySQL

	// MaxIdentifierLength is the maximum length for identifiers like table names, column names, etc.
	// MaxIdentifierLength 是标识符（如表名、列名等）的最大长度。
	MaxIdentifierLength = 64 // Common limit in MySQL // MySQL 中的常见限制
	// DefaultDataDir is the default data directory relative to the executable or user home if not specified.
	// DefaultDataDir 是未指定时的默认数据目录（相对于可执行文件或用户主目录）。
	DefaultDataDir = "guocedb_data"
	// DefaultLogLevel is the default logging level if not specified.
	// DefaultLogLevel 是未指定时的默认日志级别。
	DefaultLogLevel = "info"
)

// --- Internal Constants ---
// --- 内部常量 ---
const (
	// BadgerKeyPrefixSeparator is used internally, e.g., in Badger storage keys. Choose a char unlikely in user data.
	// BadgerKeyPrefixSeparator 在内部使用，例如在 Badger 存储键中。选择一个不太可能出现在用户数据中的字符。
	BadgerKeyPrefixSeparator = '/' // Example separator // 示例分隔符
	// MetaDBPrefix is a prefix for keys storing database metadata.
	// MetaDBPrefix 是存储数据库元数据的键的前缀。
	MetaDBPrefix byte = 'd' // Using single byte for efficiency // 为提高效率使用单字节
	// MetaTablePrefix is a prefix for keys storing table metadata.
	// MetaTablePrefix 是存储表元数据的键的前缀。
	MetaTablePrefix byte = 't'
	// MetaColumnPrefix is a prefix for keys storing column metadata (if needed separately).
	// MetaColumnPrefix 是存储列元数据的键的前缀（如果需要单独存储）。
	MetaColumnPrefix byte = 'c'
	// MetaIndexPrefix is a prefix for keys storing index metadata.
	// MetaIndexPrefix 是存储索引元数据的键的前缀。
	MetaIndexPrefix byte = 'i'
	// DataRowPrefix is a prefix for keys storing actual row data.
	// DataRowPrefix 是存储实际行数据的键的前缀。
	DataRowPrefix byte = 'r'
	// IndexEntryPrefix is a prefix for keys storing index entries.
	// IndexEntryPrefix 是存储索引条目的键的前缀。
	IndexEntryPrefix byte = 'x'
	// SequencePrefix is a prefix for keys used for generating sequence IDs.
	// SequencePrefix 是用于生成序列 ID 的键的前缀。
	SequencePrefix byte = 's'

	// RootUserID is the ID reserved for the initial superuser/root account.
	// RootUserID 是为初始超级用户/root 账户保留的 ID。
	RootUserID = 1
	// RootUsername is the default username for the initial superuser.
	// RootUsername 是初始超级用户的默认用户名。
	RootUsername = "root"
)
