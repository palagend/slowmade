package cmd

import (
	"github.com/palagend/cipherkey/internal/model/coin"
	"github.com/palagend/cipherkey/internal/view"
	"github.com/spf13/cobra"
)

var listCoinsCmd = &cobra.Command{
	Use:   "list-coins",
	Short: "List all supported cryptocurrencies",
	Run: func(cmd *cobra.Command, args []string) {
		factory := coin.GetInstance()
		coins := factory.GetAllCoins()

		view.RenderCoinList(coins)
	},
}
