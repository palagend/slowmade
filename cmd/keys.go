package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/palagend/cipherkey/internal/service"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Key and cryptographic utilities",
	Long:  "Cryptographic key generation, conversion, and validation tools",
}

var generateMnemonicKeysCmd = &cobra.Command{
	Use:   "generate-mnemonic [strength]",
	Short: "Generate a BIP39 mnemonic phrase",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		strength := 128
		if len(args) > 0 {
			if s, err := strconv.Atoi(args[0]); err == nil {
				strength = s
			}
		}

		keyService := service.NewKeyService()
		mnemonic, err := keyService.GenerateMnemonic(strength)
		if err != nil {
			log.Fatalf("Failed to generate mnemonic: %v", err)
		}

		fmt.Println("Generated mnemonic phrase:")
		fmt.Printf("\n%s\n\n", mnemonic)
		fmt.Println("Store this phrase in a secure location!")
	},
}

var validateMnemonicKeysCmd = &cobra.Command{
	Use:   "validate-mnemonic <mnemonic>",
	Short: "Validate a BIP39 mnemonic phrase",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mnemonic := args[0]

		keyService := service.NewKeyService()
		isValid := keyService.ValidateMnemonic(mnemonic)

		if isValid {
			fmt.Println("Mnemonic phrase is valid")
		} else {
			fmt.Println("Mnemonic phrase is invalid")
		}
	},
}

var seedFromMnemonicKeysCmd = &cobra.Command{
	Use:   "seed-from-mnemonic <mnemonic>",
	Short: "Derive seed from mnemonic phrase",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mnemonic := args[0]

		keyService := service.NewKeyService()
		seed, err := keyService.SeedFromMnemonic(mnemonic, passphrase)
		if err != nil {
			log.Fatalf("Failed to derive seed: %v", err)
		}

		fmt.Printf("Seed (hex): %x\n", seed)
	},
}

func init() {
	keysCmd.AddCommand(generateMnemonicKeysCmd, validateMnemonicKeysCmd, seedFromMnemonicKeysCmd)
}
