// cmd/create.go
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	walletName string
	password   string
	cloak      string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new HD wallet",
	Long:  `Create a new hierarchical deterministic (HD) cryptocurrency wallet`,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := walletController.CreateWallet(walletName, password)
		if err != nil {
			return err
		}

		cmd.Println(result)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&walletName, "name", "n", "", "Wallet name (required)")
	createCmd.Flags().StringVarP(&password, "password", "p", "", "Encryption password (required)")
	createCmd.Flags().StringVarP(&cloak, "cloak", "k", "", "Magic password for Mnemonic.")

	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(createCmd)
}
