package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io"

	"golang.org/x/crypto/scrypt"
)

// SecureKeystore 处理钱包文件的加密存储
type SecureKeystore struct{}

// EncryptedWallet 加密的钱包文件结构
type EncryptedWallet struct {
	CipherText    []byte `json:"c"`
	Nonce         []byte `json:"n"`
	Salt          []byte `json:"s"`
	KDFIterations int    `json:"i"`
	Version       string `json:"v"`
}

func NewKeystore() *SecureKeystore {
	return &SecureKeystore{}
}

// EncryptWallet 加密钱包数据
func (ks *SecureKeystore) EncryptWallet(walletData []byte, password string) ([]byte, error) {
	// 生成随机盐值
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// 派生加密密钥
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	defer ks.wipeMemory(key) // 确保密钥最终被清理

	// 创建AES-GCM加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成随机Nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密数据
	ciphertext := gcm.Seal(nil, nonce, walletData, nil)

	// 创建加密钱包结构
	encrypted := &EncryptedWallet{
		CipherText:    ciphertext,
		Nonce:         nonce,
		Salt:          salt,
		KDFIterations: 32768,
		Version:       "1.0",
	}

	return json.Marshal(encrypted)
}

// DecryptWallet 解密钱包数据
func (ks *SecureKeystore) DecryptWallet(encryptedData []byte, password string) ([]byte, error) {
	var encrypted EncryptedWallet
	if err := json.Unmarshal(encryptedData, &encrypted); err != nil {
		return nil, err
	}

	// 派生加密密钥
	key, err := scrypt.Key([]byte(password), encrypted.Salt, encrypted.KDFIterations, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	defer ks.wipeMemory(key) // 确保密钥最终被清理

	// 创建AES-GCM解密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 解密数据
	plaintext, err := gcm.Open(nil, encrypted.Nonce, encrypted.CipherText, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// wipeMemory 安全擦除内存
func (ks *SecureKeystore) wipeMemory(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
