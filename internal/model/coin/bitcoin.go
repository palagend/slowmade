package coin

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type Bitcoin struct {
	BaseCoin
}

func NewBitcoin() Coin {
	return &Bitcoin{
		BaseCoin: BaseCoin{
			Name:     "bitcoin",
			Symbol:   "btc",
			CoinType: 0,
		},
	}
}

func (b *Bitcoin) GenerateAddress(pk interface{}) (string, error) {
	pubKey, ok := pk.(ecdsa.PublicKey)
	if !ok {
		log.Fatalf("参数类型不是 *ecdsa.PublicKey, 实际类型为: %T\n", pk)
	}
	pubKeyBytes := elliptic.MarshalCompressed(pubKey.Curve, pubKey.X, pubKey.Y)

	// SHA-256
	sha256Hash := sha256.Sum256(pubKeyBytes)

	// RIPEMD-160
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash[:])
	pubKeyHash := ripemd160Hasher.Sum(nil)

	// 添加版本字节 (主网: 0x00)
	versionedPayload := append([]byte{0x00}, pubKeyHash...)

	// 计算校验和
	checksum := b.checksum(versionedPayload)

	// Base58编码
	fullPayload := append(versionedPayload, checksum...)
	address := base58.Encode(fullPayload)

	return address, nil
}

func (b *Bitcoin) ValidateAddress(address string) bool {
	decoded := base58.Decode(address)
	if len(decoded) != 25 {
		return false
	}

	version := decoded[0]
	checksum := decoded[len(decoded)-4:]
	payload := decoded[:len(decoded)-4]

	expectedChecksum := b.checksum(payload)
	return version == 0x00 && hex.EncodeToString(checksum) == hex.EncodeToString(expectedChecksum[:])
}

func (b *Bitcoin) GetHDPath(account, change, index uint32) string {
	return fmt.Sprintf("m/44'/%d'/%d'/%d/%d", b.GetCoinType(), account, change, index)
}

func (b *Bitcoin) GetBIP44Path(account, change, index uint32) string {
	return fmt.Sprintf("m/44'/%d'/%d'/%d/%d", b.GetCoinType(), account, change, index)
}

func (b *Bitcoin) checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:4]
}
