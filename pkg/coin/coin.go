package coin

import (
	"strings"

	"github.com/palagend/slowmade/pkg/logging"
)

// HardenedBit 硬化位标记
const HardenedBit uint32 = 1 << 31

// CoinType 币种类型定义
const (
	CoinTypeBTC uint32 = 0
	CoinTypeETH uint32 = 60
	CoinTypeSOL uint32 = 501
	CoinTypeBNB uint32 = 714
	CoinTypeSUI uint32 = 784
)

// CoinInfo 币种信息
type CoinInfo struct {
	Symbol  string
	Type    uint32 //Coin Type
	Decimal int    // 币种精度
}

// coinRegistry 币种注册表
var coinRegistry = map[uint32]CoinInfo{
	CoinTypeBTC: {"BTC", CoinTypeBTC, 8},
	CoinTypeETH: {"ETH", CoinTypeETH, 18},
	CoinTypeSOL: {"SOL", CoinTypeSOL, 9},
	CoinTypeBNB: {"BNB", CoinTypeBNB, 8},
	CoinTypeSUI: {"SUI", CoinTypeSUI, 9},
}

// symbolToType 符号到类型的映射
var symbolToType = func() map[string]uint32 {
	result := make(map[string]uint32)
	for _, coin := range coinRegistry {
		result[coin.Symbol] = coin.Type
	}
	return result
}()

// CoinSymbol 根据币种类型获取符号
func CoinSymbol(coinType uint32) string {
	// 清除硬化位
	baseType := coinType &^ HardenedBit

	if coin, exists := coinRegistry[baseType]; exists {
		return coin.Symbol
	}
	return ""
}

// CoinType 根据币种符号获取类型
func CoinType(coinSymbol string, hardened bool) uint32 {
	baseType, exists := symbolToType[strings.ToUpper(coinSymbol)]
	if !exists {
		logging.Warnf("%s未注册，返回默认值0", coinSymbol)
		return 0 // 默认返回0
	}

	if hardened {
		return baseType | HardenedBit
	}
	return baseType
}

// GetCoinInfo 获取完整的币种信息
func GetCoinInfo(coinType uint32) (CoinInfo, bool) {
	baseType := coinType &^ HardenedBit
	coin, exists := coinRegistry[baseType]
	return coin, exists
}

// GetAllCoins 获取所有已注册的币种
func GetAllCoins() []CoinInfo {
	coins := make([]CoinInfo, 0, len(coinRegistry))
	for _, coin := range coinRegistry {
		coins = append(coins, coin)
	}
	return coins
}

// RegisterCoin 注册新币种（线程不安全，建议在init中调用）
func RegisterCoin(coinType uint32, symbol string, decimal int) {
	coinRegistry[coinType] = CoinInfo{
		Symbol:  symbol,
		Type:    coinType,
		Decimal: decimal,
	}
	symbolToType[symbol] = coinType
}

// IsHardened 检查是否为硬化类型
func IsHardened(coinType uint32) bool {
	return coinType&HardenedBit != 0
}

// BaseType 获取基础类型（去除硬化位）
func BaseType(coinType uint32) uint32 {
	return coinType &^ HardenedBit
}
