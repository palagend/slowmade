package core

import "fmt"

// 定义了钱包生命周期管理的核心操作
type WalletManager interface {
	CreateNewWallet(password string) (*HDRootWallet, error)                     // 创建新钱包（生成助记词和种子）
	ExportMnemonic(password string) (string, error)                             // 导出助记词
	RestoreWalletFromMnemonic(mnemonic, password string) (*HDRootWallet, error) // 从助记词恢复钱包
	UnlockWallet(password string) error                                         // 解锁钱包（解密根种子）
	LockWallet()                                                                // 锁定钱包（清除内存中的敏感信息）
	IsUnlocked() bool                                                           // 检查钱包当前是否已解锁
}

// AccountManager 定义了账户管理的操作
type AccountManager interface {
	CreateNewAccount(coinType uint32, accountIndex uint32, password string) (*CoinAccount, error) // 创建新币种账户
	GetAccountsByCoin(coinType uint32) ([]*CoinAccount, error)                                    // 获取指定币种的所有账户
	DeriveAddress(accountID string, changeType uint32, addressIndex uint32) (*AddressKey, error)  // 为指定账户派生新地址
	GetAddresses(accountID string) ([]*AddressKey, error)                                         // 获取指定账户下的所有地址
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

// 派生地址
/*type DerivationPath struct {
	purpose   uint32
	coinType  uint32
	account   uint32
	change    uint32
	addrIndex uint32
}

func (d *DerivationPath) String() string {
	return fmt.Sprintf("m/%d'/%d'/%d'/%d/%d", d.purpose, d.coinType, d.account, d.change, d.addrIndex)
}*/

// 常量定义 - BIP44标准币种类型
const (
	CoinTypeBTC uint32 = 0
	CoinTypeETH uint32 = 60
	CoinTypeSOL uint32 = 501
	CoinTypeBNB uint32 = 714
	CoinTypeSUI uint32 = 784
)

// 根钱包
type HDRootWallet struct {
	EncryptedMnemonic string
	EncryptedSeed     string
	CreationTime      uint64
}

type CoinAccount struct {
	id                         string
	coinStymbol                string
	derivationPath             string
	purpose                    uint32
	coinType                   uint32
	accountIndex               uint32
	encryptedAccountPrivateKey string // 加密的账户层级私钥
}

type AddressKey struct {
	accountID           string // 关联的账户ID
	changeType          uint32 // 0-外部账户，1-内部找零
	addrIndex           uint32 // 地址索引
	encryptedPrivateKey string // 加密后的地址私钥
	publicKey           string // 对应的公钥
	address             string // 生成的区块链地址
}

func (c *CoinAccount) CoinSymbol() string {
	return c.coinStymbol
}

func (c *CoinAccount) ID() string {
	return c.id
}

// 辅助函数：生成BIP44派生路径
func (account *CoinAccount) GetDerivationPath() string {
	if account.derivationPath == "" {
		return fmt.Sprintf("m/%d'/%d'/%d'", account.purpose, account.coinType, account.accountIndex)
	}
	return account.derivationPath
}

// 辅助函数：生成完整地址派生路径
func (address *AddressKey) GetFullDerivationPath(account *CoinAccount) string {
	return fmt.Sprintf("%s/%d/%d", account.GetDerivationPath(), address.changeType, address.addrIndex)
}
