package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

// VirtualWallet 充血模型领域实体
type VirtualWallet struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Mnemonic           string             `json:"mnemonic,omitempty"`
	RootKey            []byte             `json:"root_key_encrypted"`
	CreationTime       time.Time          `json:"creation_time"`
	Coins              []Coin             `json:"coins"`
	IsLocked           bool               `json:"is_locked"`
	IsEncrypted        bool               `json:"is_encrypted"`
	EncryptionMetadata EncryptionMetadata `json:"encryption_metadata"`
	EncryptedData      string             `json:"encrypted_data"`
}

// EncryptionMetadata 加密元数据
type EncryptionMetadata struct {
	Algorithm   string `json:"algorithm"`
	Version     string `json:"version"`
	IV          string `json:"iv"`           // 初始化向量
	Salt        string `json:"salt"`         // 盐值
	EncryptedAt int64  `json:"encrypted_at"` // 加密时间戳
}

// Coin 币种配置
type Coin struct {
	Symbol  string `json:"symbol"`
	Chain   string `json:"chain"`
	Path    string `json:"path"`
	Address string `json:"address"`
	Index   uint32 `json:"index"`
}

// CoinType 定义支持的币种类型
type CoinType int

const (
	Bitcoin  CoinType = 0
	Ethereum CoinType = 60
	// 可以继续添加其他币种
)

// DerivationPath 派生路径配置
type DerivationPath struct {
	Purpose    uint32 `json:"purpose"`    // 通常为44' (BIP44)
	CoinType   uint32 `json:"coinType"`   // 币种类型
	Account    uint32 `json:"account"`    // 账户索引
	Change     uint32 `json:"change"`     // 0=外部地址，1=找零地址
	AddressIdx uint32 `json:"addressIdx"` // 地址索引
}

// 业务方法 - 体现充血模型特点[1](@ref)
func (w *VirtualWallet) Debit(amount float64) error {
	if w.IsLocked {
		return errors.New("wallet is locked")
	}
	// 业务逻辑封装在领域模型中
	return nil
}

func (w *VirtualWallet) Credit(amount float64) error {
	if w.IsLocked {
		return errors.New("wallet is locked")
	}
	return nil
}

func (w *VirtualWallet) AddCoin(symbol, chain, path, cloak string) error {
	// 派生地址逻辑
	address, err := w.deriveAddress(symbol, path, cloak)
	if err != nil {
		return err
	}

	Coin := Coin{
		Symbol:  symbol,
		Chain:   chain,
		Path:    path,
		Address: address,
	}

	w.Coins = append(w.Coins, Coin)
	return nil
}

func (w *VirtualWallet) Lock() {
	w.IsLocked = true
	w.Mnemonic = "" // 清除敏感数据
}

func (w *VirtualWallet) Unlock(password string) error {
	// 解密逻辑
	w.IsLocked = false
	return nil
}

// deriveAddress 根据币种符号和路径派生地址
func (w *VirtualWallet) deriveAddress(symbol, path, cloak string) (string, error) {
	if w.Mnemonic == "" {
		return "", errors.New("mnemonic is required to derive address")
	}

	// 验证助记词有效性
	if !bip39.IsMnemonicValid(w.Mnemonic) {
		return "", errors.New("invalid mnemonic phrase")
	}

	// 从助记词生成种子
	seed := bip39.NewSeed(w.Mnemonic, cloak) // cloak是魔法语

	// 生成主密钥
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return "", fmt.Errorf("failed to generate master key: %v", err)
	}

	// 解析派生路径
	derivationPath, err := w.parseDerivationPath(symbol, path)
	if err != nil {
		return "", err
	}

	// 根据币种类型派生地址
	switch strings.ToUpper(symbol) {
	case "BTC", "BITCOIN":
		return w.deriveBitcoinAddress(masterKey, derivationPath)
	case "ETH", "ETHEREUM":
		return w.deriveEthereumAddress(masterKey, derivationPath)
	default:
		return "", fmt.Errorf("unsupported Coin symbol: %s", symbol)
	}
}

// parseDerivationPath 解析派生路径
func (w *VirtualWallet) parseDerivationPath(symbol, path string) ([]uint32, error) {
	if path != "" {
		// 解析自定义路径格式 "m/44'/0'/0'/0/0"
		return w.parsePathString(path)
	}

	// 默认使用BIP44标准路径[6](@ref)
	return w.getDefaultPath(symbol)
}

// parsePathString 解析路径字符串
func (w *VirtualWallet) parsePathString(path string) ([]uint32, error) {
	if !strings.HasPrefix(path, "m/") {
		return nil, errors.New("path must start with 'm/'")
	}

	parts := strings.Split(path, "/")[1:] // 去掉"m"
	indices := make([]uint32, len(parts))

	for i, part := range parts {
		hardened := strings.HasSuffix(part, "'")
		cleanPart := strings.TrimSuffix(part, "'")

		var index uint32
		_, err := fmt.Sscanf(cleanPart, "%d", &index)
		if err != nil {
			return nil, fmt.Errorf("invalid path segment: %s", part)
		}

		if hardened {
			index += bip32.FirstHardenedChild
		}
		indices[i] = index
	}

	return indices, nil
}

// getDefaultPath 获取币种的默认BIP44路径[3,6](@ref)
func (w *VirtualWallet) getDefaultPath(symbol string) ([]uint32, error) {
	switch strings.ToUpper(symbol) {
	case "BTC", "BITCOIN":
		// m/44'/0'/0'/0/0
		return []uint32{
			bip32.FirstHardenedChild + 44,
			bip32.FirstHardenedChild + 0, // Bitcoin
			bip32.FirstHardenedChild + 0, // Account 0
			0,                            // External chain (receiving addresses)
			0,                            // Address index
		}, nil
	case "ETH", "ETHEREUM":
		// m/44'/60'/0'/0/0[4,5](@ref)
		return []uint32{
			bip32.FirstHardenedChild + 44,
			bip32.FirstHardenedChild + 60, // Ethereum
			bip32.FirstHardenedChild + 0,  // Account 0
			0,                             // External chain
			0,                             // Address index
		}, nil
	default:
		return nil, fmt.Errorf("no default path for symbol: %s", symbol)
	}
}

// deriveBitcoinAddress 派生比特币地址
func (w *VirtualWallet) deriveBitcoinAddress(masterKey *bip32.Key, path []uint32) (string, error) {
	currentKey := masterKey
	for _, index := range path {
		childKey, err := currentKey.NewChildKey(index)
		if err != nil {
			return "", fmt.Errorf("failed to derive child key: %v", err)
		}
		currentKey = childKey
	}

	// 从派生密钥获取公钥
	publicKey := currentKey.PublicKey().Key

	// 生成比特币地址
	pubKey, err := btcutil.NewAddressPubKey(publicKey, &chaincfg.MainNetParams)
	if err != nil {
		return "", fmt.Errorf("failed to create bitcoin public key: %v", err)
	}

	address := pubKey.AddressPubKeyHash()
	return address.String(), nil
}

// deriveEthereumAddress 派生以太坊地址
func (w *VirtualWallet) deriveEthereumAddress(masterKey *bip32.Key, path []uint32) (string, error) {
	currentKey := masterKey
	for _, index := range path {
		childKey, err := currentKey.NewChildKey(index)
		if err != nil {
			return "", fmt.Errorf("failed to derive child key: %v", err)
		}
		currentKey = childKey
	}

	// 直接使用BIP32密钥作为以太坊私钥
	privateKey, err := crypto.ToECDSA(currentKey.Key)
	if err != nil {
		return "", fmt.Errorf("failed to convert to ECDSA: %v", err)
	}

	// 使用go-ethereum的标准方法生成地址
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return address.Hex(), nil
}

// GenerateNewAddress 为现有币种生成新地址
func (w *VirtualWallet) GenerateNewAddress(symbol string, index uint32, cloak string) (string, error) {
	basePath := "m/44'"
	switch strings.ToUpper(symbol) {
	case "BTC":
		basePath += "/0'/0'/0"
	case "ETH":
		basePath += "/60'/0'/0"
	default:
		return "", fmt.Errorf("unsupported Coin: %s", symbol)
	}

	path := fmt.Sprintf("%s/%d", basePath, index)
	return w.deriveAddress(symbol, path, cloak)
}
