// Package enum defines global enumeration types.
// enum 包定义了全局枚举类型。
package enum

// StorageEngineType represents the type of storage engine.
// StorageEngineType 表示存储引擎的类型。
type StorageEngineType string

const (
	// StorageEngineBadger indicates the Badger KV storage engine.
	// StorageEngineBadger 表示 Badger KV 存储引擎。
	StorageEngineBadger StorageEngineType = "badger"
	// StorageEngineKVD indicates a generic KV storage engine (placeholder).
	// StorageEngineKVD 表示通用的 KV 存储引擎（占位符）。
	StorageEngineKVD StorageEngineType = "kvd"
	// StorageEngineMDD indicates a generic multi-dimensional data engine (placeholder).
	// StorageEngineMDD 表示通用的多维数据引擎（占位符）。
	StorageEngineMDD StorageEngineType = "mdd"
	// StorageEngineMDI indicates a generic multi-dimensional index engine (placeholder).
	// StorageEngineMDI 表示通用的多维索引引擎（占位符）。
	StorageEngineMDI StorageEngineType = "mdi"
	// StorageEngineMemory indicates the in-memory storage engine (for testing/initial setup).
	// StorageEngineMemory 表示内存存储引擎（用于测试/初始设置）。
	StorageEngineMemory StorageEngineType = "memory" // Represents the GMS default in-memory database
)

// String returns the string representation of the StorageEngineType.
// String 返回 StorageEngineType 的字符串表示。
func (s StorageEngineType) String() string {
	return string(s)
}

// ConfigKey represents keys used in configuration.
// ConfigKey 表示配置中使用的 key。
type ConfigKey string

const (
	// ConfigKeyServerListenAddress is the key for the server listen address.
	// ConfigKeyServerListenAddress 是服务器监听地址的 key。
	ConfigKeyServerListenAddress ConfigKey = "server.listen_address"

	// ConfigKeyStorageEngineType is the key for the primary storage engine type.
	// ConfigKeyStorageEngineType 是主存储引擎类型的 key。
	ConfigKeyStorageEngineType ConfigKey = "storage.engine"

	// ConfigKeyBadgerDataDir is the key for the Badger data directory.
	// ConfigKeyBadgerDataDir 是 Badger 数据目录的 key。
	ConfigKeyBadgerDataDir ConfigKey = "storage.badger.data_dir"

	// ConfigKeyBadgerWALDir is the key for the Badger WAL directory.
	// ConfigKeyBadgerWALDir 是 Badger WAL 目录的 key。
	ConfigKeyBadgerWALDir ConfigKey = "storage.badger.wal_dir"

	// Add more configuration keys as needed.
	// 根据需要添加更多配置 key。
)

// String returns the string representation of the ConfigKey.
// String 返回 ConfigKey 的字符串表示。
func (c ConfigKey) String() string {
	return string(c)
}

// SecurityMechanism represents a security mechanism type.
// SecurityMechanism 表示安全机制类型。
type SecurityMechanism string

const (
	// SecurityMechanismAuthentication is for authentication mechanisms.
	// SecurityMechanismAuthentication 用于认证机制。
	SecurityMechanismAuthentication SecurityMechanism = "authentication"
	// SecurityMechanismAuthorization is for authorization mechanisms.
	// SecurityMechanismAuthorization 用于授权机制。
	SecurityMechanismAuthorization SecurityMechanism = "authorization"
	// SecurityMechanismEncryption is for data encryption mechanisms.
	// SecurityMechanismEncryption 用于数据加密机制。
	SecurityMechanismEncryption SecurityMechanism = "encryption"
	// SecurityMechanismAuditing is for auditing mechanisms.
	// SecurityMechanismAuditing 用于审计机制。
	SecurityMechanismAuditing SecurityMechanism = "auditing"
)

// String returns the string representation of the SecurityMechanism.
// String 返回 SecurityMechanism 的字符串表示。
func (s SecurityMechanism) String() string {
	return string(s)
}