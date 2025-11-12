package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	configPath   string
	coinSymbol   string
	outputFormat string
	verbose      bool
	keystoreDir  string
	password     string
	passphrase   string
)

var rootCmd = &cobra.Command{
	Use:   "cipherkey",
	Short: "A secure HD cryptocurrency wallet CLI",
	Long: `CipherKey is a modern hierarchical deterministic cryptocurrency wallet 
with multi-coin support, customizable templates, and enterprise-grade security.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	home, _ := os.UserHomeDir()
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Config file path")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "text", "Output format")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password to open keystore")

	rootCmd.PersistentFlags().StringVarP(&coinSymbol, "coin", "C", "btc", "Coin type symbol")
	rootCmd.PersistentFlags().StringVarP(&passphrase, "passphrase", "P", "", "Passphrase for seed derivation")
	rootCmd.PersistentFlags().StringVarP(&keystoreDir, "keystore", "K", filepath.Join(home, ".cipherkey", "keystore"), "The directory to save keystore file")

	// 注册子命令
	rootCmd.AddCommand(createCmd, listCoinsCmd, qrcodeCmd, keysCmd)
}
