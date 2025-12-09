package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	require.Equal(t, 3306, cfg.Server.Port)
	require.Equal(t, "0.0.0.0", cfg.Server.Host)
	require.Equal(t, 1000, cfg.Server.MaxConnections)
	require.Equal(t, "./data", cfg.Storage.DataDir)
	require.False(t, cfg.Security.Enabled)
	require.True(t, cfg.Observability.Enabled)
	require.Equal(t, "info", cfg.Logging.Level)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{
			name:    "valid default",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name:    "invalid port zero",
			modify:  func(c *Config) { c.Server.Port = 0 },
			wantErr: true,
		},
		{
			name:    "invalid port too high",
			modify:  func(c *Config) { c.Server.Port = 70000 },
			wantErr: true,
		},
		{
			name:    "empty data dir",
			modify:  func(c *Config) { c.Storage.DataDir = "" },
			wantErr: true,
		},
		{
			name:    "invalid log level",
			modify:  func(c *Config) { c.Logging.Level = "invalid" },
			wantErr: true,
		},
		{
			name:    "negative max connections",
			modify:  func(c *Config) { c.Server.MaxConnections = -1 },
			wantErr: true,
		},
		{
			name:    "invalid memtable size",
			modify:  func(c *Config) { c.Storage.MaxMemTableSize = 100 }, // < 1MB
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.modify(cfg)
			err := cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.ApplyDefaults()

	// Verify defaults are applied
	require.Equal(t, 3306, cfg.Server.Port)
	require.Equal(t, "0.0.0.0", cfg.Server.Host)
	require.Equal(t, "./data", cfg.Storage.DataDir)
	require.Equal(t, "info", cfg.Logging.Level)
}

func TestConfigTimeouts(t *testing.T) {
	cfg := Default()

	require.Equal(t, 10*time.Second, cfg.Server.ConnectTimeout)
	require.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	require.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
	require.Equal(t, 8*time.Hour, cfg.Server.IdleTimeout)
	require.Equal(t, 30*time.Second, cfg.Server.ShutdownTimeout)
}

func TestSecurityConfig(t *testing.T) {
	cfg := Default()

	// By default, security is disabled
	require.False(t, cfg.Security.Enabled)
	require.Equal(t, "mysql_native_password", cfg.Security.AuthPlugin)
	require.Equal(t, 5, cfg.Security.MaxAuthAttempts)
	require.Equal(t, 15*time.Minute, cfg.Security.LockDuration)
}

func TestObservabilityConfig(t *testing.T) {
	cfg := Default()

	require.True(t, cfg.Observability.Enabled)
	require.Equal(t, ":9090", cfg.Observability.Address)
	require.Equal(t, "/metrics", cfg.Observability.MetricsPath)
	require.True(t, cfg.Observability.EnablePprof)
}
