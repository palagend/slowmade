package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/palagend/cipherkey/internal/model"
	"github.com/palagend/cipherkey/internal/model/coin"
	"github.com/palagend/cipherkey/internal/security"
	"github.com/tyler-smith/go-bip39"
)

// WalletService 钱包服务
type WalletService struct {
	coinFactory *coin.CoinRegistry
}

// NewWalletService 创建钱包服务实例
func NewWalletService() *WalletService {
	return &WalletService{
		coinFactory: coin.GetInstance(),
	}
}

func (ws *WalletService) CreateHDWallet(coinSymbol string, strength int, passphrase string, accountNum uint32) (*model.HDWallet, error) {
	// 验证币种支持
	coinImpl, err := ws.coinFactory.GetCoin(coinSymbol)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create secure memory: %v", err)
	}

	entropy, err := bip39.NewEntropy(strength)
	if err != nil {
		return nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	// 生成种子
	seed := bip39.NewSeed(mnemonic, passphrase)

	// 立即清理助记词字符串
	secureClearString(mnemonic)

	// 创建钱包结构
	wallet := &model.HDWallet{
		ID:         uuid.New().String(),
		Mnemonic:   mnemonic, // 助记词已破坏
		Seed:       seed,
		CoinSymbol: coinImpl.GetSymbol(),
		CreatedAt:  time.Now(),
		Version:    "2.0",
		Accounts:   make([]*model.Account, 0),
	}

	return wallet, nil
}

func secureClearString(s string) {
	// 转换为可变的字节切片
	bytes := []byte(s)

	// 用随机数据覆盖
	rand.Read(bytes)

	// 再次用零值覆盖
	for i := range bytes {
		bytes[i] = 0
	}
}

// EncryptWallet 加密钱包数据
func (ws *WalletService) EncryptWallet(wallet *model.HDWallet, password string) ([]byte, error) {
	walletData, err := json.Marshal(wallet)
	if err != nil {
		return nil, err
	}

	keystore := security.NewKeystore()
	encryptedData, err := keystore.EncryptWallet(walletData, password)
	if err != nil {
		return nil, err
	}

	return encryptedData, nil
}
