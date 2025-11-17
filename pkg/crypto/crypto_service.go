// crypto/crypto_service.go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// CryptoService 加密服务接口
type CryptoService interface {
	EncryptData(data []byte, password string) ([]byte, error)
	DecryptData(encryptedData []byte, password string) ([]byte, error)
	DeriveKey(password string, salt []byte) []byte
	GenerateRandomBytes(n int) ([]byte, error)
}

// AESCryptoService AES加密服务实现
type AESCryptoService struct {
	salt []byte
}

// NewAESCryptoService 创建新的AES加密服务实例
func NewAESCryptoService(salt []byte) *AESCryptoService {
	return &AESCryptoService{salt: salt}
}

// EncryptData 加密数据
func (cs *AESCryptoService) EncryptData(data []byte, password string) ([]byte, error) {
	key := cs.DeriveKey(password, cs.salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encrypted := gcm.Seal(nonce, nonce, data, nil)
	return encrypted, nil
}

// DecryptData 解密数据 [2](@ref)
func (cs *AESCryptoService) DecryptData(encryptedData []byte, password string) ([]byte, error) {
	key := cs.DeriveKey(password, cs.salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("加密数据过短")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// DeriveKey 从密码派生密钥 [4,9](@ref)
func (cs *AESCryptoService) DeriveKey(password string, salt []byte) []byte {
	// 使用PBKDF2进行密钥派生，增加破解难度
	if salt == nil {
		salt = make([]byte, 32)
		rand.Read(salt)
	}
	return pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
}

// GenerateRandomBytes 生成随机字节
func (cs *AESCryptoService) GenerateRandomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	return bytes, err
}

// secureWipeBytes 用随机数据或零值覆盖字节切片，然后截断。
func SecureWipeBytes(data []byte) {
	if data == nil {
		return
	}
	// 方法1：使用零值覆盖
	for i := range data {
		data[i] = 0
	}
	// 方法2（增强）：考虑使用随机数据覆盖，增加数据恢复难度
	// const zero = 0
	// rand.Read(data) // 先用随机数覆盖
	// for i := range data { // 再用0覆盖
	//  data[i] = zero
	// }

	// 重要：此操作直接修改底层数组，与创建新切片不同[6](@ref)。
}
