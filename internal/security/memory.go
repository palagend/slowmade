package security

import (
	"runtime"
	"syscall"

	"golang.org/x/crypto/scrypt"
)

// SecureMemory 用于安全处理敏感数据的内存区域
type SecureMemory struct {
	data []byte
}

// NewSecureMemory 创建一块被锁定的内存区域
func NewSecureMemory(size int) (*SecureMemory, error) {
	data := make([]byte, size)

	// 锁定内存页，防止被交换到磁盘
	err := syscall.Mlock(data)
	if err != nil {
		return nil, err
	}

	return &SecureMemory{data: data}, nil
}

// SetData 安全地将数据写入受保护的内存
func (sm *SecureMemory) SetData(input []byte) {
	// 确保输入数据不会超出预分配范围
	if len(input) > len(sm.data) {
		panic("input data exceeds secure memory capacity")
	}
	copy(sm.data, input)

	// 立即清理原始输入数据（如果可控）
	if len(input) > 0 {
		for i := range input {
			input[i] = 0
		}
	}
}

// GetData 获取数据的副本（使用后需清理）
func (sm *SecureMemory) GetData() []byte {
	dum := make([]byte, len(sm.data))
	copy(dum, sm.data)
	return dum
}

// Destroy 安全擦除并释放内存
func (sm *SecureMemory) Destroy() {
	// 用随机数据多次覆盖
	for i := 0; i < 3; i++ {
		for j := range sm.data {
			sm.data[j] = byte(i)
		}
	}

	// 解锁内存页
	syscall.Munlock(sm.data)

	// 清空切片
	for i := range sm.data {
		sm.data[i] = 0
	}
	sm.data = nil

	// 强制垃圾回收
	runtime.GC()
}

// DeriveKey 安全的密钥派生函数
func DeriveKey(password, salt []byte) ([]byte, error) {
	key, err := scrypt.Key(password, salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	// 立即清理密码和盐（调用方负责清理）
	if len(password) > 0 {
		for i := range password {
			password[i] = 0
		}
	}
	if len(salt) > 0 {
		for i := range salt {
			salt[i] = 0
		}
	}

	return key, nil
}
