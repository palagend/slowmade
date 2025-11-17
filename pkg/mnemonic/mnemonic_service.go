// mnemonic/mnemonic_service.go
package mnemonic

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"strings"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/pbkdf2"
)

// MnemonicService 助记词服务接口
type MnemonicService interface {
	GenerateMnemonic(strength int) (string, error)
	ValidateMnemonic(mnemonic string) bool
	GenerateSeedFromMnemonic(mnemonic, password string) []byte
	GenerateEntropy(strength int) ([]byte, error)
}

// BIP39MnemonicService BIP39标准助记词服务实现
type BIP39MnemonicService struct {
	wordList []string
}

// NewBIP39MnemonicService 创建新的BIP39助记词服务
func NewBIP39MnemonicService() *BIP39MnemonicService {
	return &BIP39MnemonicService{
		wordList: loadBIP39WordList(), // 加载BIP39标准单词表
	}
}

// GenerateMnemonic 生成助记词
func (ms *BIP39MnemonicService) GenerateMnemonic(strength int) (string, error) {
	// 强度必须是32的倍数，且在128-256之间
	if strength%32 != 0 || strength < 128 || strength > 256 {
		return "", errors.New("强度必须是128, 160, 192, 224, 或256")
	}

	// 生成熵
	entropy, err := ms.GenerateEntropy(strength / 8)
	if err != nil {
		return "", err
	}

	// 计算校验和
	checksum := ms.calculateChecksum(entropy)

	// 组合熵和校验和
	entropyWithChecksum := append(entropy, checksum)

	// 将熵转换为助记词
	mnemonic := ms.entropyToMnemonic(entropyWithChecksum)

	return mnemonic, nil
}

// GenerateEntropy 生成熵源
func (ms *BIP39MnemonicService) GenerateEntropy(byteSize int) ([]byte, error) {
	entropy := make([]byte, byteSize)
	_, err := rand.Read(entropy)
	return entropy, err
}

// calculateChecksum 计算校验和
func (ms *BIP39MnemonicService) calculateChecksum(entropy []byte) byte {
	hash := sha256.Sum256(entropy)
	// 校验和长度为熵长度/4位
	entropyBits := len(entropy) * 8
	checksumBits := entropyBits / 32
	return hash[0] >> (8 - uint(checksumBits))
}

// entropyToMnemonic 将熵转换为助记词
func (ms *BIP39MnemonicService) entropyToMnemonic(entropy []byte) string {
	var words []string
	bits := ms.bytesToBits(entropy)

	for i := 0; i < len(bits); i += 11 {
		if i+11 > len(bits) {
			break
		}
		index := ms.bitsToInt(bits[i : i+11])
		words = append(words, ms.wordList[index])
	}

	return strings.Join(words, " ")
}

// GenerateSeedFromMnemonic 从助记词生成种子
func (ms *BIP39MnemonicService) GenerateSeedFromMnemonic(mnemonic, password string) []byte {
	// 使用PBKDF2派生种子
	return pbkdf2.Key([]byte(mnemonic), []byte("mnemonic"+password), 2048, 64, sha256.New)
}

// ValidateMnemonic 验证助记词有效性
func (ms *BIP39MnemonicService) ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// 工具方法
func (ms *BIP39MnemonicService) bytesToBits(data []byte) []int {
	bits := make([]int, len(data)*8)
	for i, b := range data {
		for j := 0; j < 8; j++ {
			if b&(1<<uint(7-j)) != 0 {
				bits[i*8+j] = 1
			} else {
				bits[i*8+j] = 0
			}
		}
	}
	return bits
}

func (ms *BIP39MnemonicService) bitsToInt(bits []int) int {
	value := 0
	for i, bit := range bits {
		value += bit << uint(len(bits)-1-i)
	}
	return value
}

func (ms *BIP39MnemonicService) isWordInList(word string) bool {
	for _, w := range ms.wordList {
		if w == word {
			return true
		}
	}
	return false
}

func loadBIP39WordList() []string {
	return bip39.GetWordList()
}
