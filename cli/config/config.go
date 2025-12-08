// Package config handles CLI-specific configuration for guocedb.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

// Config holds the entire configuration for the CLI application.
type Config struct {
	Server      ServerConfig      `yaml:"server" json:"server"`
	Storage     StorageConfig     `yaml:"storage" json:"storage"`
	Log         LogConfig         `yaml:"log" json:"log"`
	Auth        AuthConfig        `yaml:"auth" json:"auth"`
	Performance PerformanceConfig `yaml:"performance" json:"performance"`
	Monitoring  MonitoringConfig  `yaml:"monitoring" json:"monitoring"`
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Host           string `yaml:"host" json:"host"`
	Port           int    `yaml:"port" json:"port"`
	ReadTimeout    string `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout   string `yaml:"write_timeout" json:"write_timeout"`
	MaxConnections int    `yaml:"max_connections" json:"max_connections"`
}

// StorageConfig holds storage-related configuration.
type StorageConfig struct {
	DataDir    string       `yaml:"data_dir" json:"data_dir"`
	WalDir     string       `yaml:"wal_dir" json:"wal_dir"`
	SyncWrites bool         `yaml:"sync_writes" json:"sync_writes"`
	Badger     BadgerConfig `yaml:"badger" json:"badger"`
}

// BadgerConfig holds configuration specific to the Badger storage engine.
type BadgerConfig struct {
	ValueLogFileSize int `yaml:"value_log_file_size" json:"value_log_file_size"`
	NumMemtables     int `yaml:"num_memtables" json:"num_memtables"`
	NumCompactors    int `yaml:"num_compactors" json:"num_compactors"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level     string `yaml:"level" json:"level"`
	Format    string `yaml:"format" json:"format"`
	Output    string `yaml:"output" json:"output"`
	AddSource bool   `yaml:"add_source" json:"add_source"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Enabled bool         `yaml:"enabled" json:"enabled"`
	Users   []UserConfig `yaml:"users" json:"users"`
}

// UserConfig holds user authentication configuration.
type UserConfig struct {
	Username  string   `yaml:"username" json:"username"`
	Password  string   `yaml:"password" json:"password"`
	Databases []string `yaml:"databases" json:"databases"`
}

// PerformanceConfig holds performance tuning configuration.
type PerformanceConfig struct {
	QueryCacheSize   int `yaml:"query_cache_size" json:"query_cache_size"`
	SortBufferSize   int `yaml:"sort_buffer_size" json:"sort_buffer_size"`
	JoinBufferSize   int `yaml:"join_buffer_size" json:"join_buffer_size"`
}

// MonitoringConfig holds monitoring configuration.
type MonitoringConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Port    int    `yaml:"port" json:"port"`
	Path    string `yaml:"path" json:"path"`
}

// Load initializes the configuration from a file, environment variables, and defaults.
func Load(path string) (*Config, error) {
	cfg := &Config{}
	
	// 1. Set default values
	setDefaults(cfg)
	
	// 2. If config file is specified, load YAML
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
		}
		
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
		}
	}
	
	// 3. Override with environment variables
	loadFromEnv(cfg)
	
	// 4. Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return cfg, nil
}

// Validate validates the configuration values.
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be 1-65535)", c.Server.Port)
	}
	
	if c.Storage.DataDir == "" {
		return fmt.Errorf("data_dir is required")
	}
	
	// Validate timeout strings
	if c.Server.ReadTimeout != "" {
		if _, err := time.ParseDuration(c.Server.ReadTimeout); err != nil {
			return fmt.Errorf("invalid read_timeout: %w", err)
		}
	}
	
	if c.Server.WriteTimeout != "" {
		if _, err := time.ParseDuration(c.Server.WriteTimeout); err != nil {
			return fmt.Errorf("invalid write_timeout: %w", err)
		}
	}
	
	// Validate log level
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLevels[c.Log.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.Log.Level)
	}
	
	// Validate log format
	validFormats := map[string]bool{
		"text": true, "json": true,
	}
	if !validFormats[c.Log.Format] {
		return fmt.Errorf("invalid log format: %s (must be text or json)", c.Log.Format)
	}
	
	return nil
}

// setDefaults sets default configuration values.
func setDefaults(cfg *Config) {
	// Server defaults
	cfg.Server.Host = "0.0.0.0"
	cfg.Server.Port = 3306
	cfg.Server.ReadTimeout = "30s"
	cfg.Server.WriteTimeout = "30s"
	cfg.Server.MaxConnections = 1000
	
	// Storage defaults
	cfg.Storage.DataDir = "./data"
	cfg.Storage.WalDir = ""
	cfg.Storage.SyncWrites = true
	cfg.Storage.Badger.ValueLogFileSize = 1073741824 // 1GB
	cfg.Storage.Badger.NumMemtables = 5
	cfg.Storage.Badger.NumCompactors = 4
	
	// Log defaults
	cfg.Log.Level = "info"
	cfg.Log.Format = "text"
	cfg.Log.Output = "stderr"
	cfg.Log.AddSource = false
	
	// Auth defaults
	cfg.Auth.Enabled = false
	cfg.Auth.Users = []UserConfig{}
	
	// Performance defaults
	cfg.Performance.QueryCacheSize = 0
	cfg.Performance.SortBufferSize = 262144  // 256KB
	cfg.Performance.JoinBufferSize = 262144  // 256KB
	
	// Monitoring defaults
	cfg.Monitoring.Enabled = false
	cfg.Monitoring.Port = 9090
	cfg.Monitoring.Path = "/metrics"
}

// loadFromEnv loads configuration from environment variables.
func loadFromEnv(cfg *Config) {
	if v := os.Getenv("GUOCEDB_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("GUOCEDB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("GUOCEDB_DATA_DIR"); v != "" {
		cfg.Storage.DataDir = v
	}
	if v := os.Getenv("GUOCEDB_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("GUOCEDB_LOG_FORMAT"); v != "" {
		cfg.Log.Format = v
	}
	if v := os.Getenv("GUOCEDB_READ_TIMEOUT"); v != "" {
		cfg.Server.ReadTimeout = v
	}
	if v := os.Getenv("GUOCEDB_WRITE_TIMEOUT"); v != "" {
		cfg.Server.WriteTimeout = v
	}
	if v := os.Getenv("GUOCEDB_MAX_CONNECTIONS"); v != "" {
		if maxConn, err := strconv.Atoi(v); err == nil {
			cfg.Server.MaxConnections = maxConn
		}
	}
}