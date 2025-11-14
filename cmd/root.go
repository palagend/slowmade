package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/palagend/slowmade/internal/config"
	"github.com/palagend/slowmade/internal/logging"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	appConfig *config.AppConfig
)

var rootCmd = &cobra.Command{
	Use:   "slowmade",
	Short: "A secure cryptocurrency wallet CLI",
	Long:  `A secure HD cryptocurrency wallet supporting multiple currencies,`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// 此时日志系统可能还未初始化，使用标准错误输出
		fmt.Fprintf(os.Stderr, "命令执行失败: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// 1. 首先定义命令行标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认搜索标准配置目录)")
	rootCmd.PersistentFlags().String("template-dir", "", "自定义模板目录")
	rootCmd.PersistentFlags().Bool("enable-template", false, "是否启用自定义模板")
	rootCmd.PersistentFlags().String("log-level", "", "日志级别")
	rootCmd.PersistentFlags().StringP("keystore-dir", "d", "", "密钥存储目录")

	// 2. 将初始化函数注册到OnInitialize
	cobra.OnInitialize(initConfig)
}

// initConfig 初始化配置系统
func initConfig() {
	// 设置环境变量
	viper.SetEnvPrefix("SLOWMADE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv() // 自动读取所有环境变量

	// 绑定命令行标志到Viper
	viper.BindPFlag("template.custom_template_dir", rootCmd.PersistentFlags().Lookup("template-dir"))
	viper.BindPFlag("template.enable_custom_templates", rootCmd.PersistentFlags().Lookup("enable-template"))
	viper.BindPFlag("storage.keystore_dir", rootCmd.PersistentFlags().Lookup("keystore-dir"))
	viper.BindPFlag("logging.level", rootCmd.PersistentFlags().Lookup("log-level"))

	// 加载配置
	var err error
	appConfig, err = config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	if err := logging.Initialize(&appConfig.Logging); err != nil {
		fmt.Fprintf(os.Stderr, "日志系统初始化失败: %v\n", err)
		os.Exit(1)
	}

	logging.Info("应用初始化完成", map[string]interface{}{
		"config_file": viper.ConfigFileUsed(),
		"log_level":   appConfig.Logging.Level,
	})
}

// GetConfig 提供获取配置的公共方法
func GetConfig() *config.AppConfig {
	return appConfig
}
