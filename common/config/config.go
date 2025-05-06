// Package config 提供配置加载与管理功能 / Configuration loading and management
package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// Config 定义全局配置结构体 / Config defines global configuration structure
type Config struct {
	Server struct {
		Port     int    `mapstructure:"port"`     // 监听端口 / Listen port
		LogLevel string `mapstructure:"log_level"` // 日志级别 / Log level
	} `mapstructure:"server"`
	Storage struct {
		Engine  string `mapstructure:"engine"`   // 存储引擎 (badger/mds/kv) / Storage engine (badger/mdi/kvd)
		DataDir string `mapstructure:"data_dir"` // 数据目录 / Data directory
	} `mapstructure:"storage"`
	Security struct {
		EnableAuth bool `mapstructure:"enable_auth"` // 是否启用认证 / Enable authentication
	} `mapstructure:"security"`
}

// LoadConfig 从指定路径加载配置文件 / Load configuration from given path
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err) // failed to read config file
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置结构失败: %w", err) // failed to unmarshal configuration
	}

	return &cfg, nil
}
