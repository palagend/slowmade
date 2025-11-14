package controllers

import (
	"github.com/palagend/slowmade/internal/mvc/services"
	"github.com/palagend/slowmade/internal/mvc/views"
)

type WalletController struct {
	walletService *services.WalletService
	viewRenderer  *views.TemplateRenderer
}

func NewWalletController(service *services.WalletService, renderer *views.TemplateRenderer) *WalletController {
	return &WalletController{
		walletService: service,
		viewRenderer:  renderer,
	}
}

func (c *WalletController) CreateWallet(name, password, cloak string) (string, error) {
	wallet, err := c.walletService.CreateHDWallet(name, password, cloak)
	if err != nil {
		return "", err
	}

	// 使用模板渲染结果
	return c.viewRenderer.RenderWalletCreated(wallet)
}

func (c *WalletController) RestoreWallet(mnemonic, name, password string) error {
	return c.walletService.RestoreWallet(mnemonic, name, password)
}

func (c *WalletController) ListWallets() (string, error) {
	wallets, err := c.walletService.ListWallets()
	if err != nil {
		return "", err
	}

	return c.viewRenderer.RenderWalletList(wallets)
}

func (c *WalletController) GetWalletInfo(walletID string) (string, error) {
	wallet, err := c.walletService.GetWallet(walletID)
	if err != nil {
		return "", err
	}

	return c.viewRenderer.RenderWalletInfo(wallet)
}
