package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/scrypt"
)

// 加密服务接口
type CryptoService interface {
	Encrypt(plaintext []byte, password string) (string, error)
	Decrypt(ciphertext string, password string) ([]byte, error)
	GetAlgorithm() string
}

// 密钥派生函数接口
type KDF interface {
	DeriveKey(password string, salt []byte) ([]byte, error)
	GetName() string
}

// 错误定义
var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrDecryptionFailed  = errors.New("decryption failed")
	ErrInvalidPassword   = errors.New("invalid password")
)

// ==================== 密钥派生函数实现 ====================

// Scrypt KDF
type ScryptKDF struct {
	N       int
	R       int
	P       int
	KeyLen  int
	SaltLen int
}

func NewScryptKDF() *ScryptKDF {
	return &ScryptKDF{
		N:       32768, // 适合钱包加密的标准参数
		R:       8,
		P:       1,
		KeyLen:  32,
		SaltLen: 16,
	}
}

func (s *ScryptKDF) DeriveKey(password string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(password), salt, s.N, s.R, s.P, s.KeyLen)
}

func (s *ScryptKDF) GetName() string {
	return "scrypt"
}

// Argon2 KDF
type Argon2KDF struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen int
}

func NewArgon2KDF() *Argon2KDF {
	return &Argon2KDF{
		Time:    3,
		Memory:  64 * 1024, // 64MB
		Threads: 4,
		KeyLen:  32,
		SaltLen: 16,
	}
}

func (a *Argon2KDF) DeriveKey(password string, salt []byte) ([]byte, error) {
	return argon2.IDKey([]byte(password), salt, a.Time, a.Memory, a.Threads, a.KeyLen), nil
}

func (a *Argon2KDF) GetName() string {
	return "argon2"
}

// PBKDF2 (使用SHA256)
type PBKDF2SHA256 struct {
	Iterations int
	KeyLen     int
	SaltLen    int
}

func NewPBKDF2SHA256() *PBKDF2SHA256 {
	return &PBKDF2SHA256{
		Iterations: 100000, // 适合钱包加密的迭代次数
		KeyLen:     32,
		SaltLen:    16,
	}
}

func (p *PBKDF2SHA256) DeriveKey(password string, salt []byte) ([]byte, error) {
	// 简化实现，实际使用时可以使用标准的PBKDF2实现
	key := sha256.Sum256([]byte(password))
	for i := 1; i < p.Iterations; i++ {
		key = sha256.Sum256(key[:])
	}
	return key[:p.KeyLen], nil
}

func (p *PBKDF2SHA256) GetName() string {
	return "pbkdf2-sha256"
}

// ==================== 加密服务实现 ====================

// AES-GCM 加密服务
type AESGCMService struct {
	kdf       KDF
	nonceSize int
}

func NewAESGCMService(kdf KDF) *AESGCMService {
	return &AESGCMService{
		kdf:       kdf,
		nonceSize: 12, // GCM推荐的非ce大小
	}
}

func (a *AESGCMService) Encrypt(plaintext []byte, password string) (string, error) {
	// 生成盐
	salt := make([]byte, a.getSaltLen())
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// 派生密钥
	key, err := a.kdf.DeriveKey(password, salt)
	if err != nil {
		return "", err
	}

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// 组合结果: salt + ciphertext
	result := append(salt, ciphertext...)
	return hex.EncodeToString(result), nil
}

func (a *AESGCMService) Decrypt(encodedCiphertext string, password string) ([]byte, error) {
	// 解码hex
	data, err := hex.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	saltLen := a.getSaltLen()
	if len(data) < saltLen+a.nonceSize {
		return nil, ErrInvalidCiphertext
	}

	// 提取salt和密文
	salt := data[:saltLen]
	ciphertext := data[saltLen:]

	// 派生密钥
	key, err := a.kdf.DeriveKey(password, salt)
	if err != nil {
		return nil, err
	}

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 提取nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

func (a *AESGCMService) GetAlgorithm() string {
	return fmt.Sprintf("AES-GCM-256 with %s", a.kdf.GetName())
}

func (a *AESGCMService) getSaltLen() int {
	switch kdf := a.kdf.(type) {
	case *ScryptKDF:
		return kdf.SaltLen
	case *Argon2KDF:
		return kdf.SaltLen
	case *PBKDF2SHA256:
		return kdf.SaltLen
	default:
		return 16
	}
}

// ChaCha20-Poly1305 加密服务
type ChaCha20Poly1305Service struct {
	kdf KDF
}

func NewChaCha20Poly1305Service(kdf KDF) *ChaCha20Poly1305Service {
	return &ChaCha20Poly1305Service{kdf: kdf}
}

func (c *ChaCha20Poly1305Service) Encrypt(plaintext []byte, password string) (string, error) {
	// 生成盐
	salt := make([]byte, c.getSaltLen())
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// 派生密钥
	key, err := c.kdf.DeriveKey(password, salt)
	if err != nil {
		return "", err
	}

	// 创建ChaCha20-Poly1305
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return "", err
	}

	// 生成nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	// 组合结果: salt + nonce + ciphertext
	result := append(salt, nonce...)
	result = append(result, ciphertext...)
	return hex.EncodeToString(result), nil
}

func (c *ChaCha20Poly1305Service) Decrypt(encodedCiphertext string, password string) ([]byte, error) {
	data, err := hex.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	saltLen := c.getSaltLen()
	nonceSize := chacha20poly1305.NonceSizeX
	if len(data) < saltLen+nonceSize {
		return nil, ErrInvalidCiphertext
	}

	// 提取组件
	salt := data[:saltLen]
	nonce := data[saltLen : saltLen+nonceSize]
	ciphertext := data[saltLen+nonceSize:]

	// 派生密钥
	key, err := c.kdf.DeriveKey(password, salt)
	if err != nil {
		return nil, err
	}

	// 创建AEAD
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// 解密
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

func (c *ChaCha20Poly1305Service) GetAlgorithm() string {
	return fmt.Sprintf("ChaCha20-Poly1305 with %s", c.kdf.GetName())
}

func (c *ChaCha20Poly1305Service) getSaltLen() int {
	switch kdf := c.kdf.(type) {
	case *ScryptKDF:
		return kdf.SaltLen
	case *Argon2KDF:
		return kdf.SaltLen
	case *PBKDF2SHA256:
		return kdf.SaltLen
	default:
		return 16
	}
}

// ==================== 密钥派生工厂 ====================

type KDFType string

const (
	KDFScrypt KDFType = "scrypt"
	KDFArgon2 KDFType = "argon2"
	KDFPBKDF2 KDFType = "pbkdf2"
)

// KDF工厂
type KDFFactory struct{}

func NewKDFFactory() *KDFFactory {
	return &KDFFactory{}
}

func (f *KDFFactory) CreateKDF(kdfType KDFType) KDF {
	switch kdfType {
	case KDFScrypt:
		return NewScryptKDF()
	case KDFArgon2:
		return NewArgon2KDF()
	case KDFPBKDF2:
		return NewPBKDF2SHA256()
	default:
		return NewScryptKDF() // 默认使用scrypt
	}
}

// ==================== 加密服务工厂 ====================

type EncryptionType string

const (
	EncryptionAESGCM           EncryptionType = "aes-gcm"
	EncryptionChaCha20Poly1305 EncryptionType = "chacha20-poly1305"
)

// 加密服务工厂
type CryptoServiceFactory struct {
	kdfFactory *KDFFactory
}

func NewCryptoServiceFactory() *CryptoServiceFactory {
	return &CryptoServiceFactory{
		kdfFactory: NewKDFFactory(),
	}
}

// 创建默认的加密服务（适合加密货币钱包）
func (f *CryptoServiceFactory) CreateDefault() CryptoService {
	// 对于加密货币钱包，推荐使用AES-GCM + Scrypt组合
	kdf := f.kdfFactory.CreateKDF(KDFScrypt)
	return NewAESGCMService(kdf)
}

// 创建特定类型的加密服务
func (f *CryptoServiceFactory) CreateService(encType EncryptionType, kdfType KDFType) CryptoService {
	kdf := f.kdfFactory.CreateKDF(kdfType)

	switch encType {
	case EncryptionAESGCM:
		return NewAESGCMService(kdf)
	case EncryptionChaCha20Poly1305:
		return NewChaCha20Poly1305Service(kdf)
	default:
		return f.CreateDefault()
	}
}

// ==================== 高级功能：密钥派生参数配置 ====================

// 可配置的KDF工厂，允许自定义参数
type ConfigurableKDFFactory struct{}

func NewConfigurableKDFFactory() *ConfigurableKDFFactory {
	return &ConfigurableKDFFactory{}
}

func (f *ConfigurableKDFFactory) CreateScryptWithParams(N, r, p, keyLen, saltLen int) KDF {
	return &ScryptKDF{
		N:       N,
		R:       r,
		P:       p,
		KeyLen:  keyLen,
		SaltLen: saltLen,
	}
}

func (f *ConfigurableKDFFactory) CreateArgon2WithParams(time, memory uint32, threads uint8, keyLen uint32, saltLen int) KDF {
	return &Argon2KDF{
		Time:    time,
		Memory:  memory,
		Threads: threads,
		KeyLen:  keyLen,
		SaltLen: saltLen,
	}
}

// ==================== 单例模式实现 ====================

// 全局单例实例
var (
	cryptoServiceInstance      CryptoService
	cryptoServiceFactoryOnce   sync.Once
	cryptoServiceFactory       *CryptoServiceFactory
	configurableKDFFactory     *ConfigurableKDFFactory
	kdfFactoryInstance         *KDFFactory
	kdfFactoryOnce             sync.Once
	configurableKDFFactoryOnce sync.Once
)

// CryptoManager 加密管理器（单例）
type CryptoManager struct {
	factory *CryptoServiceFactory
}

// GetCryptoServiceFactory 获取加密服务工厂单例
func GetCryptoServiceFactory() *CryptoServiceFactory {
	cryptoServiceFactoryOnce.Do(func() {
		cryptoServiceFactory = NewCryptoServiceFactory()
	})
	return cryptoServiceFactory
}

// GetKDFFactory 获取KDF工厂单例
func GetKDFFactory() *KDFFactory {
	kdfFactoryOnce.Do(func() {
		kdfFactoryInstance = NewKDFFactory()
	})
	return kdfFactoryInstance
}

// GetConfigurableKDFFactory 获取可配置KDF工厂单例
func GetConfigurableKDFFactory() *ConfigurableKDFFactory {
	configurableKDFFactoryOnce.Do(func() {
		configurableKDFFactory = NewConfigurableKDFFactory()
	})
	return configurableKDFFactory
}

// GetDefaultCryptoService 获取默认加密服务单例
func GetDefaultCryptoService() CryptoService {
	if cryptoServiceInstance == nil {
		cryptoServiceInstance = GetCryptoServiceFactory().CreateDefault()
	}
	return cryptoServiceInstance
}

// SetGlobalCryptoService 设置全局加密服务实例（用于测试或自定义配置）
func SetGlobalCryptoService(service CryptoService) {
	cryptoServiceInstance = service
}

// ResetGlobalCryptoService 重置全局加密服务实例（主要用于测试）
func ResetGlobalCryptoService() {
	cryptoServiceInstance = nil
}

// ==================== 便捷函数 ====================

// Encrypt 使用默认加密服务加密数据
func Encrypt(plaintext []byte, password string) (string, error) {
	return GetDefaultCryptoService().Encrypt(plaintext, password)
}

// Decrypt 使用默认加密服务解密数据
func Decrypt(ciphertext string, password string) ([]byte, error) {
	return GetDefaultCryptoService().Decrypt(ciphertext, password)
}

// GetCurrentAlgorithm 获取当前使用的算法信息
func GetCurrentAlgorithm() string {
	return GetDefaultCryptoService().GetAlgorithm()
}

// CreateCustomCryptoService 创建自定义加密服务（非单例）
func CreateCustomCryptoService(encType EncryptionType, kdfType KDFType) CryptoService {
	return GetCryptoServiceFactory().CreateService(encType, kdfType)
}

// CreateCryptoServiceWithCustomKDF 使用自定义KDF参数创建加密服务
func CreateCryptoServiceWithCustomKDF(encType EncryptionType, kdf KDF) CryptoService {
	switch encType {
	case EncryptionAESGCM:
		return NewAESGCMService(kdf)
	case EncryptionChaCha20Poly1305:
		return NewChaCha20Poly1305Service(kdf)
	default:
		return NewAESGCMService(kdf)
	}
}

func EncryptData(d []byte, password string) (string, error) {
	return GetDefaultCryptoService().Encrypt(d, password)
}

func DecryptData(s string, passowrd string) ([]byte, error) {
	return GetDefaultCryptoService().Decrypt(s, passowrd)
}
