package view

import (
	"embed"
	"os"

	"github.com/palagend/cipherkey/internal/model/coin"
)

//go:embed templates/coin
var coinlistTemplateFS embed.FS

// Render 渲染币种列表
func RenderCoinList(coins []coin.Coin) {
	data := struct {
		Coins      []coin.Coin
		TotalCount int
	}{
		Coins:      coins,
		TotalCount: len(coins),
	}
	engine := NewTemplateRenderer(false)
	engine.LoadTemplates(coinlistTemplateFS, "*.tmpl")
	engine.Execute(os.Stdout, "coinlist", data)

}
