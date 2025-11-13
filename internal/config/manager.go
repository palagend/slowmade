// internal/config/manager.go
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func LoadConfig(configPath string) (*AppConfig, error) {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// 搜索标准位置
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/")
		viper.AddConfigPath("/etc/slowmade/")
		viper.SetConfigName("slowmade.yaml")
	}

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// 使用默认配置
		return getDefaultConfig(), nil
	}

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// 处理路径中的波浪号
	config.Template.CustomTemplateDir = expandPath(config.Template.CustomTemplateDir)
	config.Storage.KeystoreDir = expandPath(config.Storage.KeystoreDir)

	return &config, nil
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}

func getDefaultConfig() *AppConfig {
	return &AppConfig{
		Template: TemplateConfig{
			CustomTemplateDir: "~/.slowmade/templates",
			EnableCustom:      true,
		},
		Storage: StorageConfig{
			KeystoreDir: "~/.slowmade/keystores",
			Encryption: EncryptionConfig{
				Algorithm: "scrypt",
				Cost:      16384,
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}
