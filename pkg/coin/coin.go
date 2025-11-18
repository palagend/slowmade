package coin

// 常量定义 - BIP44标准币种类型
const (
	HardenedBit uint32 = 1 << 31
	CoinTypeBTC uint32 = 0
	CoinTypeETH uint32 = 60
	CoinTypeSOL uint32 = 501
	CoinTypeBNB uint32 = 714
	CoinTypeSUI uint32 = 784
)

func CoinSymbol(coinType uint32) string {
	switch coinType {
	case CoinTypeBTC, CoinTypeBTC | HardenedBit:
		return "BTC"
	case CoinTypeETH, CoinTypeETH | HardenedBit:
		return "ETH"
	case CoinTypeSOL, CoinTypeSOL | HardenedBit:
		return "SOL"
	case CoinTypeBNB, CoinTypeBNB | HardenedBit:
		return "BNB"
	case CoinTypeSUI, CoinTypeSUI | HardenedBit:
		return "SUI"
	default:
		return ""
	}

}
