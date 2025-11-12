package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// KeyService 密钥服务接口
type KeyService interface {
	GenerateMnemonic(strength int) (string, error)
	ValidateMnemonic(mnemonic string) bool
	SeedFromMnemonic(mnemonic, passphrase string) ([]byte, error)
	EntropyToMnemonic(entropy []byte) (string, error)
	MnemonicToEntropy(mnemonic string) ([]byte, error)
}

type keyService struct {
	wordLists map[string][]string
}

// NewKeyService 创建密钥服务实例
func NewKeyService() KeyService {
	return &keyService{
		wordLists: loadWordLists(),
	}
}

// 支持的强度值及其对应的熵长度
var validStrengths = map[int]int{
	128: 16, // 12个单词
	160: 20, // 15个单词
	192: 24, // 18个单词
	224: 28, // 21个单词
	256: 32, // 24个单词
}

// GenerateMnemonic 生成BIP39助记词
func (ks *keyService) GenerateMnemonic(strength int) (string, error) {
	// 验证强度参数
	entropySize, ok := validStrengths[strength]
	if !ok {
		return "", fmt.Errorf("不支持的强度值: %d。支持的强度值: 128, 160, 192, 224, 256", strength)
	}

	// 生成密码学安全的随机熵
	entropy := make([]byte, entropySize)
	if _, err := rand.Read(entropy); err != nil {
		return "", fmt.Errorf("生成随机熵失败: %v", err)
	}

	return ks.EntropyToMnemonic(entropy)
}

// EntropyToMnemonic 将熵转换为助记词
func (ks *keyService) EntropyToMnemonic(entropy []byte) (string, error) {
	if len(entropy) < 16 || len(entropy) > 32 || len(entropy)%4 != 0 {
		return "", errors.New("熵长度必须是16-32字节且是4的倍数")
	}

	// 计算校验和 (取SHA256哈希的前ENT/32位)
	hash := sha256.Sum256(entropy)
	entropyBits := len(entropy) * 8
	checksumBits := entropyBits / 32
	checksum := hash[0] >> (8 - uint(checksumBits))

	// 将熵和校验和合并
	entropyWithChecksum := make([]byte, len(entropy)+1)
	copy(entropyWithChecksum, entropy)
	entropyWithChecksum[len(entropy)] = checksum

	// 转换为二进制位序列
	bits := bytesToBits(entropyWithChecksum)
	totalBits := entropyBits + checksumBits

	// 分割为11位的组并映射到单词表
	wordlist := ks.wordLists["english"]
	var words []string

	for i := 0; i < totalBits; i += 11 {
		end := i + 11
		if end > len(bits) {
			end = len(bits)
		}

		chunk := bits[i:end]
		index := bitsToInt(chunk)

		if index >= len(wordlist) {
			return "", errors.New("单词索引超出范围")
		}

		words = append(words, wordlist[index])
	}

	return strings.Join(words, " "), nil
}

// ValidateMnemonic 验证助记词有效性
func (ks *keyService) ValidateMnemonic(mnemonic string) bool {
	if strings.TrimSpace(mnemonic) == "" {
		return false
	}

	// 检查单词数量 (必须是12, 15, 18, 21, 24)
	words := strings.Fields(mnemonic)
	validLengths := map[int]bool{12: true, 15: true, 18: true, 21: true, 24: true}
	if !validLengths[len(words)] {
		return false
	}

	// 验证所有单词是否在单词表中
	wordlist := ks.wordLists["english"]
	wordMap := make(map[string]bool)
	for _, word := range wordlist {
		wordMap[word] = true
	}

	for _, word := range words {
		if !wordMap[word] {
			return false
		}
	}

	// 验证校验和
	_, err := ks.MnemonicToEntropy(mnemonic)
	return err == nil
}

// MnemonicToEntropy 将助记词转换回熵（包含校验和验证）
func (ks *keyService) MnemonicToEntropy(mnemonic string) ([]byte, error) {
	words := strings.Fields(mnemonic)
	if len(words) < 12 {
		return nil, errors.New("助记词至少需要12个单词")
	}

	wordlist := ks.wordLists["english"]
	wordMap := make(map[string]int)
	for i, word := range wordlist {
		wordMap[word] = i
	}

	// 将单词转换为索引
	var indices []int
	for _, word := range words {
		idx, ok := wordMap[word]
		if !ok {
			return nil, fmt.Errorf("单词不在单词表中: %s", word)
		}
		indices = append(indices, idx)
	}

	// 计算总位数和校验和位数
	totalBits := len(words) * 11
	checksumBits := totalBits / 33
	entropyBits := totalBits - checksumBits

	// 将索引转换为位序列
	bits := make([]bool, totalBits)
	for i, idx := range indices {
		for j := 0; j < 11; j++ {
			bitPos := i*11 + j
			if bitPos < totalBits {
				bits[bitPos] = (idx>>(10-uint(j)))&1 == 1
			}
		}
	}

	// 提取熵位和校验和位
	entropyBytes := bitsToBytes(bits[:entropyBits])
	actualChecksum := bitsToBytes(bits[entropyBits:])[0] >> (8 - uint(checksumBits))

	// 计算期望的校验和
	hash := sha256.Sum256(entropyBytes)
	expectedChecksum := hash[0] >> (8 - uint(checksumBits))

	// 验证校验和
	if actualChecksum != expectedChecksum {
		return nil, errors.New("助记词校验和验证失败")
	}

	return entropyBytes, nil
}

// SeedFromMnemonic 从助记词派生种子
func (ks *keyService) SeedFromMnemonic(mnemonic, passphrase string) ([]byte, error) {
	if !ks.ValidateMnemonic(mnemonic) {
		return nil, errors.New("无效的助记词")
	}

	// 使用PBKDF2派生种子
	// 根据BIP39标准，使用mnemonic作为密码，"mnemonic" + passphrase作为盐
	salt := "mnemonic" + passphrase

	// PBKDF2参数: 迭代次数2048, 输出长度64字节, HMAC-SHA512
	seed := pbkdf2.Key(
		[]byte(mnemonic),
		[]byte(salt),
		2048,
		64,
		sha512.New,
	)

	return seed, nil
}

// 辅助函数

// bytesToBits 将字节转换为位序列
func bytesToBits(data []byte) []bool {
	bits := make([]bool, len(data)*8)
	for i, b := range data {
		for j := 0; j < 8; j++ {
			bits[i*8+j] = (b>>(7-uint(j)))&1 == 1
		}
	}
	return bits
}

// bitsToBytes 将位序列转换为字节
func bitsToBytes(bits []bool) []byte {
	bytes := make([]byte, (len(bits)+7)/8)
	for i, bit := range bits {
		if bit {
			byteIndex := i / 8
			bitIndex := 7 - (i % 8)
			bytes[byteIndex] |= 1 << bitIndex
		}
	}
	return bytes
}

// bitsToInt 将位序列转换为整数
func bitsToInt(bits []bool) int {
	var result int
	for i, bit := range bits {
		if bit {
			result |= 1 << (len(bits) - 1 - i)
		}
	}
	return result
}

// loadWordLists 加载BIP39单词表
func loadWordLists() map[string][]string {
	// 这里只实现英文单词表，实际项目中可以从文件加载多语言支持
	// BIP39标准英文单词表（2048个单词）
	englishWords := []string{
		"abandon", "ability", "able", "about", "above", "absent", "absorb", "abstract",
		"absurd", "abuse", "access", "accident", "account", "accuse", "achieve", "acid",
		// ... 完整的2048个单词
		// 实际实现中需要包含完整的BIP39英文单词表
	}

	return map[string][]string{
		"english": englishWords,
		// 可以添加其他语言支持：chinese_simplified, japanese, spanish等
	}
}
