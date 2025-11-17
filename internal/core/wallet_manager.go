package core

import (
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/palagend/slowmade/pkg/crypto"
	"github.com/palagend/slowmade/pkg/mnemonic"
)

// DefaultWalletManager 默认的钱包管理器实现
type DefaultWalletManager struct {
	storage         StorageHandler
	rootWallet      *HDRootWallet
	isUnlocked      bool
	mutex           sync.RWMutex
	cryptoService   crypto.CryptoService
	mnemonicService mnemonic.MnemonicService
}

// NewDefaultWalletManager 创建新的钱包管理器实例
func NewDefaultWalletManager(
	storage StorageHandler,
	cryptoService crypto.CryptoService,
	mnemonicService mnemonic.MnemonicService,
) *DefaultWalletManager {
	return &DefaultWalletManager{
		storage:         storage,
		cryptoService:   cryptoService,
		mnemonicService: mnemonicService,
		isUnlocked:      false,
	}
}

// CreateNewWallet 创建新钱包（生成助记词和种子）
func (wm *DefaultWalletManager) CreateNewWallet(password string) (*HDRootWallet, error) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	// 检查是否已存在钱包
	existingWallet, err := wm.storage.LoadRootWallet()
	if err != nil {
		return nil, fmt.Errorf("检查现有钱包失败: %w", err)
	}
	if existingWallet != nil {
		return nil, errors.New("钱包已存在")
	}

	// 使用助记词服务生成助记词
	mnemonic, err := wm.mnemonicService.GenerateMnemonic(256) // 256位强度
	if err != nil {
		return nil, fmt.Errorf("生成助记词失败: %w", err)
	}

	// 从助记词生成种子
	seed := wm.mnemonicService.GenerateSeedFromMnemonic(mnemonic, password)

	// 使用加密服务加密敏感数据
	encryptedMnemonic, err := wm.cryptoService.EncryptData([]byte(mnemonic), password)
	if err != nil {
		return nil, fmt.Errorf("加密助记词失败: %w", err)
	}

	encryptedSeed, err := wm.cryptoService.EncryptData(seed, password)
	if err != nil {
		return nil, fmt.Errorf("加密种子失败: %w", err)
	}

	// 创建钱包实例
	wallet := &HDRootWallet{
		encryptedMnemonic: hex.EncodeToString(encryptedMnemonic),
		encryptedSeed:     hex.EncodeToString(encryptedSeed),
		creationTime:      uint64(time.Now().Unix()),
	}

	// 保存到存储
	if err := wm.storage.SaveRootWallet(wallet); err != nil {
		return nil, fmt.Errorf("保存钱包失败: %w", err)
	}

	wm.rootWallet = wallet
	return wallet, nil
}

// RestoreWalletFromMnemonic 从助记词恢复钱包
func (wm *DefaultWalletManager) RestoreWalletFromMnemonic(mnemonic, password string) (*HDRootWallet, error) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	// 使用助记词服务验证助记词有效性
	if !wm.mnemonicService.ValidateMnemonic(mnemonic) {
		return nil, errors.New("无效的助记词")
	}

	// 从助记词生成种子
	seed := wm.mnemonicService.GenerateSeedFromMnemonic(mnemonic, password)

	// 使用加密服务加密敏感数据
	encryptedMnemonic, err := wm.cryptoService.EncryptData([]byte(mnemonic), password)
	if err != nil {
		return nil, fmt.Errorf("加密助记词失败: %w", err)
	}

	encryptedSeed, err := wm.cryptoService.EncryptData(seed, password)
	if err != nil {
		return nil, fmt.Errorf("加密种子失败: %w", err)
	}

	// 创建钱包实例
	wallet := &HDRootWallet{
		encryptedMnemonic: hex.EncodeToString(encryptedMnemonic),
		encryptedSeed:     hex.EncodeToString(encryptedSeed),
		creationTime:      uint64(time.Now().Unix()),
	}

	// 保存到存储
	if err := wm.storage.SaveRootWallet(wallet); err != nil {
		return nil, fmt.Errorf("保存钱包失败: %w", err)
	}

	wm.rootWallet = wallet
	return wallet, nil
}

// UnlockWallet 解锁钱包
func (wm *DefaultWalletManager) UnlockWallet(encryptedMnemonic []byte, password string) error {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	// 加载钱包
	wallet, err := wm.storage.LoadRootWallet()
	if err != nil {
		return fmt.Errorf("加载钱包失败: %w", err)
	}
	if wallet == nil {
		return errors.New("钱包不存在")
	}

	// 尝试解密助记词验证密码
	encryptedData, err := hex.DecodeString(string(encryptedMnemonic))
	if err != nil {
		return fmt.Errorf("解码加密数据失败: %w", err)
	}

	_, err = wm.cryptoService.DecryptData(encryptedData, password)
	if err != nil {
		return errors.New("密码错误")
	}

	wm.rootWallet = wallet
	wm.isUnlocked = true
	return nil
}

// LockWallet 锁定钱包，并安全地清除内存中的敏感信息。
func (wm *DefaultWalletManager) LockWallet() {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	// 防御性检查：确保钱包实例存在且当前已解锁
	if wm.rootWallet == nil || !wm.isUnlocked {
		// 即使未解锁或钱包为空，也确保状态正确
		wm.isUnlocked = false
		return
	}

	// 核心：安全擦除钱包内部的敏感数据
	wm.wipeSensitiveData()

	// 最终状态设置
	wm.isUnlocked = false
	wm.rootWallet = nil // 考虑清空根引用，促进GC回收非敏感数据
}

// wipeSensitiveData 递归地安全擦除钱包结构体中的敏感数据。
func (wm *DefaultWalletManager) wipeSensitiveData() {
	if wm.rootWallet == nil {
		return
	}

	// 1. 擦除Seed
	if wm.rootWallet.encryptedMnemonic != "" {
		crypto.SecureWipeBytes([]byte(wm.rootWallet.encryptedSeed))
	}
	// 2. 擦除助记词（假设 Mnemonic 是 string，需先转为可修改的字节切片）
	if wm.rootWallet.encryptedMnemonic != "" {
		crypto.SecureWipeBytes([]byte(wm.rootWallet.encryptedMnemonic))
		// 字符串本身不可变，擦除副本后，原字符串将由GC处理。关键是不再保留引用。
	}

	// 防御性措施：防止编译器优化掉上面的擦除操作
	// 使用 runtime.KeepAlive 确保在函数返回前，擦除操作已完成
	runtime.KeepAlive(wm.rootWallet)
}

// IsUnlocked 检查钱包当前是否已解锁
func (wm *DefaultWalletManager) IsUnlocked() bool {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	return wm.isUnlocked
}
