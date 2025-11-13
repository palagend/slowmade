package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/palagend/slowmade/internal/config"
	"github.com/palagend/slowmade/internal/mvc/models"
)

// WalletRepository 处理钱包数据的持久化操作
type WalletRepository struct {
	mu             sync.RWMutex // 改为值类型，会自动初始化
	VirtualWallets map[string]*models.VirtualWallet
	config         *config.AppConfig
}

// NewWalletRepository 创建新的钱包存储库实例
func NewWalletRepository(appConfig *config.AppConfig) *WalletRepository {
	return &WalletRepository{
		VirtualWallets: make(map[string]*models.VirtualWallet),
		config:         appConfig,
		// mu 字段会自动初始化为零值（可用的互斥锁）
	}
}

// Save 保存钱包到存储
func (r *WalletRepository) Save(wallet *models.VirtualWallet) error {
	r.mu.Lock() // 现在可以安全调用，因为 mu 已初始化
	defer r.mu.Unlock()

	// 获取配置的keystore目录
	keystoreDir := r.config.Storage.KeystoreDir
	if keystoreDir == "" {
		keystoreDir = "~/.slowmade/keystores" // 默认值
	}

	// 处理家目录的~符号
	if strings.HasPrefix(keystoreDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %v", err)
		}
		keystoreDir = filepath.Join(homeDir, keystoreDir[2:])
	}

	// 创建目录（如果不存在）
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		return fmt.Errorf("failed to create keystore directory: %v", err)
	}

	// 构建文件路径：使用wallet ID作为文件名
	filename := wallet.ID + ".json"
	filePath := filepath.Join(keystoreDir, filename)

	// 序列化钱包数据为JSON
	walletData, err := json.MarshalIndent(wallet, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallet data: %v", err)
	}

	// 写入文件（仅限当前用户读写）
	if err := os.WriteFile(filePath, walletData, 0600); err != nil {
		return fmt.Errorf("failed to write wallet file: %v", err)
	}

	// 同时更新内存中的缓存
	r.VirtualWallets[wallet.ID] = wallet

	return nil
}

// Load 从文件系统加载钱包
func (r *WalletRepository) Load(walletID string) (*models.VirtualWallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 首先检查内存缓存
	if wallet, exists := r.VirtualWallets[walletID]; exists {
		return wallet, nil
	}

	// 构建文件路径
	keystoreDir := r.config.Storage.KeystoreDir
	if strings.HasPrefix(keystoreDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		keystoreDir = filepath.Join(homeDir, keystoreDir[2:])
	}

	filePath := filepath.Join(keystoreDir, walletID+".json")

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %v", err)
	}

	// 反序列化
	var wallet models.VirtualWallet
	if err := json.Unmarshal(data, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet data: %v", err)
	}

	// 更新内存缓存
	r.VirtualWallets[walletID] = &wallet

	return &wallet, nil
}

// LoadAll 加载所有钱包
func (r *WalletRepository) LoadAll() ([]*models.VirtualWallet, error) {
	// 加载所有钱包也需要加锁，因为会操作 VirtualWallets
	r.mu.Lock()
	defer r.mu.Unlock()

	keystoreDir := r.config.Storage.KeystoreDir
	if strings.HasPrefix(keystoreDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		keystoreDir = filepath.Join(homeDir, keystoreDir[2:])
	}

	entries, err := os.ReadDir(keystoreDir)
	if err != nil {
		// 如果目录不存在，返回空列表而不是错误
		if os.IsNotExist(err) {
			return []*models.VirtualWallet{}, nil
		}
		return nil, err
	}

	var wallets []*models.VirtualWallet
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			walletID := strings.TrimSuffix(entry.Name(), ".json")

			// 使用独立的加载逻辑，避免递归调用 Load 方法
			filePath := filepath.Join(keystoreDir, entry.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Warning: Failed to read wallet file %s: %v\n", entry.Name(), err)
				continue
			}

			var wallet models.VirtualWallet
			if err := json.Unmarshal(data, &wallet); err != nil {
				fmt.Printf("Warning: Failed to unmarshal wallet %s: %v\n", walletID, err)
				continue
			}

			wallets = append(wallets, &wallet)
			// 更新内存缓存
			r.VirtualWallets[walletID] = &wallet
		}
	}

	return wallets, nil
}
