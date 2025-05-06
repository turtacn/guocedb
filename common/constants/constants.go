// Package constants defines global constants used throughout the guocedb project.
// 定义 Guocedb 项目中使用的全局常量。
package constants

const (
	// DefaultMySQLPort is the default port for the MySQL protocol server.
	// DefaultMySQLPort 是 MySQL 协议服务器的默认端口。
	DefaultMySQLPort = 3306

	// DefaultConfigPath is the default path for the configuration file.
	// DefaultConfigPath 是配置文件的默认路径。
	DefaultConfigPath = "configs/config.yaml"

	// DefaultBadgerDataDir is the default directory for Badger data files.
	// DefaultBadgerDataDir 是 Badger 数据文件的默认目录。
	DefaultBadgerDataDir = "data/badger"

	// DefaultBadgerWALDir is the default directory for Badger Write-Ahead Log files.
	// DefaultBadgerWALDir 是 Badger Write-Ahead Log 文件的默认目录。
	DefaultBadgerWALDir = "wal/badger"

	// DefaultDatabaseName is the default database name if none is specified.
	// DefaultDatabaseName 是在未指定数据库名时的默认数据库名。
	DefaultDatabaseName = "guocedb"

	// SystemDatabaseName is the name of the system database.
	// SystemDatabaseName 是系统数据库的名称。
	SystemDatabaseName = "mysql" // Compatible with GMS system database
)

const (
	// KeySeparator is used to separate parts of keys in KV storage.
	// KeySeparator 用于分隔 KV 存储中的 key 的不同部分。
	KeySeparator byte = '/'

	// NamespaceSeparator is used to separate namespaces in keys.
	// NamespaceSeparator 用于分隔 key 中的命名空间。
	NamespaceSeparator byte = ':'
)

// Namespaces for KV encoding.
// KV 编码中的命名空间。
const (
	// NamespaceCatalog is the namespace for catalog metadata.
	// NamespaceCatalog 是用于目录元数据的命名空间。
	NamespaceCatalog = "catalog"

	// NamespaceData is the namespace for table data.
	// NamespaceData 是用于表数据的命名空间。
	NamespaceData = "data"

	// NamespaceIndex is the namespace for index data.
	// NamespaceIndex 是用于索引数据的命名空间。
	NamespaceIndex = "index"
)