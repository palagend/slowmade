package view

import (
	"embed"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/palagend/cipherkey/internal/model"
)

type WalletTemplateData struct {
	Wallet       *model.HDWallet
	Timestamp    string
	TotalBalance uint64
	QRCodeASCII  string
}

//go:embed templates/wallet
var walletTemplateFS embed.FS

func RenderWallet(wallet *model.HDWallet, qrASCII string) {
	var totalBalance uint64
	for _, acc := range wallet.Accounts {
		totalBalance += acc.Balance
	}

	data := WalletTemplateData{
		Wallet:       wallet,
		Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		TotalBalance: totalBalance,
		QRCodeASCII:  qrASCII,
	}
	engine := NewTemplateRenderer(false)
	engine.AddFuncMap(template.FuncMap{
		"title": func(s string) string {
			border := repeat("=", len(s)+4)
			return "\n" + border + "\n  " + s + "\n" + border
		},
		"bold":    func(s string) string { return "\033[1m" + s + "\033[0m" },
		"divider": func() string { return repeat("-", 60) },
		"shorten": func(s string, length int) string {
			if len(s) > length {
				return s[:length] + "..."
			}
			return s
		},
		"repeat": repeat,
	})

	err := engine.LoadTemplates(walletTemplateFS, "*.tmpl")
	if err != nil {
		log.Fatal("Failed to load templates:", err)
	}
	engine.Execute(os.Stdout, "wallet", data)
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
