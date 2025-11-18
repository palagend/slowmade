package core

// 定义了钱包生命周期管理的核心操作
type WalletManager interface {
	CreateNewWallet(password string) (*HDRootWallet, error)                     // 创建新钱包（生成助记词和种子）
	ExportMnemonic(password string) (string, error)                             // 导出助记词
	RestoreWalletFromMnemonic(mnemonic, password string) (*HDRootWallet, error) // 从助记词恢复钱包
	UnlockWallet(password string) error                                         // 解锁钱包（解密根种子）
	LockWallet()                                                                // 锁定钱包（清除内存中的敏感信息）
	IsLocked() bool                                                             // 检查钱包当前是否已解锁
	Seed() ([]byte, error)                                                      // 返回解密后的Seed
}

// AccountManager 定义了账户管理的操作
type AccountManager interface {
	CreateNewAccount(derivationPath *DerivationPath) (*CoinAccount, error)                       // 创建新币种账户
	GetAccountsByCoin(coinType uint32) ([]*CoinAccount, error)                                   // 获取指定币种的所有账户
	DeriveAddress(accountID string, changeType uint32, addressIndex uint32) (*AddressKey, error) // 为指定账户派生新地址
	GetAddresses(accountID string) ([]*AddressKey, error)                                        // 获取指定账户下的所有地址
	IDString(derivationPath string) string
}

// StorageHandler 定义了数据持久化的操作，支持不同的后端（如文件系统、数据库）
type StorageHandler interface {
	SaveRootWallet(wallet *HDRootWallet) error
	LoadRootWallet() (*HDRootWallet, error)
	SaveAccount(account *CoinAccount) error
	LoadAccounts() ([]*CoinAccount, error)
	SaveAddress(address *AddressKey) error
	LoadAddresses(accountID string) ([]*AddressKey, error)
}
