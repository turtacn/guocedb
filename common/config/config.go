// Package config handles loading and managing configuration for guocedb.
package config

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/turtacn/guocedb/common/constants"
	"github.com/turtacn/guocedb/common/log"
)

// Config holds the entire configuration for the application.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Security  SecurityConfig  `mapstructure:"security"`
	Log       LogConfig       `mapstructure:"log"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	GRPCPort      int    `mapstructure:"grpcPort"`
	MaxConnections int    `mapstructure:"maxConnections"`
	Timeout       int    `mapstructure:"timeout"`
}

// StorageConfig holds storage-related configuration.
type StorageConfig struct {
	Engine    string        `mapstructure:"engine"`
	DataDir   string        `mapstructure:"dataDir"`
	Badger    BadgerConfig  `mapstructure:"badger"`
}

// BadgerConfig holds configuration specific to the Badger storage engine.
type BadgerConfig struct {
	ValueLogFileSize int `mapstructure:"valueLogFileSize"`
	SyncWrites       bool `mapstructure:"syncWrites"`
}

// SecurityConfig holds security-related configuration.
type SecurityConfig struct {
	EnableTLS bool   `mapstructure:"enableTls"`
	CertFile  string `mapstructure:"certFile"`
	KeyFile   string `mapstructure:"keyFile"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// MetricsConfig holds metrics-related configuration.
type MetricsConfig struct {
	Enable bool `mapstructure:"enable"`
	Port   int  `mapstructure:"port"`
}

var globalConfig *Config

// Load initializes the configuration from a file, environment variables, and defaults.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", constants.DefaultPort)
	v.SetDefault("server.grpcPort", constants.DefaultGRPCPort)
	v.SetDefault("server.maxConnections", constants.DefaultMaxConnections)
	v.SetDefault("server.timeout", constants.DefaultTimeout)
	v.SetDefault("storage.engine", constants.StorageEngineBadger)
	v.SetDefault("storage.dataDir", "/var/lib/guocedb")
	v.SetDefault("storage.badger.valueLogFileSize", 1<<30) // 1GB
	v.SetDefault("storage.badger.syncWrites", true)
	v.SetDefault("security.enableTls", false)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("metrics.enable", true)
	v.SetDefault("metrics.port", constants.DefaultMetricsPort)

	// Load from config file
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	// Load from environment variables
	v.SetEnvPrefix("GUOCEDB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal the config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// TODO: Implement hot-reloading
	// v.WatchConfig()
	// v.OnConfigChange(func(e fsnotify.Event) {
	// 	log.GetLogger().Infof("Config file changed: %s", e.Name)
	// 	// Reload config logic here
	// })

	globalConfig = &cfg
	return globalConfig, nil
}

// Get returns the global configuration instance.
func Get() *Config {
	if globalConfig == nil {
		// This might happen if Get is called before Load.
		// For robustness, we could load with default path or panic.
		log.GetLogger().Warn("Configuration not loaded, using defaults.")
		cfg, err := Load("")
		if err != nil {
			panic("failed to load default configuration")
		}
		return cfg
	}
	return globalConfig
}
