package config

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/turtacn/guocedb/common/constants"
)

// Config holds the entire configuration for the application.
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Storage StorageConfig `mapstructure:"storage"`
	Security SecurityConfig `mapstructure:"security"`
	Log     LogConfig     `mapstructure:"log"`
}

// ServerConfig holds server-related configurations.
type ServerConfig struct {
	MySQLPort int    `mapstructure:"mysql_port"`
	GRPCPort  int    `mapstructure:"grpc_port"`
	HTTPPort  int    `mapstructure:"http_port"`
	Timeout   string `mapstructure:"timeout"`
}

// StorageConfig holds storage-related configurations.
type StorageConfig struct {
	Engine      string `mapstructure:"engine"`
	Path        string `mapstructure:"path"`
	CacheSizeMB int    `mapstructure:"cache_size_mb"`
}

// SecurityConfig holds security-related configurations.
type SecurityConfig struct {
	TLSCertPath string `mapstructure:"tls_cert_path"`
	TLSKeyPath  string `mapstructure:"tls_key_path"`
}

// LogConfig holds logging-related configurations.
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// GlobalConfig holds the loaded configuration and is accessible globally.
var GlobalConfig *Config

// LoadConfig loads the configuration from a file and environment variables.
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)   // path to look for the config file in
	viper.AddConfigPath(".")    // also look in the working directory
	viper.AddConfigPath("/etc/guocedb/") // and a standard system path

	// Set default values from the constants package
	viper.SetDefault("server.mysql_port", constants.DefaultMySQLPort)
	viper.SetDefault("server.grpc_port", constants.DefaultGRPCPort)
	viper.SetDefault("server.http_port", constants.DefaultHTTPPort)
	viper.SetDefault("server.timeout", constants.DefaultTimeout)
	viper.SetDefault("storage.engine", constants.StorageEngineBadger)
	viper.SetDefault("storage.path", "./guocedb-data")
	viper.SetDefault("storage.cache_size_mb", constants.DefaultCacheSizeMB)
	viper.SetDefault("log.level", "info")

	// Enable environment variable overriding, e.g., GUOCEDB_SERVER_MYSQL_PORT=3307
	viper.SetEnvPrefix("GUOCEDB")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Attempt to read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error, will use defaults/env vars
		} else {
			// Config file was found but another error was produced
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	GlobalConfig = &config
	return &config, nil
}

// TODO: Implement configuration hot-reloading using viper.WatchConfig().
