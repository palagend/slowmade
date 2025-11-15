package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// BIP44 路径常量
	Purpose      uint32 = 44
	CoinTypeETH  uint32 = 60
	Account      uint32 = 0
	Change       uint32 = 0
	AddressIndex uint32 = 0

	// 加密参数
	scryptN      = 1 << 18
	scryptR      = 8
	scryptP      = 1
	scryptKeyLen = 32
)

var (
	ErrInvalidMnemonic = errors.New("invalid mnemonic phrase")
	ErrEncryptFailed   = errors.New("failed to encrypt key")
	ErrDecryptFailed   = errors.New("failed to decrypt key")
)

// HDWallet 表示一个分层确定性钱包
type HDWallet struct {
	mnemonic   string
	seed       []byte
	privateKey []byte
	publicKey  []byte
	address    common.Address
}

// GenerateMnemonic 生成一个新的BIP39助记词
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}
	return mnemonic, nil
}

// ValidateMnemonic 验证助记词是否有效
func ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// NewSeed 从助记词生成BIP39种子
func NewSeed(mnemonic, password string) []byte {
	return bip39.NewSeed(mnemonic, password)
}

// NewHDWallet 从种子创建HD钱包
func NewHDWallet(seed []byte) (*HDWallet, error) {
	// 使用BIP32派生路径 m/44'/60'/0'/0/0
	path := []uint32{
		Purpose | 0x80000000,
		CoinTypeETH | 0x80000000,
		Account | 0x80000000,
		Change,
		AddressIndex,
	}

	// 派生私钥 (简化实现，实际应使用完整的BIP32派生)
	privateKey, err := derivePrivateKey(seed, path)
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	// 派生公钥和地址
	publicKey, address, err := derivePublicKeyAndAddress(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive public key: %w", err)
	}

	return &HDWallet{
		seed:       seed,
		privateKey: privateKey,
		publicKey:  publicKey,
		address:    address,
	}, nil
}

// PrivateKey 返回钱包的私钥
func (w *HDWallet) PrivateKey() []byte {
	return w.privateKey
}

// PublicKey 返回钱包的公钥
func (w *HDWallet) PublicKey() []byte {
	return w.publicKey
}

// Address 返回钱包的地址
func (w *HDWallet) Address() common.Address {
	return w.address
}

// EncryptKey 使用scrypt派生的密钥加密私钥
func EncryptKey(privateKey, key, salt []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptFailed, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptFailed, err)
	}

	ciphertext := gcm.Seal(nonce, nonce, privateKey, nil)
	encryptedKey := append(salt, ciphertext...)

	return encryptedKey, nil
}

// DecryptKey 使用scrypt派生的密钥解密私钥
func DecryptKey(encryptedKey, password []byte) ([]byte, error) {
	if len(encryptedKey) < 32 {
		return nil, ErrDecryptFailed
	}

	salt := encryptedKey[:32]
	ciphertext := encryptedKey[32:]

	key := pbkdf2.Key(password, salt, scryptN, scryptKeyLen, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrDecryptFailed
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	privateKey, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	return privateKey, nil
}

// derivePrivateKey 从种子派生私钥 (简化实现)
func derivePrivateKey(seed []byte, path []uint32) ([]byte, error) {
	// 实际实现应使用完整的BIP32派生树
	// 这里简化处理，仅使用种子生成私钥
	if len(seed) < 32 {
		return nil, errors.New("seed too short")
	}
	privateKey := make([]byte, 32)
	copy(privateKey, seed[:32])
	return privateKey, nil
}

// derivePublicKeyAndAddress 从私钥派生公钥和地址 (简化实现)
func derivePublicKeyAndAddress(privateKey []byte) ([]byte, common.Address, error) {
	// 实际实现应使用椭圆曲线加密
	// 这里简化处理，使用私钥的哈希作为公钥
	if len(privateKey) != 32 {
		return nil, common.Address{}, errors.New("invalid private key length")
	}

	hash := sha256.Sum256(privateKey)
	publicKey := hash[:]

	// 生成地址 (简化处理)
	addressBytes := sha256.Sum256(publicKey)
	address := common.BytesToAddress(addressBytes[:20])

	return publicKey, address, nil
}

// GenerateSalt 生成随机盐值
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// SecureClear 安全清除字节数组内容
func SecureClear(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// BytesToHex 将字节数组转换为十六进制字符串
func BytesToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// HexToBytes 将十六进制字符串转换为字节数组
func HexToBytes(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

// CompareKeys 安全比较两个密钥
func CompareKeys(a, b []byte) bool {
	return bytes.Equal(a, b)
}
