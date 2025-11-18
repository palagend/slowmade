// internal/config/config.go
package config

import (
	"fmt"
	"strings"

	"github.com/palagend/slowmade/pkg/logging"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// AppConfig 完整的应用配置结构
type AppConfig struct {
	RPC     RPCConfig     `mapstructure:"rpc"`
	Storage StorageConfig `mapstructure:"storage"`
	Log     LogConfig     `mapstructure:"log"`
	UI      UIConfig      `mapstructure:"ui"`
	Web     WebConfig     `mapstructure:"web"`
}

type RPCConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Timeout  int    `mapstructure:"timeout"`
}

type StorageConfig struct {
	BaseDir string `mapstructure:"base_dir"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	File     string `mapstructure:"file"`
	Encoding string `mapstructure:"encoding"`
}

type UIConfig struct {
	Lang string `mapstructure:"lang"`
}

type WebConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// Load 加载配置并初始化日志
// 配置优先级: 命令行参数 > 环境变量 > 配置文件 > 默认值
func Load() error {
	v := viper.GetViper()

	// 1. 首先设置默认值
	setDefaults(v)

	// 2. 绑定环境变量（在读取配置文件之前）
	bindEnvironmentVariables(v)

	// 3. 读取配置文件
	if err := setupConfigFile(v); err != nil {
		return err
	}

	// 5. 自动读取环境变量（覆盖配置文件中的值）
	v.AutomaticEnv()

	// 6. 反序列化到结构体
	if err := v.Unmarshal(&appConfig); err != nil {
		return fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// 7. 初始化日志系统
	if err := setupLogging(appConfig.Log); err != nil {
		return err
	}

	// 记录配置加载信息
	logConfigSources(v)

	return nil
}

// setDefaults 设置所有配置的默认值
func setDefaults(v *viper.Viper) {
	// RPC 配置默认值
	v.SetDefault("rpc.endpoint", "http://localhost:8545")
	v.SetDefault("rpc.timeout", 30)

	// Keystore 配置默认值
	v.SetDefault("keystore.path", "./keystore")

	// 日志配置默认值
	v.SetDefault("log.level", "info")
	v.SetDefault("log.encoding", "console")
	v.SetDefault("log.file", "")

	// UI 配置默认值
	v.SetDefault("ui.lang", "en")
}

// bindEnvironmentVariables 绑定环境变量映射
func bindEnvironmentVariables(v *viper.Viper) {
	// 设置环境变量前缀（可选）
	v.SetEnvPrefix("SLOWMADE")

	// 自动将点号分隔的键转换为下划线（环境变量标准）
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 显式绑定关键环境变量（确保正确的映射关系）
	v.BindEnv("rpc.endpoint")  // 对应 SLOWMADE_RPC_ENDPOINT
	v.BindEnv("rpc.timeout")   // 对应 SLOWMADE_RPC_TIMEOUT
	v.BindEnv("keystore.path") // 对应 SLOWMADE_KEYSTORE_PATH
	v.BindEnv("log.level")     // 对应 SLOWMADE_LOG_LEVEL
	v.BindEnv("log.file")      // 对应 SLOWMADE_LOG_FILE
	v.BindEnv("log.encoding")  // 对应 SLOWMADE_LOG_ENCODING
	v.BindEnv("ui.lang")       // 对应 SLOWMADE_UI_LANG
}

// setupConfigFile 设置和读取配置文件
func setupConfigFile(v *viper.Viper) error {
	if v.GetString("config") != "" {
		v.SetConfigFile(v.GetString("config"))
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.slowmade")
		v.AddConfigPath("/etc/slowmade/")
	}
	// 尝试读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// 如果是配置文件找到但解析错误，返回错误
			return fmt.Errorf("config file found but unable to read: %w", err)
		}
		// 配置文件不存在是正常的，使用默认值+环境变量+命令行参数
	}

	return nil
}

// setupLogging 初始化日志系统
func setupLogging(logConfig LogConfig) error {
	config := logging.Config{
		Level:    logConfig.Level,
		Encoding: logConfig.Encoding,
		File:     logConfig.File,
	}

	if err := logging.Init(config); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	return nil
}

// logConfigSources 记录配置加载的来源信息
func logConfigSources(v *viper.Viper) {
	logger := logging.Get()

	// 记录使用的配置文件
	if v.ConfigFileUsed() != "" {
		logger.Info("Configuration loaded from file",
			zap.String("file", v.ConfigFileUsed()))
	} else {
		logger.Info("Using default configuration with environment variables and command line flags")
	}

	// 记录重要的配置值（敏感信息需要脱敏）
	logger.Debug("Configuration values",
		zap.String("rpc.endpoint", v.GetString("rpc.endpoint")),
		zap.Int("rpc.timeout", v.GetInt("rpc.timeout")),
		zap.String("log.level", v.GetString("log.level")),
		zap.String("ui.lang", v.GetString("ui.lang")),
	)
}

// 辅助函数
// GetWebConfig 返回Web相关的配置，供web模块使用
func (c *AppConfig) GetWebConfig() WebConfig {
	return c.Web
}

// GetRPCConfig 返回RPC相关的配置，供需要与链交互的模块使用
func (c *AppConfig) GetRPCConfig() RPCConfig {
	return c.RPC
}

// GetKeystoreConfig 返回Keystore相关的配置，供账户管理模块使用
func (c *AppConfig) GetStorageConfig() StorageConfig {
	return c.Storage
}

// GetLogConfig 返回日志相关的配置
func (c *AppConfig) GetLogConfig() LogConfig {
	return c.Log
}

// GetUIConfig 返回用户界面相关的配置，供UI模块使用
func (c *AppConfig) GetUIConfig() UIConfig {
	return c.UI
}

var appConfig AppConfig

func GetAppConfig() AppConfig {
	return appConfig
}
