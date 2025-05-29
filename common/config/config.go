// Package config defines the configuration structure and loading functions for the Guocedb project.
// This file is responsible for parsing configuration from various sources (e.g., YAML files,
// environment variables) and providing a centralized, type-safe way to access system settings.
// It relies on common/constants for default values and common/errors for consistent error reporting.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync" // For the singleton pattern.

	"gopkg.in/yaml.v3" // For parsing YAML configuration files.

	"github.com/turtacn/guocedb/common/constants"  // Import constants for default configuration values.
	"github.com/turtacn/guocedb/common/errors"     // Import errors for consistent error handling.
	"github.com/turtacn/guocedb/common/log"        // Import log for logging configuration-related messages.
	"github.com/turtacn/guocedb/common/types/enum" // Import enum for log levels and component types.

	"go.uber.org/zap" // For structured logging in the config package.
)

// Config represents the top-level configuration structure for Guocedb.
// Each field corresponds to a configurable aspect of the database.
type Config struct {
	// ServerConfig holds configurations related to the database server.
	Server ServerConfig `yaml:"server"`
	// NetworkConfig holds configurations related to network communication.
	Network NetworkConfig `yaml:"network"`
	// StorageConfig holds configurations related to the storage layer.
	Storage StorageConfig `yaml:"storage"`
	// SecurityConfig holds configurations related to authentication and authorization.
	Security SecurityConfig `yaml:"security"`
	// LogConfig holds configurations related to logging.
	Log LogConfig `yaml:"log"`
	// CatalogConfig holds configurations related to the metadata catalog.
	Catalog CatalogConfig `yaml:"catalog"`
	// PerformanceConfig holds general performance tuning parameters.
	Performance PerformanceConfig `yaml:"performance"`
	// DebugConfig holds configurations for debugging features.
	Debug DebugConfig `yaml:"debug"`
}

// ServerConfig defines server-specific settings.
type ServerConfig struct {
	// DataPath is the base directory for all database data.
	DataPath string `yaml:"dataPath"`
	// EnableHTTPAPI enables or disables the HTTP management API.
	EnableHTTPAPI bool `yaml:"enableHttpAPI"`
	// EnableGRPCAPI enables or disables the gRPC management API.
	EnableGRPCAPI bool `yaml:"enableGrpcAPI"`
	// RunInReadOnlyMode starts the server in read-only mode.
	RunInReadOnlyMode bool `yaml:"readOnly"`
	// PidFile is the path to the process ID file.
	PidFile string `yaml:"pidFile"`
}

// NetworkConfig defines network-specific settings.
type NetworkConfig struct {
	// MySQLPort is the port for MySQL protocol connections.
	MySQLPort int `yaml:"mysqlPort"`
	// ManagementGRPCPort is the port for the gRPC management API.
	ManagementGRPCPort int `yaml:"managementGrpcPort"`
	// ManagementRESTPort is the port for the RESTful management API (future).
	ManagementRESTPort int `yaml:"managementRestPort"`
	// ListenAddress is the IP address to listen on (e.g., "0.0.0.0" for all interfaces).
	ListenAddress string `yaml:"listenAddress"`
	// ReadTimeout specifies the default timeout for network read operations.
	ReadTimeout string `yaml:"readTimeout"` // e.g., "5s", "1m"
	// WriteTimeout specifies the default timeout for network write operations.
	WriteTimeout string `yaml:"writeTimeout"` // e.g., "5s", "1m"
	// MaxConnections defines the maximum number of concurrent client connections.
	MaxConnections int `yaml:"maxConnections"`
}

// StorageConfig defines storage layer settings.
type StorageConfig struct {
	// StorageEngine specifies the primary storage engine to use (e.g., "badger").
	StorageEngine string `yaml:"storageEngine"`
	// Badger specific configurations.
	Badger BadgerConfig `yaml:"badger"`
	// BlockSize is the default block size for storage operations (e.g., in bytes).
	BlockSize int `yaml:"blockSize"`
	// CacheSizeMB is the size of the storage cache in megabytes.
	CacheSizeMB int `yaml:"cacheSizeMB"`
}

// BadgerConfig defines specific configurations for the Badger KV store.
type BadgerConfig struct {
	// Path is the directory for BadgerDB data.
	Path string `yaml:"path"`
	// ValueLogFileSize is the maximum size of a BadgerDB value log file.
	ValueLogFileSize int `yaml:"valueLogFileSize"`
	// SyncWrites enables synchronous writes for BadgerDB.
	SyncWrites bool `yaml:"syncWrites"`
	// MaxTableSize is the maximum size of a BadgerDB SST table.
	MaxTableSize int `yaml:"maxTableSize"`
}

// SecurityConfig defines security-related settings.
type SecurityConfig struct {
	// EnableAuthentication controls whether user authentication is required.
	EnableAuthentication bool `yaml:"enableAuthentication"`
	// DefaultAuthMethod is the default authentication method for new users.
	DefaultAuthMethod string `yaml:"defaultAuthMethod"` // e.g., "native", "sha256"
	// RootPassword specifies the initial password for the 'root' user.
	RootPassword string `yaml:"rootPassword"` // WARNING: For initial setup, should be changed.
	// TLSConfig holds TLS/SSL configuration.
	TLS TLSConfig `yaml:"tls"`
}

// TLSConfig defines TLS/SSL settings for secure communication.
type TLSConfig struct {
	// EnableTLS enables or disables TLS.
	EnableTLS bool `yaml:"enableTLS"`
	// CertFile is the path to the TLS certificate file.
	CertFile string `yaml:"certFile"`
	// KeyFile is the path to the TLS private key file.
	KeyFile string `yaml:"keyFile"`
	// CAFile is the path to the CA certificate file for client authentication.
	CAFile string `yaml:"caFile"`
}

// LogConfig defines logging-related settings.
type LogConfig struct {
	// Level is the minimum logging level (e.g., "DEBUG", "INFO", "WARN", "ERROR", "FATAL").
	Level string `yaml:"level"`
	// FilePath is the path to the log file. If empty, logs to stdout.
	FilePath string `yaml:"filePath"`
	// MaxSizeMB is the maximum size of a log file before rotation (in MB).
	MaxSizeMB int `yaml:"maxSizeMB"`
	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int `yaml:"maxBackups"`
	// MaxAgeDays is the maximum number of days to retain old log files.
	MaxAgeDays int `yaml:"maxAgeDays"`
	// EnableConsoleOutput controls whether logs are also printed to the console.
	EnableConsoleOutput bool `yaml:"enableConsoleOutput"`
}

// CatalogConfig defines settings for the metadata catalog.
type CatalogConfig struct {
	// PersistIntervalSeconds is the interval (in seconds) at which catalog changes are persisted.
	PersistIntervalSeconds int `yaml:"persistIntervalSeconds"`
	// MaxMemoryCatalogSizeMB is the maximum memory usage for the in-memory catalog cache (in MB).
	MaxMemoryCatalogSizeMB int `yaml:"maxMemoryCatalogSizeMB"`
}

// PerformanceConfig defines general performance tuning parameters.
type PerformanceConfig struct {
	// QueryCacheSizeMB is the size of the query result cache in megabytes.
	QueryCacheSizeMB int `yaml:"queryCacheSizeMB"`
	// MaxConcurrentQueries is the maximum number of queries that can execute concurrently.
	MaxConcurrentQueries int `yaml:"maxConcurrentQueries"`
}

// DebugConfig defines settings for debugging features.
type DebugConfig struct {
	// EnablePprof enables the pprof profiling server.
	EnablePprof bool `yaml:"enablePprof"`
	// PprofPort is the port for the pprof server.
	PprofPort int `yaml:"pprofPort"`
	// EnableSQLTracing enables detailed SQL execution tracing.
	EnableSQLTracing bool `yaml:"enableSqlTracing"`
}

// globalConfig is the singleton instance of the Guocedb configuration.
var globalConfig *Config
var configOnce sync.Once

// LoadConfig initializes and loads the global configuration.
// It first applies default values, then overrides them with values from
// the specified YAML file, and finally with environment variables.
func LoadConfig(configPath string) error {
	var err error
	configOnce.Do(func() {
		cfg := &Config{}

		// 1. Apply default values
		cfg.applyDefaults()

		// 2. Load from YAML file if provided
		if configPath != "" {
			if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
				log.GetLogger().Info("Configuration file not found, using defaults and environment variables.", zap.String("path", configPath))
			} else if statErr != nil {
				err = errors.NewGuocedbError(enum.ErrConfiguration, errors.CodeConfigLoadFailed,
					fmt.Sprintf("failed to stat configuration file %s", configPath), statErr)
				return
			} else {
				fileContent, readErr := os.ReadFile(configPath)
				if readErr != nil {
					err = errors.NewGuocedbError(enum.ErrConfiguration, errors.CodeConfigLoadFailed,
						fmt.Sprintf("failed to read configuration file %s", configPath), readErr)
					return
				}
				if unmarshalErr := yaml.Unmarshal(fileContent, cfg); unmarshalErr != nil {
					err = errors.NewGuocedbError(enum.ErrConfiguration, errors.CodeConfigInvalidValue,
						fmt.Sprintf("failed to unmarshal configuration file %s", configPath), unmarshalErr)
					return
				}
				log.GetLogger().Info("Configuration loaded from file.", zap.String("path", configPath))
			}
		} else {
			log.GetLogger().Info("No configuration file path provided, using defaults and environment variables.")
		}

		// 3. Override with environment variables
		cfg.overrideWithEnv()

		// 4. Validate and sanitize paths
		cfg.sanitizePaths()

		globalConfig = cfg
	})
	return err
}

// GetConfig returns the global configuration instance.
// It should be called after LoadConfig has successfully run.
func GetConfig() *Config {
	if globalConfig == nil {
		// This indicates LoadConfig was not called or failed.
		// In a production app, you might want to panic or return a default/error.
		// For now, logging a warning and returning a zero-value config.
		log.GetLogger().Fatal("Configuration not initialized. Call LoadConfig() first.")
		return &Config{} // Return a zero-value config to avoid nil panics in some cases.
	}
	return globalConfig
}

// applyDefaults sets the default values for the configuration.
func (c *Config) applyDefaults() {
	c.Server.DataPath = constants.DefaultDataPath
	c.Server.EnableHTTPAPI = false // Disabled by default
	c.Server.EnableGRPCAPI = true  // Enabled by default
	c.Server.RunInReadOnlyMode = false
	c.Server.PidFile = filepath.Join(os.TempDir(), constants.ProjectName+".pid")

	c.Network.MySQLPort = constants.DefaultMySQLPort
	c.Network.ManagementGRPCPort = constants.DefaultManagementGRPCPort
	c.Network.ManagementRESTPort = constants.DefaultManagementRESTPort
	c.Network.ListenAddress = "0.0.0.0" // Listen on all interfaces by default
	// Using String() method from time.Duration for NetworkReadTimeout and NetworkWriteTimeout
	c.Network.ReadTimeout = constants.NetworkReadTimeout.String()
	c.Network.WriteTimeout = constants.NetworkWriteTimeout.String()
	c.Network.MaxConnections = constants.MaxConnections

	c.Storage.StorageEngine = "badger" // Default to Badger
	c.Storage.BlockSize = 8192         // 8KB default block size
	c.Storage.CacheSizeMB = 256        // 256MB default cache

	c.Storage.Badger.Path = constants.DefaultBadgerPath
	c.Storage.Badger.ValueLogFileSize = constants.BadgerValueLogFileSize
	c.Storage.Badger.SyncWrites = constants.BadgerSyncWrites
	c.Storage.Badger.MaxTableSize = 64 << 20 // 64 MB

	c.Security.EnableAuthentication = true // Enabled by default
	c.Security.DefaultAuthMethod = "native"
	c.Security.RootPassword = constants.DefaultPassword // WARNING: Change this!
	c.Security.TLS.EnableTLS = false
	c.Security.TLS.CertFile = ""
	c.Security.TLS.KeyFile = ""
	c.Security.TLS.CAFile = ""

	c.Log.Level = constants.DefaultLogLevel
	c.Log.FilePath = constants.DefaultLogFilePath
	c.Log.MaxSizeMB = constants.LogFileMaxSizeMB
	c.Log.MaxBackups = constants.LogFileMaxBackups
	c.Log.MaxAgeDays = constants.LogFileMaxAgeDays
	c.Log.EnableConsoleOutput = true // Console output enabled by default

	c.Catalog.PersistIntervalSeconds = 300 // Persist every 5 minutes
	c.Catalog.MaxMemoryCatalogSizeMB = 128 // 128 MB

	c.Performance.QueryCacheSizeMB = 64 // 64 MB
	c.Performance.MaxConcurrentQueries = 100

	c.Debug.EnablePprof = false
	c.Debug.PprofPort = 6060
	c.Debug.EnableSQLTracing = false
}

// overrideWithEnv overrides configuration values with environment variables.
// Environment variables are expected in the format GUOCEDB_SECTION_FIELD (e.g., GUOCEDB_NETWORK_MYSQLPORT).
func (c *Config) overrideWithEnv() {
	logger := log.GetLogger().With(zap.String("component", enum.ComponentConfig.String()))

	// Use reflection or explicit checks for each field for robust handling.
	// For simplicity, showing a few examples. A more generic solution would use reflection.

	if val := os.Getenv("GUOCEDB_SERVER_DATAPATH"); val != "" {
		c.Server.DataPath = val
		logger.Debug("Config override by env", zap.String("key", "GUOCEDB_SERVER_DATAPATH"), zap.String("value", val))
	}
	if val := os.Getenv("GUOCEDB_NETWORK_MYSQLPORT"); val != "" {
		if port, err := ParseInt(val); err == nil {
			c.Network.MySQLPort = port
			logger.Debug("Config override by env", zap.String("key", "GUOCEDB_NETWORK_MYSQLPORT"), zap.Int("value", port))
		} else {
			logger.Warn("Invalid GUOCEDB_NETWORK_MYSQLPORT environment variable", zap.String("value", val), zap.Error(err))
		}
	}
	if val := os.Getenv("GUOCEDB_LOG_LEVEL"); val != "" {
		// Validate against enum values to prevent invalid log levels
		if _, err := enum.ParseLogLevel(strings.ToUpper(val)); err == nil {
			c.Log.Level = strings.ToUpper(val)
			logger.Debug("Config override by env", zap.String("key", "GUOCEDB_LOG_LEVEL"), zap.String("value", val))
		} else {
			logger.Warn("Invalid GUOCEDB_LOG_LEVEL environment variable", zap.String("value", val), zap.Error(err))
		}
	}
	if val := os.Getenv("GUOCEDB_LOG_FILEPATH"); val != "" {
		c.Log.FilePath = val
		logger.Debug("Config override by env", zap.String("key", "GUOCEDB_LOG_FILEPATH"), zap.String("value", val))
	}
	if val := os.Getenv("GUOCEDB_SECURITY_ROOTPASSWORD"); val != "" {
		c.Security.RootPassword = val
		logger.Warn("Root password overridden by environment variable. Ensure this is secure!", zap.String("key", "GUOCEDB_SECURITY_ROOTPASSWORD"))
	}
	// ... (add more environment variable checks for other critical fields)
}

// sanitizePaths ensures that paths are absolute and creates directories if they don't exist.
func (c *Config) sanitizePaths() {
	logger := log.GetLogger().With(zap.String("component", enum.ComponentConfig.String()))

	// DataPath
	c.Server.DataPath = expandPath(c.Server.DataPath)
	if err := os.MkdirAll(c.Server.DataPath, 0755); err != nil {
		logger.Error("Failed to create data directory", zap.String("path", c.Server.DataPath), zap.Error(err))
	}

	// BadgerDB Path
	if c.Storage.Badger.Path == "" {
		c.Storage.Badger.Path = filepath.Join(c.Server.DataPath, "badger") // Default if not specified in config
	}
	c.Storage.Badger.Path = expandPath(c.Storage.Badger.Path)
	if err := os.MkdirAll(c.Storage.Badger.Path, 0755); err != nil {
		logger.Error("Failed to create BadgerDB data directory", zap.String("path", c.Storage.Badger.Path), zap.Error(err))
	}

	// Log File Path
	if c.Log.FilePath != "" {
		c.Log.FilePath = expandPath(c.Log.FilePath)
		logDir := filepath.Dir(c.Log.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			logger.Error("Failed to create log directory", zap.String("path", logDir), zap.Error(err))
		}
	}

	// PidFile path
	if c.Server.PidFile != "" {
		c.Server.PidFile = expandPath(c.Server.PidFile)
		pidDir := filepath.Dir(c.Server.PidFile)
		if err := os.MkdirAll(pidDir, 0755); err != nil {
			logger.Error("Failed to create PID file directory", zap.String("path", pidDir), zap.Error(err))
		}
	}

	// TLS cert/key/ca paths
	if c.Security.TLS.EnableTLS {
		if c.Security.TLS.CertFile != "" {
			c.Security.TLS.CertFile = expandPath(c.Security.TLS.CertFile)
		}
		if c.Security.TLS.KeyFile != "" {
			c.Security.TLS.KeyFile = expandPath(c.Security.TLS.KeyFile)
		}
		if c.Security.TLS.CAFile != "" {
			c.Security.TLS.CAFile = expandPath(c.Security.TLS.CAFile)
		}
	}
}

// expandPath expands home directory (~) and makes path absolute.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[1:])
		}
	}
	absPath, err := filepath.Abs(path)
	if err == nil {
		return absPath
	}
	return path // Fallback if absolute path conversion fails
}

// ParseInt is a helper to parse string to int, used for env variables.
func ParseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

// ValidateConfig can be used to perform more complex semantic validations after loading.
func ValidateConfig() error {
	cfg := GetConfig()
	if cfg == nil {
		return errors.NewGuocedbError(enum.ErrConfiguration, errors.CodeConfigLoadFailed, "configuration not loaded", nil)
	}

	// Example validation: MySQLPort cannot be 0
	if cfg.Network.MySQLPort == 0 {
		return errors.NewGuocedbError(enum.ErrConfiguration, errors.CodeConfigInvalidValue, "MySQLPort cannot be 0", nil)
	}

	// Example validation: Check if dataPath is writable
	if _, err := os.Stat(cfg.Server.DataPath); err != nil {
		if os.IsNotExist(err) {
			return errors.NewGuocedbError(enum.ErrConfiguration, errors.CodeConfigInvalidValue,
				fmt.Sprintf("data path does not exist: %s", cfg.Server.DataPath), err)
		}
		// Attempt to create a dummy file to check writability
		testFile := filepath.Join(cfg.Server.DataPath, "test_write.tmp")
		if writeErr := os.WriteFile(testFile, []byte("test"), 0644); writeErr != nil {
			return errors.NewGuocedbError(enum.ErrConfiguration, errors.CodePermissionDenied,
				fmt.Sprintf("data path is not writable: %s", cfg.Server.DataPath), writeErr)
		}
		os.Remove(testFile) // Clean up test file
	}

	// Validate Log Level
	if _, err := enum.ParseLogLevel(cfg.Log.Level); err != nil {
		return errors.NewGuocedbError(enum.ErrConfiguration, errors.CodeConfigInvalidValue,
			fmt.Sprintf("invalid log level specified: %s", cfg.Log.Level), err)
	}

	// Validate storage engine type
	// This assumes enum.StorageEngineType has a String() method and ParseStorageEngineType function
	// For example, in common/types/enum/enum.go:
	// func ParseStorageEngineType(s string) (StorageEngineType, error) { ... }
	// if _, err := enum.ParseStorageEngineType(cfg.Storage.StorageEngine); err != nil {
	// 	return errors.NewGuocedbError(enum.ErrConfiguration, errors.CodeConfigInvalidValue,
	// 		fmt.Sprintf("unsupported storage engine: %s", cfg.Storage.StorageEngine), err)
	// }

	// Add more complex validations here...
	return nil
}

//Personal.AI order the ending
