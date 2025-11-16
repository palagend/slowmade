package cmd

import (
	"fmt"
	"os"

	"github.com/palagend/slowmade/internal/app"
	"github.com/palagend/slowmade/internal/config"
	"github.com/palagend/slowmade/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	rootCmd.PersistentFlags().String("config", "", "config file")
	rootCmd.PersistentFlags().String("lang", "en", "language preference (en/zh/ja)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug")

	cobra.OnInitialize(initConfig)
}

var debug bool

func initConfig() {
	viper.BindPFlag("config", rootCmd.Flags().Lookup("config"))
	viper.BindPFlag("ui.lang", rootCmd.Flags().Lookup("lang"))
	if debug {
		viper.GetViper().Set("log.level", "debug")
	}

	if _, err := config.Load(); err != nil {
		fmt.Printf("Failed to initialize config: %v\n", err)
		os.Exit(1)
	}
}
