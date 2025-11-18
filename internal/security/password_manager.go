// internal/security/password_manager.go
package security

import (
	"sync"

	"github.com/awnumar/memguard"
)

var (
	instance *PasswordManager
	once     sync.Once
)

// PasswordManager 密码安全管理器
type PasswordManager struct {
	mu       sync.RWMutex
	enclave  *memguard.Enclave
	isSealed bool
}

// GetPasswordManager 获取密码管理器单例实例
func GetPasswordManager() *PasswordManager {
	once.Do(func() {
		instance = &PasswordManager{
			isSealed: true, // 初始状态为已锁定
		}
	})
	return instance
}

// ResetPasswordManagerInstance 重置单例实例（主要用于测试）
func ResetPasswordManagerInstance() {
	instance = nil
	once = sync.Once{}
}

// SetPassword 安全设置密码
func (pm *PasswordManager) SetPassword(password string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 清空现有密码
	pm.unsafeClear()

	// 将密码转换为字节并创建安全缓冲区
	passwordBytes := []byte(password)
	lockedBuffer := memguard.NewBufferFromBytes(passwordBytes)

	// 立即密封到安全 enclave
	pm.enclave = lockedBuffer.Seal()

	// 销毁原始缓冲区
	lockedBuffer.Destroy()

	// 安全清空输入数据
	pm.secureWipe(passwordBytes)

	pm.isSealed = false

	return nil
}

// GetPassword 安全获取密码（返回副本）
func (pm *PasswordManager) GetPassword() ([]byte, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.isSealed || pm.enclave == nil {
		return nil, ErrPasswordNotSet
	}

	// 从 enclave 解密获取密码
	unsealed, err := pm.enclave.Open()
	if err != nil {
		return nil, err
	}
	defer unsealed.Destroy() // 使用后立即销毁

	// 返回密码的副本
	passwordCopy := make([]byte, unsealed.Size())
	copy(passwordCopy, unsealed.Bytes())

	return passwordCopy, nil
}

// VerifyPassword 验证密码（不暴露密码内容）
func (pm *PasswordManager) VerifyPassword(input string) (bool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.isSealed || pm.enclave == nil {
		return false, ErrPasswordNotSet
	}

	// 获取存储的密码进行比较
	stored, err := pm.enclave.Open()
	if err != nil {
		return false, err
	}
	defer stored.Destroy()

	// 使用恒定时间比较防止时序攻击
	inputBytes := []byte(input)
	defer pm.secureWipe(inputBytes) // 安全清空输入

	return pm.constantTimeCompare(stored.Bytes(), inputBytes), nil
}

// Clear 安全清空密码
func (pm *PasswordManager) Clear() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.unsafeClear()
}

// IsSet 检查是否设置了密码
func (pm *PasswordManager) IsSet() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return !pm.isSealed && pm.enclave != nil
}

// 内部方法
func (pm *PasswordManager) unsafeClear() {
	if pm.enclave != nil {
		// 尝试打开并安全销毁内容
		if unsealed, err := pm.enclave.Open(); err == nil {
			unsealed.Destroy()
		}
		pm.enclave = nil
	}
	pm.isSealed = true
}

func (pm *PasswordManager) secureWipe(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

func (pm *PasswordManager) constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

func WipeSensitiveData(data []byte) {
	GetPasswordManager().secureWipe(data)
}

func Password() ([]byte, error) {
	return GetPasswordManager().GetPassword()
}
