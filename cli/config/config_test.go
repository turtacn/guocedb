package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.yaml")
	
	configContent := `
server:
  host: "127.0.0.1"
  port: 3307
  max_connections: 500
storage:
  data_dir: "/tmp/test"
  sync_writes: false
log:
  level: "debug"
  format: "json"
`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Load config
	cfg, err := Load(configFile)
	require.NoError(t, err)
	
	// Verify values
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 3307, cfg.Server.Port)
	assert.Equal(t, 500, cfg.Server.MaxConnections)
	assert.Equal(t, "/tmp/test", cfg.Storage.DataDir)
	assert.False(t, cfg.Storage.SyncWrites)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
}

func TestConfig_LoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("GUOCEDB_HOST", "192.168.1.1")
	os.Setenv("GUOCEDB_PORT", "3308")
	os.Setenv("GUOCEDB_DATA_DIR", "/var/lib/guocedb")
	defer func() {
		os.Unsetenv("GUOCEDB_HOST")
		os.Unsetenv("GUOCEDB_PORT")
		os.Unsetenv("GUOCEDB_DATA_DIR")
	}()
	
	// Load config without file
	cfg, err := Load("")
	require.NoError(t, err)
	
	// Verify environment variables override defaults
	assert.Equal(t, "192.168.1.1", cfg.Server.Host)
	assert.Equal(t, 3308, cfg.Server.Port)
	assert.Equal(t, "/var/lib/guocedb", cfg.Storage.DataDir)
}

func TestConfig_Defaults(t *testing.T) {
	// Load config without file or env vars
	cfg, err := Load("")
	require.NoError(t, err)
	
	// Verify defaults
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 3306, cfg.Server.Port)
	assert.Equal(t, 1000, cfg.Server.MaxConnections)
	assert.Equal(t, "./data", cfg.Storage.DataDir)
	assert.True(t, cfg.Storage.SyncWrites)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "text", cfg.Log.Format)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 3306,
					ReadTimeout: "30s",
					WriteTimeout: "30s",
				},
				Storage: StorageConfig{
					DataDir: "/tmp/data",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: &Config{
				Server: ServerConfig{
					Port: 70000,
				},
				Storage: StorageConfig{
					DataDir: "/tmp/data",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "empty data dir",
			config: &Config{
				Server: ServerConfig{
					Port: 3306,
				},
				Storage: StorageConfig{
					DataDir: "",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: &Config{
				Server: ServerConfig{
					Port: 3306,
				},
				Storage: StorageConfig{
					DataDir: "/tmp/data",
				},
				Log: LogConfig{
					Level:  "invalid",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &Config{
				Server: ServerConfig{
					Port: 3306,
					ReadTimeout: "invalid",
				},
				Storage: StorageConfig{
					DataDir: "/tmp/data",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.yaml")
	
	// Write invalid YAML
	err := os.WriteFile(configFile, []byte("invalid: yaml: content: ["), 0644)
	require.NoError(t, err)
	
	// Should fail to load
	_, err = Load(configFile)
	assert.Error(t, err)
}

func TestConfig_NonexistentFile(t *testing.T) {
	// Should fail to load nonexistent file
	_, err := Load("/nonexistent/config.yaml")
	assert.Error(t, err)
}