package core

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/palagend/slowmade/internal/security"
	"github.com/palagend/slowmade/pkg/coin"
	"github.com/palagend/slowmade/pkg/crypto"
	"github.com/palagend/slowmade/pkg/logging"
	"github.com/tyler-smith/go-bip32"
)

var (
	ErrWalletLocked        = errors.New("wallet is locked")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrWalletAlreadyExists = errors.New("wallet already exists")
	ErrWalletNotCreated    = errors.New("wallet not created")
)

// DefaultAccountManager 默认的账户管理器实现
type DefaultAccountManager struct {
	walletManager WalletManager
	storage       StorageHandler
	maxLength     int // ID最大长度
}

// NewDefaultAccountManager 创建新的账户管理器
func NewDefaultAccountManager(walletManager WalletManager, storage StorageHandler) AccountManager {
	return &DefaultAccountManager{
		walletManager: walletManager,
		storage:       storage,
	}
}

// CreateNewAccount 创建新账户
func (am *DefaultAccountManager) CreateNewAccount(derivationPath *DerivationPath) (*CoinAccount, error) {
	if am.walletManager.IsLocked() {
		return nil, ErrWalletLocked
	}

	coinSymbol := coin.CoinSymbol(derivationPath.CoinType)
	if coinSymbol == "" {
		return nil, fmt.Errorf("该币种（coin_type=%s）暂不支持", derivationPath.CoinTypeString())
	}
	// 派生账户密钥
	dp := derivationPath.MaskSuffix()
	accountKey, err := am.deriveAccountKey(dp)
	if err != nil {
		return nil, fmt.Errorf("failed to derive account key: %w", err)
	}

	password, err := security.Password()
	if err != nil {
		return nil, err
	}
	serializedKey, err := accountKey.Serialize()
	if err != nil {
		return nil, err
	}
	logging.Debugf("serializedKey len is %d", len(serializedKey))
	encryptedPrivateKey, err := crypto.EncryptData(serializedKey, string(password))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt account private key: %w", err)
	}

	account := &CoinAccount{
		ID:                         am.IDString(dp.String()),
		CoinSymbol:                 coinSymbol,
		DerivationPath:             dp.String(),
		EncryptedAccountPrivateKey: encryptedPrivateKey,
	}

	// 保存账户
	if err := am.storage.SaveAccount(account); err != nil {
		return nil, fmt.Errorf("failed to save account: %w", err)
	}

	return account, nil
}

// GetAccountsByCoin 获取指定币种的所有账户
func (am *DefaultAccountManager) GetAccountsByCoin(coinType uint32) ([]*CoinAccount, error) {
	if am.walletManager.IsLocked() {
		return nil, ErrWalletLocked
	}
	accounts, err := am.storage.LoadAccounts()
	if err != nil {
		return nil, err
	}

	var result []*CoinAccount
	for _, account := range accounts {
		if account.CoinType() == coinType {
			result = append(result, account)
		}
	}

	return result, nil
}

// DeriveAddress 派生新地址
func (am *DefaultAccountManager) DeriveAddress(accountID string, changeType uint32, addressIndex uint32) (*AddressKey, error) {
	if am.walletManager.IsLocked() {
		return nil, ErrWalletLocked
	}

	// 获取账户
	accounts, err := am.storage.LoadAccounts()
	if err != nil {
		return nil, err
	}

	var targetAccount *CoinAccount
	for _, account := range accounts {
		if account.ID == accountID {
			targetAccount = account
			break
		}
	}

	if targetAccount == nil {
		return nil, errors.New("account not found")
	}

	// 派生地址密钥
	addressKey, err := am.deriveAddressKey(targetAccount, changeType, addressIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address key: %w", err)
	}

	// 生成地址（这里需要根据币种实现具体的地址生成逻辑）
	address, publicKey, err := am.generateAddress(targetAccount.CoinType(), addressKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate address: %w", err)
	}

	// 加密私钥（在实际应用中需要使用密码）
	password, err := security.Password()
	if err != nil {
		return nil, err
	}
	encryptedPrivateKey, err := crypto.EncryptData(addressKey.Key, string(password))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	addressKeyObj := &AddressKey{
		AccountID:           accountID,
		ChangeType:          changeType,
		AddressIndex:        addressIndex,
		EncryptedPrivateKey: encryptedPrivateKey,
		PublicKey:           hex.EncodeToString(publicKey),
		Address:             address,
		CoinSymbol:          coin.CoinSymbol(targetAccount.CoinType()),
	}

	// 保存地址
	if err := am.storage.SaveAddress(addressKeyObj); err != nil {
		return nil, fmt.Errorf("failed to save address: %w", err)
	}

	return addressKeyObj, nil
}

// GetAddresses 获取指定账户的所有地址
func (am *DefaultAccountManager) GetAddresses(accountID string) ([]*AddressKey, error) {
	return am.storage.LoadAddresses(accountID)
}

// 派生账户密钥
func (am *DefaultAccountManager) deriveAccountKey(derivationPath *DerivationPath) (*bip32.Key, error) {
	if derivationPath == nil {
		return nil, fmt.Errorf("derivationPath cannot be nil")
	}
	// BIP44 路径：m/44'/coinType'/accountIndex'
	seed, err := am.walletManager.Seed()
	if err != nil {
		return nil, err
	}
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	// purpose: 44' (硬化派生)
	purposeKey, err := masterKey.NewChildKey(derivationPath.Purpose)
	if err != nil {
		return nil, err
	}

	// coinType 0'
	coinTypeKey, err := purposeKey.NewChildKey(derivationPath.CoinType)
	if err != nil {
		return nil, err
	}

	// accountIndex 0'
	accountKey, err := coinTypeKey.NewChildKey(derivationPath.AccountIndex)
	if err != nil {
		return nil, err
	}
	return accountKey, nil
}

// 派生地址密钥
func (am *DefaultAccountManager) deriveAddressKey(account *CoinAccount, changeType, addressIndex uint32) (*bip32.Key, error) {
	password, err := security.Password()
	if err != nil {
		return nil, err
	}
	accountPrivateKey, err := crypto.DecryptData(account.EncryptedAccountPrivateKey, string(password))
	if err != nil {
		return nil, err
	}

	// 重新创建账户密钥
	accountKey, err := bip32.Deserialize(accountPrivateKey)
	if err != nil {
		return nil, err
	}

	// 派生 change 路径：changeType (0=外部, 1=找零)
	changeKey, err := accountKey.NewChildKey(changeType)
	if err != nil {
		return nil, err
	}

	// 派生地址索引
	addressKey, err := changeKey.NewChildKey(addressIndex)
	if err != nil {
		return nil, err
	}

	return addressKey, nil
}

func (am *DefaultAccountManager) generateAddress(coinType uint32, key *bip32.Key) (string, []byte, error) {
	if key == nil {
		return "", nil, errors.New("key cannot be nil")
	}

	publicKey := key.PublicKey().Key

	var generator AddressGenerator
	var address string
	var err error

	switch coinType {
	case coin.CoinTypeBTC | coin.HardenedBit:
		generator = &BTCAddressGenerator{}
		address, err = generator.GenerateAddress(publicKey)

	case coin.CoinTypeETH | coin.HardenedBit:
		generator = &ETHAddressGenerator{}
		address, err = generator.GenerateAddress(publicKey)

	case coin.CoinTypeSOL | coin.HardenedBit:
		generator = &SOLAddressGenerator{}
		address, err = generator.GenerateAddress(publicKey)

	case coin.CoinTypeBNB | coin.HardenedBit:
		generator = &BNBAddressGenerator{}
		address, err = generator.GenerateAddress(publicKey)

	case coin.CoinTypeSUI | coin.HardenedBit:
		generator = &SUIAddressGenerator{}
		address, err = generator.GenerateAddress(publicKey)

	default:
		return "", nil, fmt.Errorf("unsupported coin type: %d", coinType)
	}

	if err != nil {
		return "", nil, fmt.Errorf("failed to generate address for coin type %d: %w", coinType, err)
	}

	return address, publicKey, nil
}

func (am *DefaultAccountManager) IDString(derivationPath string) string {
	// 添加前缀和哈希
	hash := sha256.Sum256([]byte(derivationPath))
	prefix := "file_"
	hexString := hex.EncodeToString(hash[:])

	filename := prefix + hexString

	if am.maxLength > 0 && len(filename) > am.maxLength {
		// 确保前缀完整，截断哈希部分
		if am.maxLength > len(prefix) {
			return prefix + hexString[:am.maxLength-len(prefix)]
		}
		return prefix[:am.maxLength]
	}
	return filename
}
