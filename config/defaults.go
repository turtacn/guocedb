package config

import "time"

// Default returns a configuration with default values.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            3306,
			MaxConnections:  1000,
			ConnectTimeout:  10 * time.Second,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     8 * time.Hour,
			ShutdownTimeout: 30 * time.Second,
		},
		Storage: StorageConfig{
			DataDir:         "./data",
			WALDir:          "",        // Default to DataDir
			MaxMemTableSize: 64 << 20,  // 64MB
			NumCompactors:   4,
			SyncWrites:      false,
			ValueLogGC:      true,
		},
		Security: SecurityConfig{
			Enabled:         false,
			RootPassword:    "",
			AuthPlugin:      "mysql_native_password",
			MaxAuthAttempts: 5,
			LockDuration:    15 * time.Minute,
			AuditLog: AuditLogConfig{
				Enabled:  false,
				FilePath: "./audit.log",
				Async:    true,
			},
		},
		Observability: ObservabilityConfig{
			Enabled:     true,
			Address:     ":9090",
			MetricsPath: "/metrics",
			EnablePprof: true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
		},
	}
}

// ApplyDefaults fills in missing configuration values with defaults.
func (c *Config) ApplyDefaults() {
	defaults := Default()

	// Server defaults
	if c.Server.Port == 0 {
		c.Server.Port = defaults.Server.Port
	}
	if c.Server.Host == "" {
		c.Server.Host = defaults.Server.Host
	}
	if c.Server.MaxConnections == 0 {
		c.Server.MaxConnections = defaults.Server.MaxConnections
	}
	if c.Server.ConnectTimeout == 0 {
		c.Server.ConnectTimeout = defaults.Server.ConnectTimeout
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = defaults.Server.ReadTimeout
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = defaults.Server.WriteTimeout
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = defaults.Server.IdleTimeout
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = defaults.Server.ShutdownTimeout
	}

	// Storage defaults
	if c.Storage.DataDir == "" {
		c.Storage.DataDir = defaults.Storage.DataDir
	}
	if c.Storage.MaxMemTableSize == 0 {
		c.Storage.MaxMemTableSize = defaults.Storage.MaxMemTableSize
	}
	if c.Storage.NumCompactors == 0 {
		c.Storage.NumCompactors = defaults.Storage.NumCompactors
	}

	// Security defaults
	if c.Security.AuthPlugin == "" {
		c.Security.AuthPlugin = defaults.Security.AuthPlugin
	}
	if c.Security.MaxAuthAttempts == 0 {
		c.Security.MaxAuthAttempts = defaults.Security.MaxAuthAttempts
	}
	if c.Security.LockDuration == 0 {
		c.Security.LockDuration = defaults.Security.LockDuration
	}
	if c.Security.AuditLog.FilePath == "" {
		c.Security.AuditLog.FilePath = defaults.Security.AuditLog.FilePath
	}

	// Observability defaults
	if c.Observability.Address == "" {
		c.Observability.Address = defaults.Observability.Address
	}
	if c.Observability.MetricsPath == "" {
		c.Observability.MetricsPath = defaults.Observability.MetricsPath
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = defaults.Logging.Level
	}
	if c.Logging.Format == "" {
		c.Logging.Format = defaults.Logging.Format
	}
	if c.Logging.Output == "" {
		c.Logging.Output = defaults.Logging.Output
	}
	if c.Logging.MaxSize == 0 {
		c.Logging.MaxSize = defaults.Logging.MaxSize
	}
	if c.Logging.MaxBackups == 0 {
		c.Logging.MaxBackups = defaults.Logging.MaxBackups
	}
	if c.Logging.MaxAge == 0 {
		c.Logging.MaxAge = defaults.Logging.MaxAge
	}
}
