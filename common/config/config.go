// Package config handles the loading, parsing, and management of Guocedb's configuration.
// It provides a structured way to define application settings, allowing for easy
// modification and dynamic behavior based on external configurations (e.g., YAML files,
// environment variables).
//
// 此包负责 Guocedb 配置的加载、解析和管理。
// 它提供了一种结构化的方式来定义应用程序设置，允许基于外部配置（例如，YAML 文件、
// 环境变量）进行轻松修改和动态行为。
package config

import (
	"fmt"
	"os"
	"sync" // For singleton pattern
	"time" // For time.Duration in configurations

	"github.com/turtacn/guocedb/common/constants"  // For default config path
	"github.com/turtacn/guocedb/common/log"        // For logging configuration issues
	"github.com/turtacn/guocedb/common/types/enum" // For enum types used in config
	"gopkg.in/yaml.v3"                             // For YAML parsing
)

// Config represents the overall structure of the Guocedb configuration.
// It includes various sections for server, storage, logging, and security settings.
//
// Config 结构体表示 Guocedb 配置的整体结构。
// 它包括服务器、存储、日志和安全设置的各个部分。
type Config struct {
	// Server defines settings related to the MySQL server listener.
	// 服务器定义了与 MySQL 服务器监听器相关的设置。
	Server ServerConfig `yaml:"server"`
	// Storage defines settings for the underlying data storage engine.
	// 存储定义了底层数据存储引擎的设置。
	Storage StorageConfig `yaml:"storage"`
	// Log defines settings for application logging.
	// 日志定义了应用程序日志记录的设置。
	Log LogConfig `yaml:"log"`
	// Security defines security-related settings like authentication and TLS.
	// 安全定义了与安全相关的设置，如认证和 TLS。
	Security SecurityConfig `yaml:"security"`
	// Audit defines settings for auditing events.
	// 审计定义了审计事件的设置。
	Audit AuditConfig `yaml:"audit"`
}

// ServerConfig defines configuration parameters for the Guocedb MySQL server.
//
// ServerConfig 定义了 Guocedb MySQL 服务器的配置参数。
type ServerConfig struct {
	// Port is the port number the MySQL server will listen on.
	// MySQL 服务器将监听的端口号。
	Port int `yaml:"port"`
	// Host is the network interface the MySQL server will bind to.
	// MySQL 服务器将绑定的网络接口。
	Host string `yaml:"host"`
	// MaxConnections is the maximum number of concurrent client connections allowed.
	// 允许的最大并发客户端连接数。
	MaxConnections int `yaml:"max_connections"`
	// ReadTimeout specifies the duration for reading a complete packet from the client.
	// 从客户端读取完整数据包的持续时间。
	ReadTimeout time.Duration `yaml:"read_timeout"`
	// WriteTimeout specifies the duration for writing a complete packet to the client.
	// 将完整数据包写入客户端的持续时间。
	WriteTimeout time.Duration `yaml:"write_timeout"`
	// Charset specifies the default character set for connections.
	// 连接的默认字符集。
	Charset string `yaml:"charset"`
	// Collation specifies the default collation for connections.
	// 连接的默认排序规则。
	Collation string `yaml:"collation"`
}

// StorageConfig defines configuration parameters for the Guocedb storage engine.
//
// StorageConfig 定义了 Guocedb 存储引擎的配置参数。
type StorageConfig struct {
	// Engine specifies the type of storage engine to use (e.g., Badger).
	// 指定要使用的存储引擎类型（例如，Badger）。
	Engine enum.StorageEngineType `yaml:"engine"`
	// DataDir is the directory path where persistent data will be stored.
	// 持久化数据将存储的目录路径。
	DataDir string `yaml:"data_dir"`
	// Badger specific configurations
	// Badger 特有配置。
	Badger BadgerConfig `yaml:"badger"`
	// KVD specific configurations (placeholder)
	// KVD 特有配置（占位符）。
	KVD KVDConfig `yaml:"kvd"`
}

// BadgerConfig defines specific configurations for the Badger KV store.
//
// BadgerConfig 定义了 Badger KV 存储的特定配置。
type BadgerConfig struct {
	// ValueLogFileSize is the maximum size of a value log file in bytes.
	// 值日志文件的最大大小（字节）。
	ValueLogFileSize int64 `yaml:"value_log_file_size"`
	// SyncWrites determines if writes should be synced to disk immediately.
	// 确定写入是否应立即同步到磁盘。
	SyncWrites bool `yaml:"sync_writes"`
	// VLogGCInterval specifies the interval for value log garbage collection.
	// 值日志垃圾回收的间隔。
	VLogGCInterval time.Duration `yaml:"vlog_gc_interval"`
	// VLogGCDiscardRatio specifies the discard ratio for value log garbage collection.
	// 值日志垃圾回收的丢弃比率。
	VLogGCDiscardRatio float64 `yaml:"vlog_gc_discard_ratio"`
}

// KVDConfig defines specific configurations for a generic Key-Value Database (placeholder).
//
// KVDConfig 定义了通用键值数据库的特定配置（占位符）。
type KVDConfig struct {
	// Path to the KVD data files.
	// KVD 数据文件的路径。
	Path string `yaml:"path"`
}

// LogConfig defines configuration parameters for logging.
//
// LogConfig 定义了日志记录的配置参数。
type LogConfig struct {
	// Level specifies the minimum logging level to output.
	// 指定要输出的最低日志级别。
	Level enum.LogLevel `yaml:"level"`
	// FilePath is the path to the log file. If empty, logs go to stdout.
	// 日志文件的路径。如果为空，日志将输出到标准输出。
	FilePath string `yaml:"file_path"`
}

// SecurityConfig defines security-related configuration parameters.
//
// SecurityConfig 定义了与安全相关的配置参数。
type SecurityConfig struct {
	// DefaultUser is the default database user.
	// 默认数据库用户。
	DefaultUser string `yaml:"default_user"`
	// DefaultPassword is the default password for the default user.
	// 默认用户的默认密码。
	DefaultPassword string `yaml:"default_password"` // Should be hashed in production
	// AuthMethod specifies the authentication method to use.
	// 指定要使用的认证方法。
	AuthMethod enum.AuthMethodType `yaml:"auth_method"`
	// TLSEnabled indicates whether TLS encryption is enabled for connections.
	// 指示是否为连接启用 TLS 加密。
	TLSEnabled bool `yaml:"tls_enabled"`
	// TLSCertPath is the path to the TLS server certificate file.
	// TLS 服务器证书文件的路径。
	TLSCertPath string `yaml:"tls_cert_path"`
	// TLSKeyPath is the path to the TLS server key file.
	// TLS 服务器密钥文件的路径。
	TLSKeyPath string `yaml:"tls_key_path"`
	// TLSCaCertPath is the path to the TLS CA certificate file (for client verification).
	// TLS CA 证书文件的路径（用于客户端验证）。
	TLSCaCertPath string `yaml:"tls_ca_cert_path"`
}

// AuditConfig defines configuration parameters for audit logging.
//
// AuditConfig 定义了审计日志记录的配置参数。
type AuditConfig struct {
	// Enabled indicates whether audit logging is enabled.
	// 指示是否启用审计日志记录。
	Enabled bool `yaml:"enabled"`
	// FilePath is the path to the audit log file.
	// 审计日志文件的路径。
	FilePath string `yaml:"file_path"`
	// LogQueries indicates whether to log executed SQL queries.
	// 指示是否记录执行的 SQL 查询。
	LogQueries bool `yaml:"log_queries"`
	// LogConnections indicates whether to log client connection events.
	// 指示是否记录客户端连接事件。
	LogConnections bool `yaml:"log_connections"`
	// LogAuthentication indicates whether to log authentication attempts.
	// 指示是否记录认证尝试。
	LogAuthentication bool `yaml:"log_authentication"`
}

var (
	// globalConfig is the singleton instance of the application configuration.
	// globalConfig 是应用程序配置的单例实例。
	globalConfig *Config
	// configOnce ensures that the globalConfig is initialized only once.
	// configOnce 确保 globalConfig 只初始化一次。
	configOnce sync.Once
)

// GetConfig returns the singleton instance of the application configuration.
// It panics if the configuration has not been loaded successfully via LoadConfig.
//
// GetConfig 返回应用程序配置的单例实例。
// 如果尚未通过 LoadConfig 成功加载配置，它会 panic。
func GetConfig() *Config {
	if globalConfig == nil {
		// This indicates a programming error where GetConfig is called before LoadConfig.
		log.GetLogger().Fatal("Configuration not loaded. Call LoadConfig() first.")
	}
	return globalConfig
}

// LoadConfig loads the configuration from the specified YAML file path.
// It applies default values where not specified and initializes the global config instance.
// If filePath is empty, it attempts to load from constants.DefaultConfigPath.
//
// LoadConfig 从指定的 YAML 文件路径加载配置。
// 它在未指定的地方应用默认值，并初始化全局配置实例。
// 如果 filePath 为空，它会尝试从 constants.DefaultConfigPath 加载。
func LoadConfig(filePath string) error {
	var loadErr error
	configOnce.Do(func() {
		if filePath == "" {
			filePath = constants.DefaultConfigPath
		}

		log.GetLogger().Info("Attempting to load configuration from: %s", filePath)

		data, err := os.ReadFile(filePath)
		if err != nil {
			// If file not found, try to create with defaults
			if os.IsNotExist(err) {
				log.GetLogger().Warn("Config file not found at %s. Creating with default values.", filePath)
				defaultCfg := GetDefaultConfig()
				data, marshalErr := yaml.Marshal(defaultCfg)
				if marshalErr != nil {
					loadErr = fmt.Errorf("failed to marshal default config: %w", marshalErr)
					return
				}
				writeErr := os.WriteFile(filePath, data, 0644)
				if writeErr != nil {
					loadErr = fmt.Errorf("failed to write default config to %s: %w", filePath, writeErr)
					return
				}
				globalConfig = defaultCfg
				log.GetLogger().Info("Default configuration saved to %s. Please review and modify as needed.", filePath)
				return
			}
			loadErr = fmt.Errorf("failed to read config file %s: %w", filePath, err)
			return
		}

		cfg := &Config{}
		// Set sensible defaults before unmarshalling to ensure all fields have a value
		// in case they are omitted from the YAML file.
		*cfg = *GetDefaultConfig() // Apply defaults first

		if err := yaml.Unmarshal(data, cfg); err != nil {
			loadErr = fmt.Errorf("failed to unmarshal config file %s: %w", filePath, err)
			return
		}

		globalConfig = cfg
		log.GetLogger().Info("Configuration loaded successfully from: %s", filePath)
	})
	return loadErr
}

// GetDefaultConfig returns a Config struct populated with sensible default values.
// This is used if no configuration file is found or if certain fields are omitted.
//
// GetDefaultConfig 返回一个填充了合理默认值的 Config 结构体。
// 当找不到配置文件或省略某些字段时使用。
func GetDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:           constants.DefaultServerPort,
			Host:           "0.0.0.0", // Listen on all interfaces by default
			MaxConnections: constants.DefaultMaxConnections,
			ReadTimeout:    30 * time.Second, // 30 seconds
			WriteTimeout:   30 * time.Second, // 30 seconds
			Charset:        constants.MySQLCharsetUTF8mb4,
			Collation:      constants.MySQLCollationUTF8mb4Bin,
		},
		Storage: StorageConfig{
			Engine:  enum.StorageEngineType_Badger,
			DataDir: constants.DefaultDataDirPath,
			Badger: BadgerConfig{
				ValueLogFileSize:   constants.DefaultBadgerValueLogFileSize,
				SyncWrites:         constants.DefaultBadgerSyncWrites,
				VLogGCInterval:     5 * time.Minute, // Every 5 minutes
				VLogGCDiscardRatio: 0.5,             // Discard ratio for GC
			},
			KVD: KVDConfig{ // Placeholder defaults
				Path: "./data/kvd",
			},
		},
		Log: LogConfig{
			Level:    enum.LogLevel_INFO,
			FilePath: constants.DefaultLogFilePath,
		},
		Security: SecurityConfig{
			DefaultUser:     constants.DefaultUser,
			DefaultPassword: constants.DefaultPassword,
			AuthMethod:      enum.AuthMethod_CachingSha2Password, // Modern MySQL default
			TLSEnabled:      false,
			TLSCertPath:     "",
			TLSKeyPath:      "",
			TLSCaCertPath:   "",
		},
		Audit: AuditConfig{
			Enabled:           true, // Audit logging enabled by default
			FilePath:          constants.DefaultAuditLogPath,
			LogQueries:        true,
			LogConnections:    true,
			LogAuthentication: true,
		},
	}
}

// ValidateConfig performs basic validation on the loaded configuration.
// It checks for sensible values and potential inconsistencies.
// Returns an error if validation fails.
//
// ValidateConfig 对加载的配置执行基本验证。
// 它检查合理的值和潜在的不一致性。
// 如果验证失败，则返回错误。
func (c *Config) ValidateConfig() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", c.Server.Port)
	}
	if c.Server.MaxConnections <= 0 {
		return fmt.Errorf("max_connections must be a positive integer, got %d", c.Server.MaxConnections)
	}
	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be a positive duration")
	}
	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be a positive duration")
	}

	if c.Storage.DataDir == "" {
		return fmt.Errorf("storage data_dir cannot be empty")
	}
	if _, err := enum.ParseStorageEngineType(c.Storage.Engine.String()); err != nil {
		return fmt.Errorf("invalid storage engine type: %w", err)
	}

	// Specific Badger validation
	if c.Storage.Engine == enum.StorageEngineType_Badger {
		if c.Storage.Badger.ValueLogFileSize <= 0 {
			return fmt.Errorf("badger.value_log_file_size must be a positive integer")
		}
		if c.Storage.Badger.VLogGCInterval <= 0 {
			return fmt.Errorf("badger.vlog_gc_interval must be a positive duration")
		}
		if c.Storage.Badger.VLogGCDiscardRatio < 0 || c.Storage.Badger.VLogGCDiscardRatio > 1 {
			return fmt.Errorf("badger.vlog_gc_discard_ratio must be between 0 and 1")
		}
	}

	if _, err := enum.ParseLogLevel(c.Log.Level.String()); err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	if c.Security.TLSEnabled {
		if c.Security.TLSCertPath == "" || c.Security.TLSKeyPath == "" {
			return fmt.Errorf("tls_cert_path and tls_key_path must be specified when TLS is enabled")
		}
		// Further validation for existence of files could be added here
	}

	if c.Audit.Enabled && c.Audit.FilePath == "" {
		return fmt.Errorf("audit.file_path cannot be empty when audit logging is enabled")
	}

	// Add more validation rules as the config grows
	return nil
}
