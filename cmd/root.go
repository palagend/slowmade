package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/palagend/slowmade/internal/app"
	"github.com/palagend/slowmade/internal/config"
	"github.com/palagend/slowmade/internal/core"
	"github.com/palagend/slowmade/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	debug      bool
	cloak      string
	walletMgr  core.WalletManager
	accountMgr core.AccountManager
)

var rootCmd = &cobra.Command{
	Use:   "slowmade",
	Short: "A secure cryptocurrency wallet",
	Long:  `Slowmade is a secure HD wallet supporting multiple cryptocurrencies with REPL interface.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initDependencies()
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 进入 REPL 模式
		replApp, err := app.NewREPL(walletMgr, accountMgr)
		if err != nil {
			fmt.Printf("Error creating REPL: %v\n", err)
			os.Exit(1)
		}
		replApp.Run()
	},
}

func initDependencies() {
	// 创建 WalletManager 实例（具体实现）
	appConfig := config.GetAppConfig()
	if debug {
		appConfigStr, _ := json.MarshalIndent(appConfig, "", "  ")
		logging.Debugf("AppConfig is: %s", appConfigStr)
	}
	stor, err := core.NewFileStorage(appConfig.GetStorageConfig())
	if err != nil {
		log.Error(err.Error())
	}
	walletMgr = core.NewDefaultWalletManager(stor, cloak)
	accountMgr = core.NewDefaultAccountManager(walletMgr, stor)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logging.Get().Error("Command execution failed", zap.Error(err))
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "config file")
	rootCmd.PersistentFlags().String("lang", "en", "language preference (en/zh/ja)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug")
	rootCmd.PersistentFlags().String("data-dir", "", "storage base directory")
	rootCmd.PersistentFlags().StringVar(&cloak, "cloak", "", "Advanced feature: a cloak provides optional added security, but it is not stored so it must be remembered!")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// 正确绑定配置标志
	if err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		fmt.Printf("Failed to bind config flag: %v\n", err)
	}
	if err := viper.BindPFlag("ui.lang", rootCmd.PersistentFlags().Lookup("lang")); err != nil {
		fmt.Printf("Failed to bind lang flag: %v\n", err)
	}

	if err := viper.BindPFlag("storage.base_dir", rootCmd.PersistentFlags().Lookup("data-dir")); err != nil {
		fmt.Printf("Failed to bind data-dir flag: %v\n", err)
	}

	if debug {
		viper.Set("log.level", "debug")
	}

	if err := config.Load(); err != nil {
		fmt.Printf("Failed to initialize config: %v\n", err)
		os.Exit(1)
	}
}
