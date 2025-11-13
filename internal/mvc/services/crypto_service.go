package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/palagend/slowmade/internal/mvc/models"
	"golang.org/x/crypto/pbkdf2"
)

// CryptoService 提供加密解密功能
type CryptoService struct {
	encryptionKey []byte
}

// NewCryptoService 创建新的加密服务实例
func NewCryptoService() *CryptoService {
	// TODO 在实际应用中，密钥应该从安全配置中获取
	// 这里使用示例密钥，生产环境需要从环境变量或配置文件中读取
	key := sha256.Sum256([]byte("your-secret-key-here"))
	return &CryptoService{
		encryptionKey: key[:],
	}
}

// EncryptWallet 使用用户密码加密整个钱包
// EncryptWallet 修复后的方法
func (cs *CryptoService) EncryptWallet(wallet *models.VirtualWallet, password string) (*models.VirtualWallet, error) {
	if wallet == nil {
		return nil, errors.New("wallet cannot be nil")
	}
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	// 序列化钱包数据
	walletData, err := json.Marshal(wallet)
	if err != nil {
		return nil, err
	}

	// 生成随机盐值
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// 安全地派生加密密钥
	encryptionKey, err := cs.deriveKey(password, salt, 32)
	if err != nil {
		return nil, err
	}

	// 创建 AES 密码块
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	// 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成正确长度的 Nonce
	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// 使用 GCM 加密
	ciphertext := gcm.Seal(nil, iv, walletData, nil)

	// 创建加密后的钱包
	encryptedWallet := &models.VirtualWallet{
		ID:           wallet.ID,
		Name:         wallet.Name,
		CreationTime: wallet.CreationTime,
		Coins:        wallet.Coins,
		IsEncrypted:  true,
		EncryptionMetadata: models.EncryptionMetadata{
			Algorithm:   "AES-256-GCM",
			Version:     "1.0",
			IV:          base64.StdEncoding.EncodeToString(iv),
			Salt:        base64.StdEncoding.EncodeToString(salt),
			EncryptedAt: time.Now().Unix(),
		},
		EncryptedData: base64.StdEncoding.EncodeToString(ciphertext),
	}

	return encryptedWallet, nil
}

// DecryptWallet 使用密码解密钱包
func (cs *CryptoService) DecryptWallet(encryptedWallet *models.VirtualWallet, password string) (*models.VirtualWallet, error) {
	if encryptedWallet == nil || !encryptedWallet.IsEncrypted {
		return nil, errors.New("wallet is not encrypted or is nil")
	}

	// 1. 解码盐值和IV
	salt, err := base64.StdEncoding.DecodeString(encryptedWallet.EncryptionMetadata.Salt)
	if err != nil {
		return nil, err
	}

	iv, err := base64.StdEncoding.DecodeString(encryptedWallet.EncryptionMetadata.IV)
	if err != nil {
		return nil, err
	}

	// 2. 从密码和盐值派生加密密钥
	encryptionKey, err := cs.deriveKey(password, salt, 32)
	if err != nil {
		return nil, err
	}

	// 3. 解码加密数据
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedWallet.EncryptedData)
	if err != nil {
		return nil, err
	}

	// 4. 使用AES-GCM解密
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// GCM模式解密和认证
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed: invalid password or corrupted data")
	}

	// 5. 反序列化钱包数据
	var wallet models.VirtualWallet
	if err := json.Unmarshal(plaintext, &wallet); err != nil {
		return nil, err
	}

	return &wallet, nil
}

// deriveKey 使用 PBKDF2 安全地派生密钥
func (cs *CryptoService) deriveKey(password string, salt []byte, keyLen int) ([]byte, error) {
	// 使用 PBKDF2 进行密钥派生
	return pbkdf2.Key(
		[]byte(password),
		salt,
		10000, // 迭代次数，可根据安全要求调整
		keyLen,
		sha256.New,
	), nil
}

// ChangeWalletPassword 更改钱包密码
func (cs *CryptoService) ChangeWalletPassword(encryptedWallet *models.VirtualWallet, oldPassword, newPassword string) (*models.VirtualWallet, error) {
	// 1. 先用旧密码解密钱包
	wallet, err := cs.DecryptWallet(encryptedWallet, oldPassword)
	if err != nil {
		return nil, err
	}

	// 2. 用新密码重新加密
	return cs.EncryptWallet(wallet, newPassword)
}

// VerifyWalletPassword 验证钱包密码是否正确
func (cs *CryptoService) VerifyWalletPassword(encryptedWallet *models.VirtualWallet, password string) bool {
	_, err := cs.DecryptWallet(encryptedWallet, password)
	return err == nil
}

// GetWalletInfo 获取钱包加密信息（不暴露敏感数据）
func (cs *CryptoService) GetWalletInfo(encryptedWallet *models.VirtualWallet) map[string]interface{} {
	return map[string]interface{}{
		"id":           encryptedWallet.ID,
		"name":         encryptedWallet.Name,
		"is_encrypted": encryptedWallet.IsEncrypted,
		"algorithm":    encryptedWallet.EncryptionMetadata.Algorithm,
		"version":      encryptedWallet.EncryptionMetadata.Version,
		"encrypted_at": encryptedWallet.EncryptionMetadata.EncryptedAt,
		"Coins":        encryptedWallet.Coins,
	}
}
