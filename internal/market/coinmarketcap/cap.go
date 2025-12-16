package coinmarketcap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"coin/internal/market/types"
	"coin/pkg/config"
)

type ClientCoinMarketCap struct {
	BaseURL     string
	MarketURL   string
	apiKey      string
	cache       []types.InfoCap
	lastFetched time.Time
	cacheTTL    time.Duration
	mu          sync.RWMutex
}

func NewClient(cfg *config.CoinMarketCapConfig) *ClientCoinMarketCap {
	return &ClientCoinMarketCap{
		BaseURL:   cfg.BaseURL,
		MarketURL: cfg.MarketURL,
		apiKey:    cfg.ApiKey,
		cache:     []types.InfoCap{},
		cacheTTL:  1 * time.Minute,
	}
}

func (c *ClientCoinMarketCap) GetMarketData(coinCfg *config.CoinConfig, currency string) ([]types.InfoCap, error) {
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
	url := fmt.Sprintf("%s%s?convert=%s&limit=100", c.BaseURL, c.MarketURL, currency)
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-CMC_PRO_API_KEY", c.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinMarketCap API returned status %d", resp.StatusCode)
	}
	var result struct {
		Data []types.InfoCap `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache = result.Data
	c.lastFetched = time.Now()
	c.mu.Unlock()

	return result.Data, nil
}
