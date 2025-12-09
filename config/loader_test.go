package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestLoadFromYAML(t *testing.T) {
	yaml := `
server:
  host: "127.0.0.1"
  port: 13306
  max_connections: 500
storage:
  data_dir: "/var/lib/guocedb"
  sync_writes: true
security:
  enabled: true
  root_password: "secret123"
logging:
  level: "debug"
  format: "text"
`
	tmpFile := writeTempFile(t, yaml)
	defer os.Remove(tmpFile)

	cfg, err := LoadWithFlags(tmpFile, nil)
	require.NoError(t, err)

	require.Equal(t, "127.0.0.1", cfg.Server.Host)
	require.Equal(t, 13306, cfg.Server.Port)
	require.Equal(t, 500, cfg.Server.MaxConnections)
	require.Equal(t, "/var/lib/guocedb", cfg.Storage.DataDir)
	require.True(t, cfg.Storage.SyncWrites)
	require.True(t, cfg.Security.Enabled)
	require.Equal(t, "debug", cfg.Logging.Level)
}

func TestLoadFromEnv(t *testing.T) {
	envVars := map[string]string{
		"GUOCEDB_STORAGE_DATA_DIR":  "/tmp/guocedb",
		"GUOCEDB_SECURITY_ENABLED":  "true",
		"GUOCEDB_LOGGING_LEVEL":     "warn",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg, err := LoadWithFlags("", nil)
	require.NoError(t, err)

	require.Equal(t, "/tmp/guocedb", cfg.Storage.DataDir)
	require.True(t, cfg.Security.Enabled)
	require.Equal(t, "warn", cfg.Logging.Level)
}

func TestLoadPriority(t *testing.T) {
	// File: port=3306
	yaml := `server: { port: 3306 }`
	tmpFile := writeTempFile(t, yaml)
	defer os.Remove(tmpFile)

	// Environment variable: port=3307
	os.Setenv("GUOCEDB_SERVER_PORT", "3307")
	defer os.Unsetenv("GUOCEDB_SERVER_PORT")

	// Command line: port=3308
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.Int("port", 0, "")
	flags.Set("port", "3308")

	cfg, err := LoadWithFlags(tmpFile, flags)
	require.NoError(t, err)

	// Command line should have highest priority
	require.Equal(t, 3308, cfg.Server.Port)
}

func TestLoadDefaultLocations(t *testing.T) {
	// No config file specified
	cfg, err := LoadWithFlags("", nil)
	require.NoError(t, err)

	// Should get default values
	require.NotNil(t, cfg)
	require.Equal(t, 3306, cfg.Server.Port)
}

func TestLoadInvalidFile(t *testing.T) {
	cfg, err := LoadWithFlags("/nonexistent/config.yaml", nil)
	require.Error(t, err)
	require.Nil(t, cfg)
}

func TestLoadInvalidYAML(t *testing.T) {
	yaml := `
invalid: yaml: content:
  - missing
  structure
`
	tmpFile := writeTempFile(t, yaml)
	defer os.Remove(tmpFile)

	cfg, err := LoadWithFlags(tmpFile, nil)
	require.Error(t, err)
	require.Nil(t, cfg)
}

// Helper function to write temp file
func writeTempFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)
	return tmpFile
}
