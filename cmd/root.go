package cmd

import (
	"fmt"
	"os"

	"github.com/palagend/slowmade/internal/config"
	"github.com/palagend/slowmade/internal/logging"
	"github.com/palagend/slowmade/internal/mvc/controllers"
	"github.com/palagend/slowmade/internal/mvc/services"
	"github.com/palagend/slowmade/internal/mvc/views"
	"github.com/palagend/slowmade/internal/storage"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile          string
	appConfig        *config.AppConfig
	walletController *controllers.WalletController
)

var rootCmd = &cobra.Command{
	Use:   "slowmade",
	Short: "A secure cryptocurrency wallet CLI",
	Long: `A secure HD cryptocurrency wallet supporting multiple currencies,
built with MVC architecture and enterprise-grade security.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logging.Error("命令执行失败", map[string]interface{}{"error": err})
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "~/.config/slowmade.yaml", "config file")
	rootCmd.PersistentFlags().String("template-dir", "", "custom template directory")
	rootCmd.PersistentFlags().StringP("keystore-dir", "k", "~/.slowmade/keystores", "keystore directory")
	// 绑定环境变量
	viper.SetEnvPrefix("SLOWMADE")
	viper.BindPFlag("template.custom_template_dir", rootCmd.PersistentFlags().Lookup("template-dir"))
	viper.BindPFlag("storage.keystore_dir", rootCmd.PersistentFlags().Lookup("keystore-dir"))
}

func initializeConfig() error {
	var err error
	appConfig, err = config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	// 初始化全局日志系统
	if err := logging.Initialize(&appConfig.Logging); err != nil {
		return fmt.Errorf("初始化日志系统失败: %v", err)
	}

	logging.Debug("配置加载成功", map[string]interface{}{
		"config_file": cfgFile,
		"log_level":   appConfig.Logging.Level,
	})

	initializeDependencies()
	return nil
}

func initializeDependencies() {
	logging.Debug("开始初始化依赖项")

	// 初始化存储
	repo := storage.NewWalletRepository(appConfig)
	logging.Debug("存储仓库初始化完成")

	// 初始化服务
	cryptoService := services.NewCryptoService()
	walletService := services.NewWalletService(repo, cryptoService)
	logging.Debug("服务层初始化完成")

	// 初始化视图
	renderer := views.NewTemplateRenderer(&appConfig.Template)
	logging.Debug("视图渲染器初始化完成")

	// 打印模板状态（调试用）
	renderer.PrintStatus()

	// 初始化控制器
	walletController = controllers.NewWalletController(walletService, renderer)
	logging.Debug("所有依赖项初始化完成")
}
