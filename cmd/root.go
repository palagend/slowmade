package cmd

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/palagend/slowmade/internal/app"
	"github.com/palagend/slowmade/internal/config"
	"github.com/palagend/slowmade/internal/core"
	"github.com/palagend/slowmade/pkg/crypto"
	"github.com/palagend/slowmade/pkg/logging"
	"github.com/palagend/slowmade/pkg/mnemonic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	debug     bool
	dataDir   string
	cloak     string
	walletMgr core.WalletManager // 改为接口类型而非具体实现
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
		replApp, err := app.NewREPL(walletMgr)
		if err != nil {
			fmt.Printf("Error creating REPL: %v\n", err)
			os.Exit(1)
		}
		replApp.Run()
	},
}

func initDependencies() {
	// 创建 WalletManager 实例（具体实现）
	stor, err := core.NewFileStorage(dataDir)
	if err != nil {
		log.Error(err.Error())
	}
	cs := crypto.NewAESCryptoService([]byte(cloak))
	ms := mnemonic.NewBIP39MnemonicService()
	walletMgr = core.NewDefaultWalletManager(stor, cs, ms)

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
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "", "storage base directory")
	rootCmd.PersistentFlags().StringVar(&cloak, "cloak", "", "cloak")

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

	if debug {
		viper.Set("log.level", "debug")
	}

	if err := config.Load(); err != nil {
		fmt.Printf("Failed to initialize config: %v\n", err)
		os.Exit(1)
	}
}
