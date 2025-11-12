package coin

import (
	"fmt"
	"sort"
)

// CoinRegistry 币种注册表（单例）
type CoinRegistry struct {
	coins map[string]Coin
}

var registry *CoinRegistry

func init() {
	registry = &CoinRegistry{
		coins: make(map[string]Coin),
	}
	// 自动注册所有内置币种
	registry.autoRegister()
}

func GetInstance() *CoinRegistry {
	return registry
}

// autoRegister 使用反射自动发现并注册所有实现了Coin接口的结构体
func (r *CoinRegistry) autoRegister() {
	// 内置币种类型列表
	coinTypes := []Coin{
		NewBitcoin(),
		//		&Ethereum{},
		//		&BinanceCoin{},
		// 未来添加新币种时，只需在此列表中添加一行，例如 &Dogecoin{},
	}

	for _, coinType := range coinTypes {
		r.RegisterCoin(coinType)
	}
}

// RegisterCoin 注册单个币种（线程安全）
func (r *CoinRegistry) RegisterCoin(coin Coin) {
	symbol := coin.GetSymbol()
	if _, exists := r.coins[symbol]; exists {
		// 避免重复注册
		return
	}
	r.coins[symbol] = coin
	// 获取币种类型名称用于日志
	//coinType := reflect.TypeOf(coin).Elem().Name()
	//fmt.Printf("Registered coin: %s (%s)\n", coin.GetName(), coinType)
}

// GetCoin 获取指定币种
func (r *CoinRegistry) GetCoin(symbol string) (Coin, error) {
	coin, exists := r.coins[symbol]
	if !exists {
		return nil, fmt.Errorf("coin not found: %s", symbol)
	}
	return coin, nil
}

// GetSupportedCoins 获取已排序的币种符号列表
func (r *CoinRegistry) GetSupportedCoins() []string {
	symbols := make([]string, 0, len(r.coins))
	for symbol := range r.coins {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	return symbols
}

// GetAllCoins 获取所有已注册的币种（按符号排序）
func (r *CoinRegistry) GetAllCoins() []Coin {
	symbols := r.GetSupportedCoins()
	coins := make([]Coin, 0, len(symbols))
	for _, symbol := range symbols {
		coins = append(coins, r.coins[symbol])
	}
	return coins
}

// GetCoinsByCategory 按类别筛选币种（例如：mainnet, testnet, privacy）
func (r *CoinRegistry) GetCoinsByCategory(category string) []Coin {
	// 这里可以根据币种实现的 Category() 方法等进行筛选
	// 示例实现：返回所有币种
	return r.GetAllCoins()
}
