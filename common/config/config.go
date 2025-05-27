// Package config 实现GuoceDB的配置管理系统
// Package config implements configuration management system for GuoceDB
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/turtacn/guocedb/common/constants"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
)

// DatabaseConfig 数据库配置
// DatabaseConfig database configuration
type DatabaseConfig struct {
	Name           string            `yaml:"name" json:"name"`                       // 数据库名称 Database name
	DataDir        string            `yaml:"data_dir" json:"data_dir"`               // 数据目录 Data directory
	WALDir         string            `yaml:"wal_dir" json:"wal_dir"`                 // WAL目录 WAL directory
	TempDir        string            `yaml:"temp_dir" json:"temp_dir"`               // 临时目录 Temporary directory
	MaxConnections int               `yaml:"max_connections" json:"max_connections"` // 最大连接数 Maximum connections
	DefaultCharset string            `yaml:"default_charset" json:"default_charset"` // 默认字符集 Default charset
	DefaultCollate string            `yaml:"default_collate" json:"default_collate"` // 默认排序规则 Default collation
	TimeZone       string            `yaml:"timezone" json:"timezone"`               // 时区 Time zone
	ReadOnly       bool              `yaml:"read_only" json:"read_only"`             // 只读模式 Read-only mode
	Metadata       map[string]string `yaml:"metadata" json:"metadata"`               // 元数据 Metadata
}

// StorageConfig 存储配置
// StorageConfig storage configuration
type StorageConfig struct {
	Engine             string                 `yaml:"engine" json:"engine"`                           // 存储引擎 Storage engine
	PageSize           int                    `yaml:"page_size" json:"page_size"`                     // 页面大小 Page size
	CacheSize          int64                  `yaml:"cache_size" json:"cache_size"`                   // 缓存大小 Cache size
	BufferPoolSize     int64                  `yaml:"buffer_pool_size" json:"buffer_pool_size"`       // 缓冲池大小 Buffer pool size
	WALBufferSize      int64                  `yaml:"wal_buffer_size" json:"wal_buffer_size"`         // WAL缓冲区大小 WAL buffer size
	MaxLogFileSize     int64                  `yaml:"max_log_file_size" json:"max_log_file_size"`     // 最大日志文件大小 Maximum log file size
	CheckpointInterval time.Duration          `yaml:"checkpoint_interval" json:"checkpoint_interval"` // 检查点间隔 Checkpoint interval
	FlushInterval      time.Duration          `yaml:"flush_interval" json:"flush_interval"`           // 刷新间隔 Flush interval
	SyncWrites         bool                   `yaml:"sync_writes" json:"sync_writes"`                 // 同步写入 Synchronous writes
	Compression        string                 `yaml:"compression" json:"compression"`                 // 压缩算法 Compression algorithm
	Encryption         *EncryptionConfig      `yaml:"encryption" json:"encryption"`                   // 加密配置 Encryption configuration
	Options            map[string]interface{} `yaml:"options" json:"options"`                         // 引擎特定选项 Engine-specific options
}

// EncryptionConfig 加密配置
// EncryptionConfig encryption configuration
type EncryptionConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`     // 是否启用加密 Whether encryption is enabled
	Algorithm string `yaml:"algorithm" json:"algorithm"` // 加密算法 Encryption algorithm
	KeyFile   string `yaml:"key_file" json:"key_file"`   // 密钥文件路径 Key file path
	KeySize   int    `yaml:"key_size" json:"key_size"`   // 密钥大小 Key size
}

// NetworkConfig 网络配置
// NetworkConfig network configuration
type NetworkConfig struct {
	Host           string        `yaml:"host" json:"host"`                       // 监听地址 Listen address
	Port           int           `yaml:"port" json:"port"`                       // 监听端口 Listen port
	MaxConnections int           `yaml:"max_connections" json:"max_connections"` // 最大连接数 Maximum connections
	ReadTimeout    time.Duration `yaml:"read_timeout" json:"read_timeout"`       // 读取超时 Read timeout
	WriteTimeout   time.Duration `yaml:"write_timeout" json:"write_timeout"`     // 写入超时 Write timeout
	IdleTimeout    time.Duration `yaml:"idle_timeout" json:"idle_timeout"`       // 空闲超时 Idle timeout
	KeepAlive      bool          `yaml:"keep_alive" json:"keep_alive"`           // 保持连接 Keep alive
	TLS            *TLSConfig    `yaml:"tls" json:"tls"`                         // TLS配置 TLS configuration
}

// TLSConfig TLS配置
// TLSConfig TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`     // 是否启用TLS Whether TLS is enabled
	CertFile string `yaml:"cert_file" json:"cert_file"` // 证书文件路径 Certificate file path
	KeyFile  string `yaml:"key_file" json:"key_file"`   // 私钥文件路径 Private key file path
	CAFile   string `yaml:"ca_file" json:"ca_file"`     // CA文件路径 CA file path
}

// SecurityConfig 安全配置
// SecurityConfig security configuration
type SecurityConfig struct {
	Authentication *AuthConfig      `yaml:"authentication" json:"authentication"` // 认证配置 Authentication configuration
	Authorization  *AuthzConfig     `yaml:"authorization" json:"authorization"`   // 授权配置 Authorization configuration
	Audit          *AuditConfig     `yaml:"audit" json:"audit"`                   // 审计配置 Audit configuration
	RateLimit      *RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`         // 限流配置 Rate limit configuration
	IPWhitelist    []string         `yaml:"ip_whitelist" json:"ip_whitelist"`     // IP白名单 IP whitelist
	IPBlacklist    []string         `yaml:"ip_blacklist" json:"ip_blacklist"`     // IP黑名单 IP blacklist
}

// AuthConfig 认证配置
// AuthConfig authentication configuration
type AuthConfig struct {
	Enabled     bool          `yaml:"enabled" json:"enabled"`           // 是否启用认证 Whether authentication is enabled
	Method      string        `yaml:"method" json:"method"`             // 认证方法 Authentication method
	TokenExpiry time.Duration `yaml:"token_expiry" json:"token_expiry"` // 令牌过期时间 Token expiry time
	SecretKey   string        `yaml:"secret_key" json:"secret_key"`     // 密钥 Secret key
	LDAP        *LDAPConfig   `yaml:"ldap" json:"ldap"`                 // LDAP配置 LDAP configuration
}

// LDAPConfig LDAP配置
// LDAPConfig LDAP configuration
type LDAPConfig struct {
	Server   string `yaml:"server" json:"server"`       // LDAP服务器 LDAP server
	Port     int    `yaml:"port" json:"port"`           // LDAP端口 LDAP port
	BaseDN   string `yaml:"base_dn" json:"base_dn"`     // 基础DN Base DN
	BindDN   string `yaml:"bind_dn" json:"bind_dn"`     // 绑定DN Bind DN
	BindPW   string `yaml:"bind_pw" json:"bind_pw"`     // 绑定密码 Bind password
	UserAttr string `yaml:"user_attr" json:"user_attr"` // 用户属性 User attribute
}

// AuthzConfig 授权配置
// AuthzConfig authorization configuration
type AuthzConfig struct {
	Enabled     bool                `yaml:"enabled" json:"enabled"`           // 是否启用授权 Whether authorization is enabled
	Model       string              `yaml:"model" json:"model"`               // 授权模型 Authorization model
	Policy      string              `yaml:"policy" json:"policy"`             // 策略文件 Policy file
	Roles       map[string][]string `yaml:"roles" json:"roles"`               // 角色权限映射 Role permission mapping
	DefaultRole string              `yaml:"default_role" json:"default_role"` // 默认角色 Default role
}

// AuditConfig 审计配置
// AuditConfig audit configuration
type AuditConfig struct {
	Enabled    bool     `yaml:"enabled" json:"enabled"`         // 是否启用审计 Whether audit is enabled
	LogFile    string   `yaml:"log_file" json:"log_file"`       // 审计日志文件 Audit log file
	LogLevel   string   `yaml:"log_level" json:"log_level"`     // 审计日志级别 Audit log level
	Events     []string `yaml:"events" json:"events"`           // 审计事件类型 Audit event types
	MaxSize    int64    `yaml:"max_size" json:"max_size"`       // 最大日志大小 Maximum log size
	MaxBackups int      `yaml:"max_backups" json:"max_backups"` // 最大备份数量 Maximum backup count
}

// RateLimitConfig 限流配置
// RateLimitConfig rate limit configuration
type RateLimitConfig struct {
	Enabled           bool          `yaml:"enabled" json:"enabled"`                         // 是否启用限流 Whether rate limit is enabled
	RequestsPerSecond int           `yaml:"requests_per_second" json:"requests_per_second"` // 每秒请求数 Requests per second
	BurstSize         int           `yaml:"burst_size" json:"burst_size"`                   // 突发大小 Burst size
	WindowSize        time.Duration `yaml:"window_size" json:"window_size"`                 // 时间窗口大小 Window size
	WhitelistIPs      []string      `yaml:"whitelist_ips" json:"whitelist_ips"`             // 白名单IP Whitelist IPs
}

// LogConfig 日志配置
// LogConfig log configuration
type LogConfig struct {
	Level      string            `yaml:"level" json:"level"`             // 日志级别 Log level
	Format     string            `yaml:"format" json:"format"`           // 日志格式 Log format
	Output     string            `yaml:"output" json:"output"`           // 输出目标 Output target
	File       string            `yaml:"file" json:"file"`               // 日志文件路径 Log file path
	MaxSize    int               `yaml:"max_size" json:"max_size"`       // 最大文件大小 Maximum file size
	MaxBackups int               `yaml:"max_backups" json:"max_backups"` // 最大备份数量 Maximum backup count
	MaxAge     int               `yaml:"max_age" json:"max_age"`         // 最大保留天数 Maximum retention days
	Compress   bool              `yaml:"compress" json:"compress"`       // 是否压缩 Whether to compress
	Async      bool              `yaml:"async" json:"async"`             // 是否异步 Whether async
	Fields     map[string]string `yaml:"fields" json:"fields"`           // 附加字段 Additional fields
}

// MetricsConfig 监控配置
// MetricsConfig metrics configuration
type MetricsConfig struct {
	Enabled   bool              `yaml:"enabled" json:"enabled"`     // 是否启用监控 Whether metrics is enabled
	Host      string            `yaml:"host" json:"host"`           // 监控服务地址 Metrics service address
	Port      int               `yaml:"port" json:"port"`           // 监控服务端口 Metrics service port
	Path      string            `yaml:"path" json:"path"`           // 监控路径 Metrics path
	Interval  time.Duration     `yaml:"interval" json:"interval"`   // 收集间隔 Collection interval
	Retention time.Duration     `yaml:"retention" json:"retention"` // 数据保留时间 Data retention time
	Labels    map[string]string `yaml:"labels" json:"labels"`       // 标签 Labels
}

// ClusterConfig 集群配置
// ClusterConfig cluster configuration
type ClusterConfig struct {
	Enabled           bool              `yaml:"enabled" json:"enabled"`                       // 是否启用集群 Whether cluster is enabled
	NodeID            string            `yaml:"node_id" json:"node_id"`                       // 节点ID Node ID
	Seeds             []string          `yaml:"seeds" json:"seeds"`                           // 种子节点 Seed nodes
	Port              int               `yaml:"port" json:"port"`                             // 集群端口 Cluster port
	HeartbeatInterval time.Duration     `yaml:"heartbeat_interval" json:"heartbeat_interval"` // 心跳间隔 Heartbeat interval
	ElectionTimeout   time.Duration     `yaml:"election_timeout" json:"election_timeout"`     // 选举超时 Election timeout
	ReplicaFactor     int               `yaml:"replica_factor" json:"replica_factor"`         // 副本因子 Replica factor
	Consistency       string            `yaml:"consistency" json:"consistency"`               // 一致性级别 Consistency level
	Metadata          map[string]string `yaml:"metadata" json:"metadata"`                     // 节点元数据 Node metadata
}

// Config 主配置结构
// Config main configuration structure
type Config struct {
	Database *DatabaseConfig `yaml:"database" json:"database"` // 数据库配置 Database configuration
	Storage  *StorageConfig  `yaml:"storage" json:"storage"`   // 存储配置 Storage configuration
	Network  *NetworkConfig  `yaml:"network" json:"network"`   // 网络配置 Network configuration
	Security *SecurityConfig `yaml:"security" json:"security"` // 安全配置 Security configuration
	Log      *LogConfig      `yaml:"log" json:"log"`           // 日志配置 Log configuration
	Metrics  *MetricsConfig  `yaml:"metrics" json:"metrics"`   // 监控配置 Metrics configuration
	Cluster  *ClusterConfig  `yaml:"cluster" json:"cluster"`   // 集群配置 Cluster configuration
}

// DefaultConfig 默认配置
// DefaultConfig default configuration
func DefaultConfig() *Config {
	return &Config{
		Database: &DatabaseConfig{
			Name:           constants.DefaultDatabaseName,
			DataDir:        constants.DefaultDataDir,
			WALDir:         constants.DefaultWALDir,
			TempDir:        constants.DefaultTempDir,
			MaxConnections: constants.DefaultMaxConnections,
			DefaultCharset: constants.DefaultCharset,
			DefaultCollate: constants.DefaultCollate,
			TimeZone:       constants.DefaultTimeZone,
			ReadOnly:       false,
			Metadata:       make(map[string]string),
		},
		Storage: &StorageConfig{
			Engine:             constants.DefaultStorageEngine,
			PageSize:           constants.DefaultPageSize,
			CacheSize:          constants.DefaultCacheSize,
			BufferPoolSize:     constants.DefaultBufferPoolSize,
			WALBufferSize:      constants.DefaultWALBufferSize,
			MaxLogFileSize:     constants.DefaultMaxLogFileSize,
			CheckpointInterval: constants.DefaultCheckpointInterval,
			FlushInterval:      constants.DefaultFlushInterval,
			SyncWrites:         constants.DefaultSyncWrites,
			Compression:        constants.DefaultCompression,
			Encryption: &EncryptionConfig{
				Enabled:   false,
				Algorithm: "AES-256-GCM",
				KeySize:   256,
			},
			Options: make(map[string]interface{}),
		},
		Network: &NetworkConfig{
			Host:           constants.DefaultHost,
			Port:           constants.DefaultPort,
			MaxConnections: constants.DefaultMaxConnections,
			ReadTimeout:    constants.DefaultReadTimeout,
			WriteTimeout:   constants.DefaultWriteTimeout,
			IdleTimeout:    constants.DefaultIdleTimeout,
			KeepAlive:      true,
			TLS: &TLSConfig{
				Enabled: false,
			},
		},
		Security: &SecurityConfig{
			Authentication: &AuthConfig{
				Enabled:     false,
				Method:      "password",
				TokenExpiry: 24 * time.Hour,
			},
			Authorization: &AuthzConfig{
				Enabled:     false,
				Model:       "rbac",
				DefaultRole: "guest",
				Roles:       make(map[string][]string),
			},
			Audit: &AuditConfig{
				Enabled:    false,
				LogLevel:   "info",
				Events:     []string{"login", "logout", "query", "update"},
				MaxSize:    100 * 1024 * 1024, // 100MB
				MaxBackups: 5,
			},
			RateLimit: &RateLimitConfig{
				Enabled:           false,
				RequestsPerSecond: 1000,
				BurstSize:         100,
				WindowSize:        time.Minute,
				WhitelistIPs:      []string{},
			},
			IPWhitelist: []string{},
			IPBlacklist: []string{},
		},
		Log: &LogConfig{
			Level:      constants.DefaultLogLevel,
			Format:     constants.DefaultLogFormat,
			Output:     "file",
			File:       constants.DefaultLogFile,
			MaxSize:    constants.DefaultLogMaxSize,
			MaxBackups: constants.DefaultLogMaxBackups,
			MaxAge:     constants.DefaultLogMaxAge,
			Compress:   constants.DefaultLogCompress,
			Async:      true,
			Fields:     make(map[string]string),
		},
		Metrics: &MetricsConfig{
			Enabled:   false,
			Host:      "0.0.0.0",
			Port:      9090,
			Path:      "/metrics",
			Interval:  30 * time.Second,
			Retention: 24 * time.Hour,
			Labels:    make(map[string]string),
		},
		Cluster: &ClusterConfig{
			Enabled:           false,
			Port:              7000,
			HeartbeatInterval: 5 * time.Second,
			ElectionTimeout:   10 * time.Second,
			ReplicaFactor:     3,
			Consistency:       "quorum",
			Metadata:          make(map[string]string),
		},
	}
}

// LoadFromFile 从文件加载配置
// LoadFromFile loads configuration from file
func LoadFromFile(filename string) (*Config, error) {
	if filename == "" {
		return nil, errors.NewError(errors.ErrCodeInvalidParameter, "Configuration file path is empty")
	}

	// 检查文件是否存在
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, errors.NewErrorf(errors.ErrCodeFileNotFound, "Configuration file not found: %s", filename)
	}

	// 读取文件内容
	// Read file content
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.WrapErrorf(errors.ErrCodeFileReadError, "Failed to read configuration file: %s", err, filename)
	}

	// 获取文件扩展名
	// Get file extension
	ext := strings.ToLower(filepath.Ext(filename))

	// 创建默认配置
	// Create default configuration
	config := DefaultConfig()

	// 根据文件类型解析配置
	// Parse configuration based on file type
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, errors.WrapErrorf(errors.ErrCodeInvalidFormat, "Failed to parse YAML configuration: %s", err, filename)
		}
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return nil, errors.WrapErrorf(errors.ErrCodeInvalidFormat, "Failed to parse JSON configuration: %s", err, filename)
		}
	default:
		return nil, errors.NewErrorf(errors.ErrCodeInvalidFormat, "Unsupported configuration file format: %s", ext)
	}

	// 应用环境变量覆盖
	// Apply environment variable overrides
	if err := applyEnvOverrides(config); err != nil {
		return nil, errors.WrapError(errors.ErrCodeInvalidFormat, "Failed to apply environment variable overrides", err)
	}

	// 验证配置
	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return nil, errors.WrapError(errors.ErrCodeInvalidParameter, "Configuration validation failed", err)
	}

	return config, nil
}

// SaveToFile 保存配置到文件
// SaveToFile saves configuration to file
func SaveToFile(config *Config, filename string) error {
	if config == nil {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Configuration is nil")
	}

	if filename == "" {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Configuration file path is empty")
	}

	// 确保目录存在
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.WrapErrorf(errors.ErrCodeSystemFailure, "Failed to create configuration directory: %s", err, dir)
	}

	// 获取文件扩展名
	// Get file extension
	ext := strings.ToLower(filepath.Ext(filename))

	var data []byte
	var err error

	// 根据文件类型序列化配置
	// Serialize configuration based on file type
	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
		if err != nil {
			return errors.WrapError(errors.ErrCodeInvalidFormat, "Failed to marshal configuration to YAML", err)
		}
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return errors.WrapError(errors.ErrCodeInvalidFormat, "Failed to marshal configuration to JSON", err)
		}
	default:
		return errors.NewErrorf(errors.ErrCodeInvalidFormat, "Unsupported configuration file format: %s", ext)
	}

	// 写入文件
	// Write to file
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return errors.WrapErrorf(errors.ErrCodeFileWriteError, "Failed to write configuration file: %s", err, filename)
	}

	return nil
}

// applyEnvOverrides 应用环境变量覆盖
// applyEnvOverrides applies environment variable overrides
func applyEnvOverrides(config *Config) error {
	// 使用反射遍历配置结构
	// Use reflection to traverse configuration structure
	return applyEnvOverridesToStruct(reflect.ValueOf(config).Elem(), "GUOCEDB")
}

// applyEnvOverridesToStruct 递归应用环境变量覆盖到结构体
// applyEnvOverridesToStruct recursively applies environment variable overrides to struct
func applyEnvOverridesToStruct(v reflect.Value, prefix string) error {
	if !v.IsValid() || !v.CanSet() {
		return nil
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 跳过私有字段
		// Skip private fields
		if !field.CanSet() {
			continue
		}

		// 构建环境变量名
		// Build environment variable name
		envName := prefix + "_" + strings.ToUpper(fieldType.Name)

		// 获取yaml标签作为字段名
		// Get yaml tag as field name
		if yamlTag := fieldType.Tag.Get("yaml"); yamlTag != "" {
			tagParts := strings.Split(yamlTag, ",")
			if len(tagParts) > 0 && tagParts[0] != "" {
				envName = prefix + "_" + strings.ToUpper(strings.ReplaceAll(tagParts[0], "_", "_"))
			}
		}

		switch field.Kind() {
		case reflect.String:
			if value := os.Getenv(envName); value != "" {
				field.SetString(value)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if value := os.Getenv(envName); value != "" {
				if field.Type() == reflect.TypeOf(time.Duration(0)) {
					// 处理Duration类型
					// Handle Duration type
					if duration, err := time.ParseDuration(value); err == nil {
						field.Set(reflect.ValueOf(duration))
					}
				} else {
					if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
						field.SetInt(intValue)
					}
				}
			}
		case reflect.Bool:
			if value := os.Getenv(envName); value != "" {
				if boolValue, err := strconv.ParseBool(value); err == nil {
					field.SetBool(boolValue)
				}
			}
		case reflect.Slice:
			if value := os.Getenv(envName); value != "" {
				// 处理字符串切片
				// Handle string slice
				if field.Type().Elem().Kind() == reflect.String {
					values := strings.Split(value, ",")
					slice := reflect.MakeSlice(field.Type(), len(values), len(values))
					for j, v := range values {
						slice.Index(j).SetString(strings.TrimSpace(v))
					}
					field.Set(slice)
				}
			}
		case reflect.Map:
			if value := os.Getenv(envName); value != "" {
				// 处理字符串映射 key1=value1,key2=value2
				// Handle string map key1=value1,key2=value2
				if field.Type().Key().Kind() == reflect.String && field.Type().Elem().Kind() == reflect.String {
					pairs := strings.Split(value, ",")
					mapValue := reflect.MakeMap(field.Type())
					for _, pair := range pairs {
						kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
						if len(kv) == 2 {
							mapValue.SetMapIndex(reflect.ValueOf(kv[0]), reflect.ValueOf(kv[1]))
						}
					}
					field.Set(mapValue)
				}
			}
		case reflect.Ptr:
			if field.IsNil() {
				// 创建新的指针实例
				// Create new pointer instance
				field.Set(reflect.New(field.Type().Elem()))
			}
			if err := applyEnvOverridesToStruct(field.Elem(), envName); err != nil {
				return err
			}
		case reflect.Struct:
			if err := applyEnvOverridesToStruct(field, envName); err != nil {
				return err
			}
		}
	}

	return nil
}

// ValidateConfig 验证配置
// ValidateConfig validates configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Configuration is nil")
	}

	// 验证数据库配置
	// Validate database configuration
	if err := validateDatabaseConfig(config.Database); err != nil {
		return errors.WrapError(errors.ErrCodeInvalidParameter, "Invalid database configuration", err)
	}

	// 验证存储配置
	// Validate storage configuration
	if err := validateStorageConfig(config.Storage); err != nil {
		return errors.WrapError(errors.ErrCodeInvalidParameter, "Invalid storage configuration", err)
	}

	// 验证网络配置
	// Validate network configuration
	if err := validateNetworkConfig(config.Network); err != nil {
		return errors.WrapError(errors.ErrCodeInvalidParameter, "Invalid network configuration", err)
	}

	// 验证安全配置
	// Validate security configuration
	if err := validateSecurityConfig(config.Security); err != nil {
		return errors.WrapError(errors.ErrCodeInvalidParameter, "Invalid security configuration", err)
	}

	// 验证日志配置
	// Validate log configuration
	if err := validateLogConfig(config.Log); err != nil {
		return errors.WrapError(errors.ErrCodeInvalidParameter, "Invalid log configuration", err)
	}

	// 验证监控配置
	// Validate metrics configuration
	if err := validateMetricsConfig(config.Metrics); err != nil {
		return errors.WrapError(errors.ErrCodeInvalidParameter, "Invalid metrics configuration", err)
	}

	// 验证集群配置
	// Validate cluster configuration
	if err := validateClusterConfig(config.Cluster); err != nil {
		return errors.WrapError(errors.ErrCodeInvalidParameter, "Invalid cluster configuration", err)
	}

	return nil
}

// validateDatabaseConfig 验证数据库配置
// validateDatabaseConfig validates database configuration
func validateDatabaseConfig(config *DatabaseConfig) error {
	if config == nil {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Database configuration is nil")
	}

	if config.Name == "" {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Database name cannot be empty")
	}

	if config.DataDir == "" {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Data directory cannot be empty")
	}

	if config.MaxConnections <= 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Max connections must be positive")
	}

	if config.DefaultCharset == "" {
		config.DefaultCharset = constants.DefaultCharset
	}

	if config.DefaultCollate == "" {
		config.DefaultCollate = constants.DefaultCollate
	}

	if config.TimeZone == "" {
		config.TimeZone = constants.DefaultTimeZone
	}

	return nil
}

// validateStorageConfig 验证存储配置
// validateStorageConfig validates storage configuration
func validateStorageConfig(config *StorageConfig) error {
	if config == nil {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Storage configuration is nil")
	}

	if config.Engine == "" {
		config.Engine = constants.DefaultStorageEngine
	}

	if config.PageSize <= 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Page size must be positive")
	}

	if config.CacheSize < 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Cache size cannot be negative")
	}

	if config.BufferPoolSize < 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Buffer pool size cannot be negative")
	}

	if config.WALBufferSize < 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "WAL buffer size cannot be negative")
	}

	if config.MaxLogFileSize <= 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Max log file size must be positive")
	}

	if config.CheckpointInterval <= 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Checkpoint interval must be positive")
	}

	if config.FlushInterval <= 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Flush interval must be positive")
	}

	// 验证压缩算法
	// Validate compression algorithm
	validCompressions := []string{"none", "gzip", "lz4", "snappy", "zstd"}
	if config.Compression != "" {
		found := false
		for _, valid := range validCompressions {
			if config.Compression == valid {
				found = true
				break
			}
		}
		if !found {
			return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid compression algorithm: %s", config.Compression)
		}
	}

	// 验证加密配置
	// Validate encryption configuration
	if config.Encryption != nil && config.Encryption.Enabled {
		if config.Encryption.Algorithm == "" {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Encryption algorithm cannot be empty when encryption is enabled")
		}

		if config.Encryption.KeySize <= 0 {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Encryption key size must be positive")
		}

		validAlgorithms := []string{"AES-128-GCM", "AES-192-GCM", "AES-256-GCM", "ChaCha20-Poly1305"}
		found := false
		for _, valid := range validAlgorithms {
			if config.Encryption.Algorithm == valid {
				found = true
				break
			}
		}
		if !found {
			return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid encryption algorithm: %s", config.Encryption.Algorithm)
		}
	}

	return nil
}

// validateNetworkConfig 验证网络配置
// validateNetworkConfig validates network configuration
func validateNetworkConfig(config *NetworkConfig) error {
	if config == nil {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Network configuration is nil")
	}

	if config.Host == "" {
		config.Host = constants.DefaultHost
	}

	if config.Port <= 0 || config.Port > 65535 {
		return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid port number: %d", config.Port)
	}

	if config.MaxConnections <= 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Max connections must be positive")
	}

	if config.ReadTimeout <= 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Read timeout must be positive")
	}

	if config.WriteTimeout <= 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Write timeout must be positive")
	}

	if config.IdleTimeout <= 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Idle timeout must be positive")
	}

	// 验证TLS配置
	// Validate TLS configuration
	if config.TLS != nil && config.TLS.Enabled {
		if config.TLS.CertFile == "" {
			return errors.NewError(errors.ErrCodeInvalidParameter, "TLS certificate file cannot be empty when TLS is enabled")
		}

		if config.TLS.KeyFile == "" {
			return errors.NewError(errors.ErrCodeInvalidParameter, "TLS key file cannot be empty when TLS is enabled")
		}

		// 检查证书文件是否存在
		// Check if certificate files exist
		if _, err := os.Stat(config.TLS.CertFile); os.IsNotExist(err) {
			return errors.NewErrorf(errors.ErrCodeFileNotFound, "TLS certificate file not found: %s", config.TLS.CertFile)
		}

		if _, err := os.Stat(config.TLS.KeyFile); os.IsNotExist(err) {
			return errors.NewErrorf(errors.ErrCodeFileNotFound, "TLS key file not found: %s", config.TLS.KeyFile)
		}
	}

	return nil
}

// validateSecurityConfig 验证安全配置
// validateSecurityConfig validates security configuration
func validateSecurityConfig(config *SecurityConfig) error {
	if config == nil {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Security configuration is nil")
	}

	// 验证认证配置
	// Validate authentication configuration
	if config.Authentication != nil && config.Authentication.Enabled {
		validMethods := []string{"password", "token", "ldap", "oauth"}
		found := false
		for _, valid := range validMethods {
			if config.Authentication.Method == valid {
				found = true
				break
			}
		}
		if !found {
			return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid authentication method: %s", config.Authentication.Method)
		}

		if config.Authentication.TokenExpiry <= 0 {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Token expiry must be positive")
		}

		// 验证LDAP配置
		// Validate LDAP configuration
		if config.Authentication.Method == "ldap" && config.Authentication.LDAP != nil {
			if config.Authentication.LDAP.Server == "" {
				return errors.NewError(errors.ErrCodeInvalidParameter, "LDAP server cannot be empty")
			}

			if config.Authentication.LDAP.Port <= 0 || config.Authentication.LDAP.Port > 65535 {
				return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid LDAP port: %d", config.Authentication.LDAP.Port)
			}
		}
	}

	// 验证授权配置
	// Validate authorization configuration
	if config.Authorization != nil && config.Authorization.Enabled {
		validModels := []string{"rbac", "abac", "acl"}
		found := false
		for _, valid := range validModels {
			if config.Authorization.Model == valid {
				found = true
				break
			}
		}
		if !found {
			return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid authorization model: %s", config.Authorization.Model)
		}
	}

	// 验证限流配置
	// Validate rate limit configuration
	if config.RateLimit != nil && config.RateLimit.Enabled {
		if config.RateLimit.RequestsPerSecond <= 0 {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Requests per second must be positive")
		}

		if config.RateLimit.BurstSize <= 0 {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Burst size must be positive")
		}

		if config.RateLimit.WindowSize <= 0 {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Window size must be positive")
		}
	}

	return nil
}

// validateLogConfig 验证日志配置
// validateLogConfig validates log configuration
func validateLogConfig(config *LogConfig) error {
	if config == nil {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Log configuration is nil")
	}

	// 验证日志级别
	// Validate log level
	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal"}
	found := false
	for _, valid := range validLevels {
		if strings.ToLower(config.Level) == valid {
			found = true
			break
		}
	}
	if !found {
		return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid log level: %s", config.Level)
	}

	// 验证日志格式
	// Validate log format
	validFormats := []string{"text", "json"}
	found = false
	for _, valid := range validFormats {
		if strings.ToLower(config.Format) == valid {
			found = true
			break
		}
	}
	if !found {
		return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid log format: %s", config.Format)
	}

	// 验证输出目标
	// Validate output target
	validOutputs := []string{"stdout", "stderr", "file"}
	found = false
	for _, valid := range validOutputs {
		if strings.ToLower(config.Output) == valid {
			found = true
			break
		}
	}
	if !found {
		return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid log output: %s", config.Output)
	}

	// 如果输出到文件，验证文件路径
	// If output to file, validate file path
	if strings.ToLower(config.Output) == "file" && config.File == "" {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Log file path cannot be empty when output is file")
	}

	if config.MaxSize <= 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Log max size must be positive")
	}

	if config.MaxBackups < 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Log max backups cannot be negative")
	}

	if config.MaxAge < 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Log max age cannot be negative")
	}

	return nil
}

// validateMetricsConfig 验证监控配置
// validateMetricsConfig validates metrics configuration
func validateMetricsConfig(config *MetricsConfig) error {
	if config == nil {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Metrics configuration is nil")
	}

	if config.Enabled {
		if config.Host == "" {
			config.Host = "0.0.0.0"
		}

		if config.Port <= 0 || config.Port > 65535 {
			return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid metrics port: %d", config.Port)
		}

		if config.Path == "" {
			config.Path = "/metrics"
		}

		if config.Interval <= 0 {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Metrics interval must be positive")
		}

		if config.Retention <= 0 {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Metrics retention must be positive")
		}
	}

	return nil
}

// validateClusterConfig 验证集群配置
// validateClusterConfig validates cluster configuration
func validateClusterConfig(config *ClusterConfig) error {
	if config == nil {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Cluster configuration is nil")
	}

	if config.Enabled {
		if config.NodeID == "" {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Node ID cannot be empty when cluster is enabled")
		}

		if config.Port <= 0 || config.Port > 65535 {
			return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid cluster port: %d", config.Port)
		}

		if config.HeartbeatInterval <= 0 {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Heartbeat interval must be positive")
		}

		if config.ElectionTimeout <= 0 {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Election timeout must be positive")
		}

		if config.ReplicaFactor <= 0 {
			return errors.NewError(errors.ErrCodeInvalidParameter, "Replica factor must be positive")
		}

		// 验证一致性级别
		// Validate consistency level
		validConsistencies := []string{"eventual", "strong", "quorum"}
		found := false
		for _, valid := range validConsistencies {
			if config.Consistency == valid {
				found = true
				break
			}
		}
		if !found {
			return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Invalid consistency level: %s", config.Consistency)
		}
	}

	return nil
}

// Clone 克隆配置
// Clone clones configuration
func (c *Config) Clone() *Config {
	data, err := json.Marshal(c)
	if err != nil {
		return nil
	}

	var clone Config
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil
	}

	return &clone
}

// Merge 合并配置
// Merge merges configurations
func (c *Config) Merge(other *Config) error {
	if other == nil {
		return nil
	}

	// 使用反射合并配置
	// Use reflection to merge configurations
	return mergeStructs(reflect.ValueOf(c).Elem(), reflect.ValueOf(other).Elem())
}

// mergeStructs 递归合并结构体
// mergeStructs recursively merges structs
func mergeStructs(dst, src reflect.Value) error {
	if !dst.IsValid() || !src.IsValid() {
		return nil
	}

	if dst.Type() != src.Type() {
		return errors.NewError(errors.ErrCodeInvalidParameter, "Cannot merge different types")
	}

	switch dst.Kind() {
	case reflect.Ptr:
		if src.IsNil() {
			return nil
		}

		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}

		return mergeStructs(dst.Elem(), src.Elem())

	case reflect.Struct:
		for i := 0; i < dst.NumField(); i++ {
			dstField := dst.Field(i)
			srcField := src.Field(i)

			if !dstField.CanSet() {
				continue
			}

			if err := mergeStructs(dstField, srcField); err != nil {
				return err
			}
		}

	case reflect.Map:
		if src.IsNil() {
			return nil
		}

		if dst.IsNil() {
			dst.Set(reflect.MakeMap(dst.Type()))
		}

		for _, key := range src.MapKeys() {
			dst.SetMapIndex(key, src.MapIndex(key))
		}

	case reflect.Slice:
		if src.IsNil() {
			return nil
		}

		dst.Set(src)

	default:
		if !src.IsZero() {
			dst.Set(src)
		}
	}

	return nil
}

// ToJSON 转换为JSON字符串
// ToJSON converts to JSON string
func (c *Config) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", errors.WrapError(errors.ErrCodeInvalidFormat, "Failed to marshal configuration to JSON", err)
	}
	return string(data), nil
}

// ToYAML 转换为YAML字符串
// ToYAML converts to YAML string
func (c *Config) ToYAML() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", errors.WrapError(errors.ErrCodeInvalidFormat, "Failed to marshal configuration to YAML", err)
	}
	return string(data), nil
}

// GetDataDirs 获取所有数据目录
// GetDataDirs gets all data directories
func (c *Config) GetDataDirs() []string {
	dirs := []string{}

	if c.Database != nil {
		if c.Database.DataDir != "" {
			dirs = append(dirs, c.Database.DataDir)
		}
		if c.Database.WALDir != "" {
			dirs = append(dirs, c.Database.WALDir)
		}
		if c.Database.TempDir != "" {
			dirs = append(dirs, c.Database.TempDir)
		}
	}

	return dirs
}

// EnsureDirectories 确保所有必要的目录存在
// EnsureDirectories ensures all necessary directories exist
func (c *Config) EnsureDirectories() error {
	dirs := c.GetDataDirs()

	// 添加日志目录
	// Add log directory
	if c.Log != nil && c.Log.File != "" {
		dirs = append(dirs, filepath.Dir(c.Log.File))
	}

	// 添加审计日志目录
	// Add audit log directory
	if c.Security != nil && c.Security.Audit != nil && c.Security.Audit.LogFile != "" {
		dirs = append(dirs, filepath.Dir(c.Security.Audit.LogFile))
	}

	// 创建目录
	// Create directories
	for _, dir := range dirs {
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return errors.WrapErrorf(errors.ErrCodeSystemFailure, "Failed to create directory: %s", err, dir)
			}
		}
	}

	return nil
}

// GetConnectionString 获取连接字符串
// GetConnectionString gets connection string
func (c *Config) GetConnectionString() string {
	if c.Network == nil {
		return fmt.Sprintf("%s:%d", constants.DefaultHost, constants.DefaultPort)
	}

	host := c.Network.Host
	if host == "" {
		host = constants.DefaultHost
	}

	port := c.Network.Port
	if port == 0 {
		port = constants.DefaultPort
	}

	return fmt.Sprintf("%s:%d", host, port)
}

// IsSecure 检查是否启用了安全功能
// IsSecure checks if security features are enabled
func (c *Config) IsSecure() bool {
	if c.Security == nil {
		return false
	}

	// 检查认证
	// Check authentication
	if c.Security.Authentication != nil && c.Security.Authentication.Enabled {
		return true
	}

	// 检查TLS
	// Check TLS
	if c.Network != nil && c.Network.TLS != nil && c.Network.TLS.Enabled {
		return true
	}

	// 检查加密
	// Check encryption
	if c.Storage != nil && c.Storage.Encryption != nil && c.Storage.Encryption.Enabled {
		return true
	}

	return false
}

// GetLogLevel 获取日志级别
// GetLogLevel gets log level
func (c *Config) GetLogLevel() log.LogLevel {
	if c.Log == nil || c.Log.Level == "" {
		return log.ParseLogLevel(constants.DefaultLogLevel)
	}
	return log.ParseLogLevel(c.Log.Level)
}

// GetLogFormat 获取日志格式
// GetLogFormat gets log format
func (c *Config) GetLogFormat() log.LogFormat {
	if c.Log == nil || c.Log.Format == "" {
		return log.ParseLogFormat(constants.DefaultLogFormat)
	}
	return log.ParseLogFormat(c.Log.Format)
}

// 全局配置管理 Global configuration management
var (
	globalConfig *Config
	configMu     sync.RWMutex
)

// SetGlobalConfig 设置全局配置
// SetGlobalConfig sets global configuration
func SetGlobalConfig(config *Config) {
	configMu.Lock()
	defer configMu.Unlock()
	globalConfig = config
}

// GetGlobalConfig 获取全局配置
// GetGlobalConfig gets global configuration
func GetGlobalConfig() *Config {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalConfig == nil {
		return DefaultConfig()
	}

	return globalConfig
}

// LoadGlobalConfig 加载全局配置
// LoadGlobalConfig loads global configuration
func LoadGlobalConfig(filename string) error {
	config, err := LoadFromFile(filename)
	if err != nil {
		return err
	}

	SetGlobalConfig(config)
	return nil
}

// ReloadConfig 重新加载配置
// ReloadConfig reloads configuration
func ReloadConfig(filename string) error {
	return LoadGlobalConfig(filename)
}

// ConfigWatcher 配置文件监控器
// ConfigWatcher configuration file watcher
type ConfigWatcher struct {
	filename string
	config   *Config
	callback func(*Config)
	stopCh   chan struct{}
	mu       sync.RWMutex
}

// NewConfigWatcher 创建配置文件监控器
// NewConfigWatcher creates configuration file watcher
func NewConfigWatcher(filename string, callback func(*Config)) *ConfigWatcher {
	return &ConfigWatcher{
		filename: filename,
		callback: callback,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动配置文件监控
// Start starts configuration file monitoring
func (w *ConfigWatcher) Start() error {
	// 初始加载配置
	// Initial configuration loading
	config, err := LoadFromFile(w.filename)
	if err != nil {
		return errors.WrapErrorf(errors.ErrCodeFileReadError, "Failed to load initial configuration: %s", err, w.filename)
	}

	w.mu.Lock()
	w.config = config
	w.mu.Unlock()

	// 启动监控协程
	// Start monitoring goroutine
	go w.watchLoop()

	return nil
}

// Stop 停止配置文件监控
// Stop stops configuration file monitoring
func (w *ConfigWatcher) Stop() {
	close(w.stopCh)
}

// GetConfig 获取当前配置
// GetConfig gets current configuration
func (w *ConfigWatcher) GetConfig() *Config {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.config
}

// watchLoop 监控循环
// watchLoop monitoring loop
func (w *ConfigWatcher) watchLoop() {
	ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次文件变化
	defer ticker.Stop()

	var lastModTime time.Time

	// 获取初始修改时间
	// Get initial modification time
	if info, err := os.Stat(w.filename); err == nil {
		lastModTime = info.ModTime()
	}

	for {
		select {
		case <-ticker.C:
			// 检查文件是否被修改
			// Check if file has been modified
			info, err := os.Stat(w.filename)
			if err != nil {
				log.Errorf("Failed to stat configuration file %s: %v", w.filename, err)
				continue
			}

			if info.ModTime().After(lastModTime) {
				log.Infof("Configuration file %s has been modified, reloading...", w.filename)

				// 重新加载配置
				// Reload configuration
				config, err := LoadFromFile(w.filename)
				if err != nil {
					log.Errorf("Failed to reload configuration from %s: %v", w.filename, err)
					continue
				}

				w.mu.Lock()
				w.config = config
				w.mu.Unlock()

				// 调用回调函数
				// Call callback function
				if w.callback != nil {
					w.callback(config)
				}

				lastModTime = info.ModTime()
				log.Infof("Configuration reloaded successfully from %s", w.filename)
			}

		case <-w.stopCh:
			return
		}
	}
}

// ConfigManager 配置管理器
// ConfigManager configuration manager
type ConfigManager struct {
	configs map[string]*Config
	mu      sync.RWMutex
}

// NewConfigManager 创建配置管理器
// NewConfigManager creates configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		configs: make(map[string]*Config),
	}
}

// LoadConfig 加载配置
// LoadConfig loads configuration
func (m *ConfigManager) LoadConfig(name, filename string) error {
	config, err := LoadFromFile(filename)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs[name] = config
	return nil
}

// GetConfig 获取配置
// GetConfig gets configuration
func (m *ConfigManager) GetConfig(name string) *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if config, exists := m.configs[name]; exists {
		return config
	}

	return nil
}

// SetConfig 设置配置
// SetConfig sets configuration
func (m *ConfigManager) SetConfig(name string, config *Config) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs[name] = config
}

// RemoveConfig 移除配置
// RemoveConfig removes configuration
func (m *ConfigManager) RemoveConfig(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.configs, name)
}

// ListConfigs 列出所有配置名称
// ListConfigs lists all configuration names
func (m *ConfigManager) ListConfigs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.configs))
	for name := range m.configs {
		names = append(names, name)
	}

	return names
}

// 配置模板 Configuration templates

// ServerConfigTemplate 服务器配置模板
// ServerConfigTemplate server configuration template
func ServerConfigTemplate() *Config {
	config := DefaultConfig()

	// 服务器优化配置
	// Server optimized configuration
	config.Database.MaxConnections = 1000
	config.Storage.CacheSize = 1024 * 1024 * 1024          // 1GB
	config.Storage.BufferPoolSize = 2 * 1024 * 1024 * 1024 // 2GB
	config.Network.MaxConnections = 1000
	config.Log.Level = "info"
	config.Log.Async = true

	return config
}

// DevelopmentConfigTemplate 开发环境配置模板
// DevelopmentConfigTemplate development environment configuration template
func DevelopmentConfigTemplate() *Config {
	config := DefaultConfig()

	// 开发环境配置
	// Development environment configuration
	config.Database.MaxConnections = 100
	config.Storage.CacheSize = 128 * 1024 * 1024      // 128MB
	config.Storage.BufferPoolSize = 256 * 1024 * 1024 // 256MB
	config.Network.MaxConnections = 100
	config.Log.Level = "debug"
	config.Log.Output = "stdout"
	config.Log.Async = false

	return config
}

// ProductionConfigTemplate 生产环境配置模板
// ProductionConfigTemplate production environment configuration template
func ProductionConfigTemplate() *Config {
	config := DefaultConfig()

	// 生产环境配置
	// Production environment configuration
	config.Database.MaxConnections = 2000
	config.Storage.CacheSize = 4 * 1024 * 1024 * 1024      // 4GB
	config.Storage.BufferPoolSize = 8 * 1024 * 1024 * 1024 // 8GB
	config.Storage.SyncWrites = true
	config.Network.MaxConnections = 2000
	config.Network.TLS = &TLSConfig{
		Enabled: true,
	}
	config.Security.Authentication = &AuthConfig{
		Enabled: true,
		Method:  "password",
	}
	config.Security.Authorization = &AuthzConfig{
		Enabled: true,
		Model:   "rbac",
	}
	config.Security.Audit = &AuditConfig{
		Enabled: true,
	}
	config.Log.Level = "info"
	config.Log.Async = true
	config.Metrics.Enabled = true

	return config
}

// ClusterConfigTemplate 集群配置模板
// ClusterConfigTemplate cluster configuration template
func ClusterConfigTemplate() *Config {
	config := ProductionConfigTemplate()

	// 集群配置
	// Cluster configuration
	config.Cluster.Enabled = true
	config.Cluster.ReplicaFactor = 3
	config.Cluster.Consistency = "quorum"

	return config
}

// 配置验证规则 Configuration validation rules

// ValidationRule 验证规则接口
// ValidationRule validation rule interface
type ValidationRule interface {
	Validate(*Config) error
	Name() string
}

// PortRangeRule 端口范围验证规则
// PortRangeRule port range validation rule
type PortRangeRule struct{}

func (r *PortRangeRule) Name() string {
	return "PortRangeRule"
}

func (r *PortRangeRule) Validate(config *Config) error {
	if config.Network != nil {
		if config.Network.Port < 1024 || config.Network.Port > 65535 {
			return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Network port %d is out of valid range (1024-65535)", config.Network.Port)
		}
	}

	if config.Cluster != nil && config.Cluster.Enabled {
		if config.Cluster.Port < 1024 || config.Cluster.Port > 65535 {
			return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Cluster port %d is out of valid range (1024-65535)", config.Cluster.Port)
		}
	}

	if config.Metrics != nil && config.Metrics.Enabled {
		if config.Metrics.Port < 1024 || config.Metrics.Port > 65535 {
			return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Metrics port %d is out of valid range (1024-65535)", config.Metrics.Port)
		}
	}

	return nil
}

// MemoryLimitRule 内存限制验证规则
// MemoryLimitRule memory limit validation rule
type MemoryLimitRule struct {
	MaxMemory int64
}

func (r *MemoryLimitRule) Name() string {
	return "MemoryLimitRule"
}

func (r *MemoryLimitRule) Validate(config *Config) error {
	if config.Storage != nil {
		totalMemory := config.Storage.CacheSize + config.Storage.BufferPoolSize + config.Storage.WALBufferSize
		if totalMemory > r.MaxMemory {
			return errors.NewErrorf(errors.ErrCodeInvalidParameter, "Total memory usage %d exceeds limit %d", totalMemory, r.MaxMemory)
		}
	}

	return nil
}

// DiskSpaceRule 磁盘空间验证规则
// DiskSpaceRule disk space validation rule
type DiskSpaceRule struct {
	MinFreeSpace int64
}

func (r *DiskSpaceRule) Name() string {
	return "DiskSpaceRule"
}

func (r *DiskSpaceRule) Validate(config *Config) error {
	if config.Database != nil {
		dirs := []string{config.Database.DataDir, config.Database.WALDir, config.Database.TempDir}

		for _, dir := range dirs {
			if dir == "" {
				continue
			}

			// 检查磁盘空间（这里简化实现）
			// Check disk space (simplified implementation)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return errors.WrapErrorf(errors.ErrCodeSystemFailure, "Failed to create directory %s", err, dir)
			}
		}
	}

	return nil
}

// ConfigValidator 配置验证器
// ConfigValidator configuration validator
type ConfigValidator struct {
	rules []ValidationRule
}

// NewConfigValidator 创建配置验证器
// NewConfigValidator creates configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		rules: []ValidationRule{
			&PortRangeRule{},
			&MemoryLimitRule{MaxMemory: 16 * 1024 * 1024 * 1024}, // 16GB
			&DiskSpaceRule{MinFreeSpace: 1024 * 1024 * 1024},     // 1GB
		},
	}
}

// AddRule 添加验证规则
// AddRule adds validation rule
func (v *ConfigValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// RemoveRule 移除验证规则
// RemoveRule removes validation rule
func (v *ConfigValidator) RemoveRule(name string) {
	for i, rule := range v.rules {
		if rule.Name() == name {
			v.rules = append(v.rules[:i], v.rules[i+1:]...)
			break
		}
	}
}

// Validate 验证配置
// Validate validates configuration
func (v *ConfigValidator) Validate(config *Config) error {
	// 首先进行基本验证
	// First perform basic validation
	if err := ValidateConfig(config); err != nil {
		return err
	}

	// 然后应用自定义规则
	// Then apply custom rules
	for _, rule := range v.rules {
		if err := rule.Validate(config); err != nil {
			return errors.WrapErrorf(errors.ErrCodeInvalidParameter, "Validation rule %s failed", err, rule.Name())
		}
	}

	return nil
}

// 配置差异比较 Configuration diff comparison

// ConfigDiff 配置差异
// ConfigDiff configuration difference
type ConfigDiff struct {
	Path     string      `json:"path"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Action   string      `json:"action"` // "added", "removed", "modified"
}

// CompareConfigs 比较两个配置
// CompareConfigs compares two configurations
func CompareConfigs(old, new *Config) ([]ConfigDiff, error) {
	oldJSON, err := json.Marshal(old)
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeInvalidFormat, "Failed to marshal old configuration", err)
	}

	newJSON, err := json.Marshal(new)
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeInvalidFormat, "Failed to marshal new configuration", err)
	}

	var oldMap, newMap map[string]interface{}

	if err := json.Unmarshal(oldJSON, &oldMap); err != nil {
		return nil, errors.WrapError(errors.ErrCodeInvalidFormat, "Failed to unmarshal old configuration", err)
	}

	if err := json.Unmarshal(newJSON, &newMap); err != nil {
		return nil, errors.WrapError(errors.ErrCodeInvalidFormat, "Failed to unmarshal new configuration", err)
	}

	return compareMap(oldMap, newMap, ""), nil
}

// compareMap 递归比较映射
// compareMap recursively compares maps
func compareMap(old, new map[string]interface{}, prefix string) []ConfigDiff {
	var diffs []ConfigDiff

	// 检查新增和修改的键
	// Check added and modified keys
	for key, newValue := range new {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		if oldValue, exists := old[key]; exists {
			// 键存在，检查值是否变化
			// Key exists, check if value changed
			if !reflect.DeepEqual(oldValue, newValue) {
				// 如果都是映射，递归比较
				// If both are maps, compare recursively
				if oldMap, oldOk := oldValue.(map[string]interface{}); oldOk {
					if newMap, newOk := newValue.(map[string]interface{}); newOk {
						diffs = append(diffs, compareMap(oldMap, newMap, path)...)
						continue
					}
				}

				diffs = append(diffs, ConfigDiff{
					Path:     path,
					OldValue: oldValue,
					NewValue: newValue,
					Action:   "modified",
				})
			}
		} else {
			// 键不存在，表示新增
			// Key doesn't exist, indicates addition
			diffs = append(diffs, ConfigDiff{
				Path:     path,
				OldValue: nil,
				NewValue: newValue,
				Action:   "added",
			})
		}
	}

	// 检查删除的键
	// Check removed keys
	for key, oldValue := range old {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		if _, exists := new[key]; !exists {
			diffs = append(diffs, ConfigDiff{
				Path:     path,
				OldValue: oldValue,
				NewValue: nil,
				Action:   "removed",
			})
		}
	}

	return diffs
}

// init 初始化函数
// init initialization function
func init() {
	// 设置默认全局配置
	// Set default global configuration
	SetGlobalConfig(DefaultConfig())
}
