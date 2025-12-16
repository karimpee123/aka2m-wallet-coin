package market

import "coin/internal/market/types"

type CryptoClient interface {
	GetMarketData(currency string) ([]types.Info, error)
}
