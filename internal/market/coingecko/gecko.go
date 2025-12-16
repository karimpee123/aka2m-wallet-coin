package coingecko

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"coin/internal/market/types"
	"coin/pkg/config"
)

type ClientGecko struct {
	BaseURL     string
	MarketURL   string
	cache       []types.Info
	lastFetched time.Time
	cacheTTL    time.Duration
	mu          sync.RWMutex
}

func NewClient(cfg *config.CoinGeckoConfig) *ClientGecko {
	return &ClientGecko{
		BaseURL:   cfg.BaseURL,
		MarketURL: cfg.MarketURL,
		cache:     []types.Info{},
		cacheTTL:  1 * time.Minute,
	}
}

func (c *ClientGecko) GetMarketData(coinCfg *config.CoinConfig, currency string) ([]types.Info, error) {
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

	url := fmt.Sprintf("%s%s?vs_currency=%s&ids=%s&order=market_cap_desc&sparkline=true",
		c.BaseURL, c.MarketURL, currency, ids)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinGecko API returned status %d", resp.StatusCode)
	}

	var result []types.Info
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache = result
	c.lastFetched = time.Now()
	c.mu.Unlock()

	return result, nil
}
