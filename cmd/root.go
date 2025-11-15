package cmd

import (
	"fmt"
	"os"

	"github.com/palagend/slowmade/internal/app"
	"github.com/palagend/slowmade/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cfgFile string
	debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "slowmade",
	Short: "A secure cryptocurrency wallet",
	Long:  `Slowmade is a secure HD wallet supporting multiple cryptocurrencies with REPL interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 进入 REPL 模式
		replApp, err := app.NewREPL()
		if err != nil {
			fmt.Printf("Error creating REPL: %v\n", err)
			os.Exit(1)
		}
		replApp.Run()
	},
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		logging.Get().Error("Command execution failed", zap.Error(err))
		return err
	}
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.toml)")
	rootCmd.PersistentFlags().String("lang", "en", "language preference (en/zh/ja)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")

	viper.BindPFlag("lang", rootCmd.PersistentFlags().Lookup("lang"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		createDefaultConfig()
	}

	logConfig := logging.Config{
		Level:    viper.GetString("log.level"),
		Encoding: "json",
	}
	if debug {
		logConfig.Level = "debug"
	}
	if err := logging.Init(logConfig); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logging.Get().Info("Configuration loaded",
		zap.String("config", viper.ConfigFileUsed()))
}

func createDefaultConfig() {
	viper.SetDefault("rpc.endpoint", "http://localhost:8545")
	viper.SetDefault("rpc.timeout", 30)
	viper.SetDefault("keystore.path", "./keystore")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("ui.lang", "en")

	if err := viper.SafeWriteConfig(); err != nil {
		fmt.Printf("Warning: Failed to create config file: %v\n", err)
	}
}
