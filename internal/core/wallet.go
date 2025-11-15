package core

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/palagend/slowmade/pkg/crypto"
	"github.com/palagend/slowmade/pkg/i18n"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/scrypt"
)

type WalletStatus int

const (
	StatusLocked WalletStatus = iota
	StatusUnlocked
)

type Wallet struct {
	account    *accounts.Account
	keystore   *keystore.KeyStore
	privateKey []byte
	status     WalletStatus
	createdAt  time.Time
}

type WalletManager struct {
	wallets map[common.Address]*Wallet
	current *Wallet
}

func NewWalletManager() *WalletManager {
	return &WalletManager{
		wallets: make(map[common.Address]*Wallet),
	}
}

func (wm *WalletManager) CreateWallet(password, mnemonic string) (string, error) {
	// 生成助记词（如果未提供）
	if mnemonic == "" {
		var err error
		mnemonic, err = crypto.GenerateMnemonic()
		if err != nil {
			return "", fmt.Errorf(i18n.Tr("ERR_GENERATE_MNEMONIC"), err)
		}
	}

	// 验证助记词
	if !crypto.ValidateMnemonic(mnemonic) {
		return "", fmt.Errorf(i18n.Tr("ERR_INVALID_MNEMONIC"))
	}

	// 从助记词派生种子
	seed := crypto.NewSeed(mnemonic, "")

	// 创建HD钱包
	wallet, err := crypto.NewHDWallet(seed)
	if err != nil {
		return "", fmt.Errorf(i18n.Tr("ERR_CREATE_WALLET"), err)
	}

	// 生成加密密钥
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key, err := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, 32)
	if err != nil {
		return "", err
	}

	// 安全存储私钥
	encryptedKey, err := crypto.EncryptKey(wallet.PrivateKey(), key, salt)
	if err != nil {
		return "", err
	}

	// 创建钱包实例
	account := &accounts.Account{
		Address: wallet.Address(),
	}

	newWallet := &Wallet{
		account:    account,
		privateKey: encryptedKey,
		status:     StatusUnlocked,
		createdAt:  time.Now(),
	}

	wm.wallets[wallet.Address()] = newWallet
	wm.current = newWallet

	return mnemonic, nil
}

func (wm *WalletManager) UnlockWallet(address common.Address, password string) error {
	wallet, exists := wm.wallets[address]
	if !exists {
		return fmt.Errorf(i18n.Tr("ERR_WALLET_NOT_FOUND"))
	}

	// 解密私钥（实际实现需要从keystore读取）
	// 这里简化处理
	wallet.status = StatusUnlocked
	wm.current = wallet

	return nil
}

func (wm *WalletManager) LockWallet() {
	if wm.current != nil {
		// 安全清理内存中的私钥
		wm.current.secureClear()
		wm.current.status = StatusLocked
		wm.current = nil
	}
}

func (w *Wallet) secureClear() {
	// 安全清理私钥内存
	for i := range w.privateKey {
		w.privateKey[i] = 0
	}
	w.privateKey = nil
}

func (wm *WalletManager) GetCurrentWallet() *Wallet {
	return wm.current
}

func (wm *WalletManager) WalletExists(address common.Address) bool {
	_, exists := wm.wallets[address]
	return exists
}
