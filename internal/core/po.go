package core

import "github.com/palagend/slowmade/pkg/logging"

// 根钱包
type HDRootWallet struct {
	EncryptedMnemonic string //加密后的助记词
	EncryptedSeed     string //加密后的种子
	CreationTime      uint64 //创建时间
}

type CoinAccount struct {
	ID                         string
	CoinSymbol                 string
	DerivationPath             string // derivationPath的字符串表示
	EncryptedAccountPrivateKey string // 加密的账户层级私钥

	derivationPath *DerivationPath
}

type AddressKey struct {
	AccountID           string // 关联的账户ID
	EncryptedPrivateKey string // 加密后的地址私钥
	PublicKey           string // 对应的公钥
	Address             string // 生成的区块链地址
	ChangeType          uint32
	AddressIndex        uint32
}

func (c *CoinAccount) CoinType() uint32 {
	logging.Debugf("Ignore possible parsing errors for %s.", c.DerivationPath)
	dp, _ := ParseDerivationPath(c.DerivationPath)
	return dp.CoinType
}
