package coin

// Coin 币种接口
type Coin interface {
	// 基本信息
	GetName() string
	GetSymbol() string
	GetCoinType() uint32

	// 高级功能
	GenerateAddress(pubKey interface{}) (string, error)
	ValidateAddress(address string) bool
	GetHDPath(account, change, index uint32) string

	// 分类和状态信息
	GetCategory() string // 如: mainnet, testnet, privacy
	GetStatus() string   // 如: active, beta, deprecated

	// 支持的功能特性
	SupportsSegwit() bool
	SupportsSmartContracts() bool
	IsPrivacyFocused() bool
}

// BaseCoin 基础实现，可被具体币种嵌入
type BaseCoin struct {
	Name     string
	Symbol   string
	CoinType uint32
	Category string
	Status   string
}

func (b *BaseCoin) GetName() string     { return b.Name }
func (b *BaseCoin) GetSymbol() string   { return b.Symbol }
func (b *BaseCoin) GetCoinType() uint32 { return b.CoinType }
func (b *BaseCoin) GetCategory() string {
	if b.Category == "" {
		return "mainnet"
	}
	return b.Category
}
func (b *BaseCoin) GetStatus() string {
	if b.Status == "" {
		return "active"
	}
	return b.Status
}
func (b *BaseCoin) SupportsSegwit() bool {
	return false
}
func (b *BaseCoin) SupportsSmartContracts() bool {
	return false
}
func (b *BaseCoin) IsPrivacyFocused() bool {
	return false
}
