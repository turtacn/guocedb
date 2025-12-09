package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Loader handles configuration loading from multiple sources.
type Loader struct {
	v *viper.Viper
}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	v := viper.New()
	v.SetEnvPrefix("GUOCEDB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
	
	// Bind all possible env vars so viper knows to look for them
	v.BindEnv("server.host")
	v.BindEnv("server.port")
	v.BindEnv("server.max_connections")
	v.BindEnv("storage.data_dir")
	v.BindEnv("storage.sync_writes")
	v.BindEnv("security.enabled")
	v.BindEnv("logging.level")
	v.BindEnv("logging.format")
	
	return &Loader{v: v}
}

// Load loads configuration from file and environment variables.
func (l *Loader) Load(configPath string) (*Config, error) {
	// Load from config file if specified
	if configPath != "" {
		if err := l.loadFile(configPath); err != nil {
			return nil, fmt.Errorf("load config file: %w", err)
		}
	} else {
		// Try default locations
		l.tryDefaultLocations()
	}

	// Start with an empty config
	cfg := &Config{}

	// Unmarshal to config struct (environment variables override file)
	if err := l.v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Fill in defaults for any missing values
	cfg.ApplyDefaults()

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadFile loads a specific config file.
func (l *Loader) loadFile(path string) error {
	l.v.SetConfigFile(path)
	return l.v.ReadInConfig()
}

// tryDefaultLocations attempts to load config from default locations.
func (l *Loader) tryDefaultLocations() {
	l.v.SetConfigName("guocedb")
	l.v.SetConfigType("yaml")
	l.v.AddConfigPath(".")
	l.v.AddConfigPath("./configs")
	l.v.AddConfigPath("/etc/guocedb")
	l.v.AddConfigPath("$HOME/.guocedb")
	// Ignore errors - use defaults if no config file found
	l.v.ReadInConfig()
}

// BindFlags binds command-line flags to viper keys.
func (l *Loader) BindFlags(flags *pflag.FlagSet) {
	if flags == nil {
		return
	}

	// Bind flags if they exist
	if f := flags.Lookup("port"); f != nil {
		l.v.BindPFlag("server.port", f)
	}
	if f := flags.Lookup("host"); f != nil {
		l.v.BindPFlag("server.host", f)
	}
	if f := flags.Lookup("data-dir"); f != nil {
		l.v.BindPFlag("storage.data_dir", f)
	}
	if f := flags.Lookup("log-level"); f != nil {
		l.v.BindPFlag("logging.level", f)
	}
	if f := flags.Lookup("auth"); f != nil {
		l.v.BindPFlag("security.enabled", f)
	}
	if f := flags.Lookup("metrics"); f != nil {
		l.v.BindPFlag("observability.enabled", f)
	}
}

// LoadWithFlags loads configuration and applies command-line flags.
func LoadWithFlags(configPath string, flags *pflag.FlagSet) (*Config, error) {
	loader := NewLoader()
	loader.BindFlags(flags)
	return loader.Load(configPath)
}
