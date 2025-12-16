package cryptocompare

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"coin/internal/market/types"
	"coin/pkg/config"
)

type ClientCryptoCompare struct {
	BaseURL     string
	MarketURL   string
	apiKey      string
	cache       []types.Info
	lastFetched time.Time
	cacheTTL    time.Duration
	mu          sync.RWMutex
}

func NewClient(cfg *config.CryptoCompareConfig) *ClientCryptoCompare {
	return &ClientCryptoCompare{
		BaseURL:   cfg.BaseURL,
		MarketURL: cfg.MarketURL,
		apiKey:    cfg.ApiKey,
		cache:     []types.Info{},
		cacheTTL:  1 * time.Minute,
	}
}

func (c *ClientCryptoCompare) GetMarketData(coinCfg *config.CoinConfig, currency string) ([]types.Info, error) {
	c.mu.RLock()
	if time.Since(c.lastFetched) < c.cacheTTL {
		data := c.cache
		c.mu.RUnlock()
		return data, nil
	}
	c.mu.RUnlock()

	ids := ""
	for i, coin := range coinCfg.Coins {
		if i > 0 {
			ids += ","
		}
		ids += coin.ID
	}
	url := fmt.Sprintf("%s%s?fsyms=%s&tsyms=%s", c.BaseURL, c.MarketURL, ids, currency)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Apikey "+c.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CryptoCompare API returned status %d", resp.StatusCode)
	}
	var result map[string]map[string]map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var marketData []types.Info
	for _, coin := range coinCfg.Coins {
		rawData, ok := result["RAW"][strings.ToUpper(coin.ID)]
		if !ok {
			continue
		}
		priceData, ok := rawData[strings.ToUpper(currency)]
		if !ok {
			continue
		}
		if price, ok := priceData["PRICE"]; ok {
			priceFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", price), 64)
			if err != nil {
				continue
			}
			marketData = append(marketData, types.Info{
				ID:     coin.ID,
				Symbol: coin.Symbol,
				Name:   coin.Name,
				Price:  priceFloat,
			})
		}
	}

	c.mu.Lock()
	c.cache = marketData
	c.lastFetched = time.Now()
	c.mu.Unlock()

	return marketData, nil
}
