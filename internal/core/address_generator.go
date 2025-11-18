package core

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/ripemd160" // 需要导入：go get golang.org/x/crypto/ripemd160
)

// 币种特定的地址生成器接口
type AddressGenerator interface {
	GenerateAddress(publicKey []byte) (string, error)
}

// BTC地址生成器
type BTCAddressGenerator struct{}

func (g *BTCAddressGenerator) GenerateAddress(publicKey []byte) (string, error) {
	if len(publicKey) != 33 {
		return "", errors.New("BTC requires compressed public key (33 bytes)")
	}

	// SHA256哈希
	sha256Hash := sha256.Sum256(publicKey)

	// RIPEMD160哈希
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash[:])
	ripemd160Hash := ripemd160Hasher.Sum(nil)

	// 返回十六进制格式的地址（简化版，实际应该进行Base58Check编码）
	return "1" + hex.EncodeToString(ripemd160Hash)[:40], nil
}

// ETH地址生成器
type ETHAddressGenerator struct{}

func (g *ETHAddressGenerator) GenerateAddress(publicKey []byte) (string, error) {
	if len(publicKey) != 64 {
		return "", errors.New("ETH requires 64-byte uncompressed public key")
	}

	// Keccak256哈希（这里用SHA256简化，实际应该用Keccak256）
	hash := sha256.Sum256(publicKey)

	// 取后20字节作为地址
	addressBytes := hash[len(hash)-20:]

	return "0x" + hex.EncodeToString(addressBytes), nil
}

// SOL地址生成器
type SOLAddressGenerator struct{}

func (g *SOLAddressGenerator) GenerateAddress(publicKey []byte) (string, error) {
	if len(publicKey) != 32 {
		return "", errors.New("SOL requires 32-byte public key")
	}

	// Solana使用Base58编码，这里简化返回
	return hex.EncodeToString(publicKey)[:44], nil
}

// BNB地址生成器（类似ETH）
type BNBAddressGenerator struct{}

func (g *BNBAddressGenerator) GenerateAddress(publicKey []byte) (string, error) {
	if len(publicKey) != 64 {
		return "", errors.New("BNB requires 64-byte uncompressed public key")
	}

	hash := sha256.Sum256(publicKey)
	addressBytes := hash[len(hash)-20:]

	return "bnb1" + hex.EncodeToString(addressBytes)[:39], nil
}

// SUI地址生成器
type SUIAddressGenerator struct{}

func (g *SUIAddressGenerator) GenerateAddress(publicKey []byte) (string, error) {
	if len(publicKey) != 32 {
		return "", errors.New("SUI requires 32-byte public key")
	}

	// Sui地址以0x开头，使用特定编码
	return "0x" + hex.EncodeToString(publicKey)[:64], nil
}
