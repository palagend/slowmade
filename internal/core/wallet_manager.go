package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/palagend/slowmade/internal/security"
	"github.com/palagend/slowmade/pkg/crypto"
	"github.com/palagend/slowmade/pkg/logging"
	"github.com/palagend/slowmade/pkg/mnemonic"
	"github.com/tyler-smith/go-bip39"
)

// DefaultWalletManager 默认的钱包管理器实现
type DefaultWalletManager struct {
	storage         StorageHandler
	mnemonicService mnemonic.MnemonicService

	rootWallet *HDRootWallet
	isLocked   bool
	isLoaded   bool
	mutex      sync.RWMutex
	once       sync.Once
	cloak      string // A cloak is not a password! Any variation entered in future loads a valid wallet, but with different addresses.
}

// NewDefaultWalletManager 创建新的钱包管理器实例
func NewDefaultWalletManager(storage StorageHandler, cloak string) *DefaultWalletManager {
	return &DefaultWalletManager{
		storage:         storage,
		mnemonicService: mnemonic.NewBIP39MnemonicService(),
		isLocked:        true,
		cloak:           cloak,
	}
}
func (wm *DefaultWalletManager) Seed() ([]byte, error) {
	password, err := security.Password()
	if err != nil {
		return nil, err
	}
	seed, err := crypto.DecryptData(wm.rootWallet.EncryptedMnemonic, string(password))
	if err != nil {
		return nil, err
	}
	return seed, nil
}

// CreateNewWallet 创建新钱包（生成助记词和种子）
func (wm *DefaultWalletManager) CreateNewWallet(password string) (*HDRootWallet, error) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	// 检查是否已存在钱包
	if wm.rootWallet != nil {
		return nil, errors.New("钱包已存在")
	}
	logging.Debug("Generating mnemonic...")
	// 使用助记词服务生成助记词
	mnemonic, err := wm.mnemonicService.GenerateMnemonic(256) // 256位强度
	if err != nil {
		return nil, fmt.Errorf("生成助记词失败: %w", err)
	}
	logging.Debug("Generating seed...")
	// 从助记词生成种子
	seed := wm.mnemonicService.GenerateSeedFromMnemonic(mnemonic, wm.cloak)

	logging.Debug("Encrypting mnemonic...")
	// 使用加密服务加密敏感数据
	encryptedMnemonic, err := crypto.EncryptData([]byte(mnemonic), password)
	if err != nil {
		return nil, fmt.Errorf("加密助记词失败: %w", err)
	}

	logging.Debug("Encrypting seed...")
	encryptedSeed, err := crypto.EncryptData(seed, password)
	if err != nil {
		return nil, fmt.Errorf("加密种子失败: %w", err)
	}

	// 创建钱包实例
	wallet := &HDRootWallet{
		EncryptedMnemonic: encryptedMnemonic,
		EncryptedSeed:     encryptedSeed,
		CreationTime:      uint64(time.Now().Unix()),
	}

	// 保存到存储
	if err := wm.storage.SaveRootWallet(wallet); err != nil {
		return nil, fmt.Errorf("保存钱包失败: %w", err)
	}

	wm.rootWallet = wallet
	return wallet, nil
}

// ExportMnemonic 导出助记词
func (wm *DefaultWalletManager) ExportMnemonic(password string) (string, error) {
	mne, err := crypto.DecryptData(wm.rootWallet.EncryptedMnemonic, password)
	if err != nil {
		return "", fmt.Errorf("解密失败！")
	}
	if mne != nil {
		return string(mne), nil
	}
	return "", fmt.Errorf("导出助记词失败！")
}

// RestoreWalletFromMnemonic 从助记词恢复钱包
func (wm *DefaultWalletManager) RestoreWalletFromMnemonic(mnemonic, password string) (*HDRootWallet, error) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	// 使用助记词服务验证助记词有效性
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("无效的助记词")
	}

	// 从助记词生成种子
	seed := wm.mnemonicService.GenerateSeedFromMnemonic(mnemonic, password)

	// 使用加密服务加密敏感数据
	encryptedMnemonic, err := crypto.EncryptData([]byte(mnemonic), password)
	if err != nil {
		return nil, fmt.Errorf("加密助记词失败: %w", err)
	}

	encryptedSeed, err := crypto.EncryptData(seed, password)
	if err != nil {
		return nil, fmt.Errorf("加密种子失败: %w", err)
	}

	// 创建钱包实例
	wallet := &HDRootWallet{
		EncryptedMnemonic: encryptedMnemonic,
		EncryptedSeed:     encryptedSeed,
		CreationTime:      uint64(time.Now().Unix()),
	}

	// 保存到存储
	if err := wm.storage.SaveRootWallet(wallet); err != nil {
		return nil, fmt.Errorf("保存钱包失败: %w", err)
	}

	wm.rootWallet = wallet
	return wallet, nil
}

// UnlockWallet 解锁钱包
func (wm *DefaultWalletManager) UnlockWallet(password string) error {
	wm.once.Do(func() {
		if wm.rootWallet == nil {
			wm.rootWallet, _ = wm.storage.LoadRootWallet()
		}
	})
	if wm.rootWallet == nil {
		return errors.New("钱包不存在")
	}
	_, err := crypto.DecryptData(wm.rootWallet.EncryptedSeed, password)
	if err != nil {
		return errors.New("密码错误")
	}

	wm.isLocked = false
	return nil
}

// LockWallet 锁定钱包，并安全地清除内存中的敏感信息。
func (wm *DefaultWalletManager) LockWallet() {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	// 防御性检查：确保钱包实例存在且当前已解锁
	if wm.rootWallet == nil || wm.isLocked {
		// 即使未解锁或钱包为空，也确保状态正确
		wm.isLocked = true
		return
	}

	// 最终状态设置
	wm.isLocked = true
	wm.rootWallet = nil // 考虑清空根引用，促进GC回收非敏感数据
}

// IsUnlocked 检查钱包当前是否已解锁
func (wm *DefaultWalletManager) IsLocked() bool {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	return wm.isLocked
}
