package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileStorage 基于本地文件系统的存储实现
type FileStorage struct {
	baseDir      string
	walletsDir   string
	accountsDir  string
	addressesDir string
	mutex        sync.RWMutex
}

// NewFileStorage 创建新的文件存储实例
func NewFileStorage(baseDir string) (*FileStorage, error) {
	storage := &FileStorage{
		baseDir:      baseDir,
		walletsDir:   filepath.Join(baseDir, "wallets"),
		accountsDir:  filepath.Join(baseDir, "accounts"),
		addressesDir: filepath.Join(baseDir, "addresses"),
	}

	// 创建必要的目录结构
	dirs := []string{storage.walletsDir, storage.accountsDir, storage.addressesDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}
	}

	return storage, nil
}

// SaveRootWallet 保存根钱包数据到JSON文件
func (fs *FileStorage) SaveRootWallet(wallet *HDRootWallet) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	walletFile := filepath.Join(fs.walletsDir, "root_wallet.json")
	return fs.saveToFile(walletFile, wallet)
}

// LoadRootWallet 从JSON文件加载根钱包数据
func (fs *FileStorage) LoadRootWallet() (*HDRootWallet, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	walletFile := filepath.Join(fs.walletsDir, "root_wallet.json")
	var wallet HDRootWallet
	if err := fs.loadFromFile(walletFile, &wallet); err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 文件不存在返回nil而不是错误
		}
		return nil, err
	}
	return &wallet, nil
}

// SaveAccount 保存账户数据到JSON文件
func (fs *FileStorage) SaveAccount(account *CoinAccount) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// 先更新账户列表
	accounts, err := fs.loadAllAccounts()
	if err != nil {
		return err
	}

	// 检查账户是否已存在，更新或添加
	found := false
	for i, acc := range accounts {
		if acc.id == account.id {
			accounts[i] = account
			found = true
			break
		}
	}
	if !found {
		accounts = append(accounts, account)
	}

	// 保存更新后的账户列表
	accountsFile := filepath.Join(fs.accountsDir, "accounts.json")
	return fs.saveToFile(accountsFile, accounts)
}

// LoadAccounts 加载所有账户数据
func (fs *FileStorage) LoadAccounts() ([]*CoinAccount, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	return fs.loadAllAccounts()
}

// loadAllAccounts 内部方法：加载所有账户
func (fs *FileStorage) loadAllAccounts() ([]*CoinAccount, error) {
	accountsFile := filepath.Join(fs.accountsDir, "accounts.json")
	var accounts []*CoinAccount
	if err := fs.loadFromFile(accountsFile, &accounts); err != nil {
		if os.IsNotExist(err) {
			return []*CoinAccount{}, nil // 文件不存在返回空列表
		}
		return nil, err
	}
	return accounts, nil
}

// SaveAddress 保存地址数据到对应账户的文件
func (fs *FileStorage) SaveAddress(address *AddressKey) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	addressFile := filepath.Join(fs.addressesDir, fmt.Sprintf("%s_addresses.json", address.accountID))

	var addresses []*AddressKey
	if err := fs.loadFromFile(addressFile, &addresses); err != nil && !os.IsNotExist(err) {
		return err
	}

	// 检查地址是否已存在，更新或添加
	found := false
	for i, addr := range addresses {
		if addr.accountID == address.accountID &&
			addr.changeType == address.changeType &&
			addr.addrIndex == address.addrIndex {
			addresses[i] = address
			found = true
			break
		}
	}
	if !found {
		addresses = append(addresses, address)
	}

	return fs.saveToFile(addressFile, addresses)
}

// LoadAddresses 加载指定账户的所有地址
func (fs *FileStorage) LoadAddresses(accountID string) ([]*AddressKey, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	addressFile := filepath.Join(fs.addressesDir, fmt.Sprintf("%s_addresses.json", accountID))
	var addresses []*AddressKey
	if err := fs.loadFromFile(addressFile, &addresses); err != nil {
		if os.IsNotExist(err) {
			return []*AddressKey{}, nil // 文件不存在返回空列表
		}
		return nil, err
	}
	return addresses, nil
}

// saveToFile 通用方法：保存数据到JSON文件
func (fs *FileStorage) saveToFile(filename string, data interface{}) error {
	// 创建临时文件以确保写入原子性
	tempFile := filename + ".tmp"

	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // 美化JSON输出
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("编码JSON失败: %w", err)
	}

	// 确保数据写入磁盘
	if err := file.Sync(); err != nil {
		return fmt.Errorf("同步文件失败: %w", err)
	}

	// 重命名临时文件为正式文件（原子操作）
	if err := os.Rename(tempFile, filename); err != nil {
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}

// loadFromFile 通用方法：从JSON文件加载数据
func (fs *FileStorage) loadFromFile(filename string, v interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("解码JSON失败: %w", err)
	}

	return nil
}

// CheckStorageHealth 检查存储系统健康状态
func (fs *FileStorage) CheckStorageHealth() error {
	// 检查目录权限
	dirs := []string{fs.walletsDir, fs.accountsDir, fs.addressesDir}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("目录不可访问 %s: %w", dir, err)
		}

		// 测试写入权限
		testFile := filepath.Join(dir, ".healthcheck")
		if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
			return fmt.Errorf("目录不可写 %s: %w", dir, err)
		}
		os.Remove(testFile) // 清理测试文件
	}
	return nil
}
