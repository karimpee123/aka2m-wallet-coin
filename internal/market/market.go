package market

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"coin/internal/market/binance"
	"coin/internal/market/coingecko"
	"coin/internal/market/coinmarketcap"
	"coin/internal/market/cryptocompare"
	"coin/internal/market/types"
	"coin/pkg/config"
)

type ProviderPrice interface {
	Fetch(provider string, currency string) (any, error)
	GetGeckoMarket(currency string) (any, error)
	GetBinanceMarket(currency string) (map[string]float64, error)
}

type Market struct {
	coinCfg *config.CoinConfig
	apiCfg  *config.APIConfig
	mapCoin map[string]config.Coin
}

var CryptoInstant *Market

func init() {
	coinCfg, err := config.LoadCoinConfig("config/coin.json")
	if err != nil {
		log.Fatalf("Failed to load coin config: %v", err)
	}
	apiCfg, err := config.LoadApiConfig("config/api.json")
	if err != nil {
		log.Fatalf("Failed to load api config: %v", err)
	}
	mapCoin := make(map[string]config.Coin)
	for _, coin := range coinCfg.Coins {
		mapCoin[coin.Symbol] = coin
	}
	CryptoInstant = &Market{coinCfg: coinCfg, apiCfg: apiCfg, mapCoin: mapCoin}
}

func (m *Market) Fetch(provider string, currency string) (any, error) {
	switch provider {
	case "binance":
		return m.GetBinanceMarket(currency)
	case "gecko":
		return m.GetGeckoMarket(currency)
	case "cryptocompare":
		return m.GetCryptoCompareMarket(currency)
	case "coinmarketcap":
		return m.GetCoinMarketCapMarket2(currency)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func (m *Market) GetCryptoCompareMarket(currency string) ([]types.Info, error) {
	client := cryptocompare.NewClient(&m.apiCfg.CryptoCompare)
	return client.GetMarketData(m.coinCfg, currency)
}

func (m *Market) GetCoinMarketCapMarket2(currency string) ([]types.Info, error) {
	client := coinmarketcap.NewClient(&m.apiCfg.CoinMarketCap)
	infoCap, err := client.GetMarketData(m.coinCfg, currency)
	if err != nil {
		return nil, err
	}

	fmt.Println(infoCap)

	var infoList []types.Info
	for _, info := range infoCap {
		id := fmt.Sprintf("%d", info.ID)
		price := info.Quote.USD.Price
		marketCap := info.Quote.USD.MarketCap
		totalVolume := info.Quote.USD.Volume24h

		imgUrl := ""
		symbol := strings.ToUpper(info.Symbol)
		coin, ok := m.mapCoin[symbol]
		if ok {
			imgUrl = coin.ImgURL
		}

		data := types.Info{
			ID:          id,
			Symbol:      info.Symbol,
			Name:        info.Name,
			Image:       imgUrl,
			Price:       price,
			MarketCap:   marketCap,
			TotalVolume: totalVolume,
			Sparkline:   types.Sparkline{},
		}
		infoList = append(infoList, data)
	}
	return infoList, nil
}

func (m *Market) GetGeckoMarket(currency string) (any, error) {
	client := coingecko.NewClient(&m.apiCfg.CoinGecko)
	marketData, err := client.GetMarketData(m.coinCfg, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data from CoinGecko: %v", err)
	}
	result := make(map[string]types.Info)
	for _, d := range marketData {
		result[strings.ToUpper(d.Symbol)] = d
	}
	return marketData, nil
}

func (m *Market) GetBinanceMarket(currency string) ([]types.Info, error) {
	client := binance.NewClient(m.apiCfg.Binance)
	binancePrices, err := client.GetAllPrices()
	if err != nil {
		return nil, fmt.Errorf("failed to get all prices from Binance: %v", err)
	}
	var marketData []types.Info
	for _, coin := range m.coinCfg.Coins {
		binancePair := coin.Symbol + strings.ToUpper(currency)
		binancePriceStr, ok := binancePrices[binancePair]
		if !ok {
			continue
		}
		binancePrice, err := strconv.ParseFloat(binancePriceStr, 64)
		if err != nil {
			continue
		}
		coinData := types.Info{
			ID:          coin.ID,
			Symbol:      coin.Symbol,
			Name:        coin.Name,
			Price:       binancePrice,
			MarketCap:   0,
			TotalVolume: 0,
			Image:       coin.ImgURL,
			Sparkline: struct {
				Price []float64 `json:"price"`
			}{Price: []float64{binancePrice}},
		}
		marketData = append(marketData, coinData)
	}
	return marketData, nil
}
