package cmd

import (
	"fmt"
	"os"

	"github.com/palagend/slowmade/internal/config"
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
	Use:   "wallet-cli",
	Short: "A secure cryptocurrency wallet CLI",
	Long: `A secure HD cryptocurrency wallet supporting multiple currencies,
built with MVC architecture and enterprise-grade security.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/slowmade.yaml)")
	// 绑定环境变量
	viper.SetEnvPrefix("SLOWMADE")
	viper.BindPFlag("template.custom_template_dir", rootCmd.PersistentFlags().Lookup("template-dir"))
}

func initializeConfig() error {
	var err error
	appConfig, err = config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	initializeDependencies()
	return nil
}

func initializeDependencies() {
	// 初始化存储
	repo := storage.NewWalletRepository(appConfig)

	// 初始化服务
	cryptoService := services.NewCryptoService()
	walletService := services.NewWalletService(repo, cryptoService)

	// 初始化视图
	renderer := views.NewTemplateRenderer(&appConfig.Template)

	// 打印模板状态（调试用）
	renderer.PrintStatus()

	// 初始化控制器[2](@ref)
	walletController = controllers.NewWalletController(walletService, renderer)
}
