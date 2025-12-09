// Package config handles centralized configuration management for GuoceDB.
package config

import (
	"time"
)

// Config holds the entire configuration for GuoceDB.
type Config struct {
	Server        ServerConfig        `yaml:"server" mapstructure:"server"`
	Storage       StorageConfig       `yaml:"storage" mapstructure:"storage"`
	Security      SecurityConfig      `yaml:"security" mapstructure:"security"`
	Observability ObservabilityConfig `yaml:"observability" mapstructure:"observability"`
	Logging       LoggingConfig       `yaml:"logging" mapstructure:"logging"`
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Host            string        `yaml:"host" mapstructure:"host"`
	Port            int           `yaml:"port" mapstructure:"port"`
	MaxConnections  int           `yaml:"max_connections" mapstructure:"max_connections"`
	ConnectTimeout  time.Duration `yaml:"connect_timeout" mapstructure:"connect_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" mapstructure:"shutdown_timeout"`
}

// StorageConfig holds storage-related configuration.
type StorageConfig struct {
	DataDir         string `yaml:"data_dir" mapstructure:"data_dir"`
	WALDir          string `yaml:"wal_dir" mapstructure:"wal_dir"`
	MaxMemTableSize int64  `yaml:"max_memtable_size" mapstructure:"max_memtable_size"`
	NumCompactors   int    `yaml:"num_compactors" mapstructure:"num_compactors"`
	SyncWrites      bool   `yaml:"sync_writes" mapstructure:"sync_writes"`
	ValueLogGC      bool   `yaml:"valuelog_gc" mapstructure:"valuelog_gc"`
}

// SecurityConfig holds security-related configuration.
type SecurityConfig struct {
	Enabled         bool           `yaml:"enabled" mapstructure:"enabled"`
	RootPassword    string         `yaml:"root_password" mapstructure:"root_password"`
	AuthPlugin      string         `yaml:"auth_plugin" mapstructure:"auth_plugin"`
	MaxAuthAttempts int            `yaml:"max_auth_attempts" mapstructure:"max_auth_attempts"`
	LockDuration    time.Duration  `yaml:"lock_duration" mapstructure:"lock_duration"`
	AuditLog        AuditLogConfig `yaml:"audit_log" mapstructure:"audit_log"`
}

// AuditLogConfig holds audit logging configuration.
type AuditLogConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	FilePath string `yaml:"file_path" mapstructure:"file_path"`
	Async    bool   `yaml:"async" mapstructure:"async"`
}

// ObservabilityConfig holds observability configuration.
type ObservabilityConfig struct {
	Enabled     bool   `yaml:"enabled" mapstructure:"enabled"`
	Address     string `yaml:"address" mapstructure:"address"`
	MetricsPath string `yaml:"metrics_path" mapstructure:"metrics_path"`
	EnablePprof bool   `yaml:"enable_pprof" mapstructure:"enable_pprof"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level      string `yaml:"level" mapstructure:"level"`
	Format     string `yaml:"format" mapstructure:"format"`        // json, text
	Output     string `yaml:"output" mapstructure:"output"`        // stdout, file path
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`    // MB
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"`
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`      // days
}

// Validate validates the entire configuration.
func (c *Config) Validate() error {
	var errs []error

	if err := c.Server.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Storage.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Security.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}
