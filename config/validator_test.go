package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"valid port 3306", 3306, false},
		{"valid port 1", 1, false},
		{"valid port 65535", 65535, false},
		{"invalid port 0", 0, true},
		{"invalid port -1", -1, true},
		{"invalid port 70000", 70000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ServerConfig{
				Port:            tt.port,
				MaxConnections:  100,
				ShutdownTimeout: time.Second,
			}
			err := cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	// Empty data dir should fail
	cfg := StorageConfig{
		DataDir:         "",
		MaxMemTableSize: 1 << 20,
		NumCompactors:   1,
	}
	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "data_dir")
}

func TestValidatePath(t *testing.T) {
	// Valid data dir
	cfg := StorageConfig{
		DataDir:         "/tmp/data",
		MaxMemTableSize: 1 << 20,
		NumCompactors:   1,
	}
	err := cfg.Validate()
	require.NoError(t, err)
}

func TestValidateMemTableSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{"valid 1MB", 1 << 20, false},
		{"valid 64MB", 64 << 20, false},
		{"invalid 100B", 100, true},
		{"invalid 0", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := StorageConfig{
				DataDir:         "/tmp/data",
				MaxMemTableSize: tt.size,
				NumCompactors:   1,
			}
			err := cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateAuthPlugin(t *testing.T) {
	tests := []struct {
		name    string
		plugin  string
		wantErr bool
	}{
		{"valid mysql_native_password", "mysql_native_password", false},
		{"valid caching_sha2_password", "caching_sha2_password", false},
		{"invalid plugin", "invalid_plugin", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := SecurityConfig{
				Enabled:         true,
				AuthPlugin:      tt.plugin,
				MaxAuthAttempts: 5,
			}
			err := cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{"valid debug", "debug", false},
		{"valid info", "info", false},
		{"valid warn", "warn", false},
		{"valid error", "error", false},
		{"invalid level", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := LoggingConfig{
				Level:  tt.level,
				Format: "json",
			}
			err := cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateLogFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"valid json", "json", false},
		{"valid text", "text", false},
		{"invalid format", "xml", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := LoggingConfig{
				Level:  "info",
				Format: tt.format,
			}
			err := cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	// Multiple validation errors
	cfg := ServerConfig{
		Port:            0, // Invalid
		MaxConnections:  -1, // Invalid
		ShutdownTimeout: 0, // Invalid
	}
	err := cfg.Validate()
	require.Error(t, err)

	// Should contain all error messages
	errMsg := err.Error()
	require.Contains(t, errMsg, "port")
	require.Contains(t, errMsg, "max_connections")
	require.Contains(t, errMsg, "shutdown_timeout")
}
