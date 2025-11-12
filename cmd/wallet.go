package cmd

import (
	"log"
	"path/filepath"
	"strconv"

	"github.com/palagend/cipherkey/internal/security"
	"github.com/palagend/cipherkey/internal/service"
	"github.com/palagend/cipherkey/internal/view"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [strength]",
	Short: "Create a new secure HD wallet",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		strength := 128
		if len(args) > 0 {
			if s, err := strconv.Atoi(args[0]); err == nil {
				strength = s
			}
		}

		walletService := service.NewWalletService()
		wallet, err := walletService.CreateHDWallet(coinSymbol, strength, "", 1)
		if err != nil {
			log.Fatalf("Failed to create wallet: %v\n", err)
		}
		wd, err := walletService.EncryptWallet(wallet, password)
		if err != nil {
			log.Fatalf("Failed to encrypt wallet: %v\n", err)
		}
		if err := security.WriteSecureFile(filepath.Join(keystoreDir, wallet.ID), wd); err != nil {
			log.Fatalf("Failed to save wallet keystore: %v\n", err)
		}

		view.RenderWallet(wallet, "")
	},
}
