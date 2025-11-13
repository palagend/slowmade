package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/palagend/slowmade/internal/mvc/models"
	"github.com/palagend/slowmade/internal/storage"
	"github.com/tyler-smith/go-bip39"
)

type WalletService struct {
	walletRepo    *storage.WalletRepository
	cryptoService *CryptoService
}

func NewWalletService(repo *storage.WalletRepository, cryptoService *CryptoService) *WalletService {
	return &WalletService{
		walletRepo:    repo,
		cryptoService: cryptoService,
	}
}

func (s *WalletService) CreateHDWallet(name, password string) (*models.VirtualWallet, error) {
	// 生成助记词
	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)

	// 创建钱包实例
	wallet := &models.VirtualWallet{
		ID:           generateID(),
		Name:         name,
		Mnemonic:     mnemonic,
		CreationTime: time.Now(),
		Coins:        []models.Coin{},
	}

	// 加密并保存
	encryptedWallet, err := s.cryptoService.EncryptWallet(wallet, password)
	if err != nil {
		return nil, err
	}

	if err := s.walletRepo.Save(encryptedWallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *WalletService) ListWallets() ([]*models.VirtualWallet, error) {
	return s.walletRepo.LoadAll()
}

func (s *WalletService) GetWallet(id string) (*models.VirtualWallet, error) {
	return s.walletRepo.Load(id)
}

func (s *WalletService) RestoreWallet(mnemonic, name, password string) error {
	// 验证助记词
	if !bip39.IsMnemonicValid(mnemonic) {
		return errors.New("invalid mnemonic phrase")
	}

	// 从助记词恢复钱包
	wallet := &models.VirtualWallet{
		ID:           generateID(),
		Name:         name,
		Mnemonic:     mnemonic,
		CreationTime: time.Now(),
	}

	encryptedWallet, err := s.cryptoService.EncryptWallet(wallet, password)
	if err != nil {
		return err
	}

	return s.walletRepo.Save(encryptedWallet)
}

// 辅助函数
func generateID() string {
	return uuid.New().String()
}
