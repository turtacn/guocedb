package config

import (
	"fmt"
	"strings"
	"time"
)

// ValidationError holds multiple validation errors.
type ValidationError struct {
	Errors []error
}

func (e *ValidationError) Error() string {
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	return "config validation failed:\n  - " + strings.Join(msgs, "\n  - ")
}

// Validate validates ServerConfig.
func (c *ServerConfig) Validate() error {
	var errs []error

	if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.port: must be between 1 and 65535, got %d", c.Port))
	}

	if c.MaxConnections < 1 {
		errs = append(errs, fmt.Errorf("server.max_connections: must be positive, got %d", c.MaxConnections))
	}

	if c.ShutdownTimeout < time.Second {
		errs = append(errs, fmt.Errorf("server.shutdown_timeout: must be at least 1s, got %v", c.ShutdownTimeout))
	}

	if c.ConnectTimeout < 0 {
		errs = append(errs, fmt.Errorf("server.connect_timeout: must be non-negative, got %v", c.ConnectTimeout))
	}

	if c.ReadTimeout < 0 {
		errs = append(errs, fmt.Errorf("server.read_timeout: must be non-negative, got %v", c.ReadTimeout))
	}

	if c.WriteTimeout < 0 {
		errs = append(errs, fmt.Errorf("server.write_timeout: must be non-negative, got %v", c.WriteTimeout))
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// Validate validates StorageConfig.
func (c *StorageConfig) Validate() error {
	var errs []error

	if c.DataDir == "" {
		errs = append(errs, fmt.Errorf("storage.data_dir: required"))
	}

	if c.MaxMemTableSize < 1<<20 { // Minimum 1MB
		errs = append(errs, fmt.Errorf("storage.max_memtable_size: must be at least 1MB, got %d", c.MaxMemTableSize))
	}

	if c.NumCompactors < 1 {
		errs = append(errs, fmt.Errorf("storage.num_compactors: must be positive, got %d", c.NumCompactors))
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// Validate validates SecurityConfig.
func (c *SecurityConfig) Validate() error {
	var errs []error

	if c.Enabled {
		if c.MaxAuthAttempts < 1 {
			errs = append(errs, fmt.Errorf("security.max_auth_attempts: must be positive, got %d", c.MaxAuthAttempts))
		}

		validPlugins := map[string]bool{
			"mysql_native_password": true,
			"caching_sha2_password": true,
		}
		if !validPlugins[c.AuthPlugin] {
			errs = append(errs, fmt.Errorf("security.auth_plugin: unsupported plugin %q", c.AuthPlugin))
		}

		if c.LockDuration < 0 {
			errs = append(errs, fmt.Errorf("security.lock_duration: must be non-negative, got %v", c.LockDuration))
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// Validate validates LoggingConfig.
func (c *LoggingConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[strings.ToLower(c.Level)] {
		return fmt.Errorf("logging.level: invalid level %q (must be debug, info, warn, or error)", c.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validFormats[strings.ToLower(c.Format)] {
		return fmt.Errorf("logging.format: invalid format %q (must be json or text)", c.Format)
	}

	if c.MaxSize < 0 {
		return fmt.Errorf("logging.max_size: must be non-negative, got %d", c.MaxSize)
	}

	if c.MaxBackups < 0 {
		return fmt.Errorf("logging.max_backups: must be non-negative, got %d", c.MaxBackups)
	}

	if c.MaxAge < 0 {
		return fmt.Errorf("logging.max_age: must be non-negative, got %d", c.MaxAge)
	}

	return nil
}
