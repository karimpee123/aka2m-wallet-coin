package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"coin/pkg/config"
)

type PriceBinance struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type ClientBinance struct {
	Config      config.BinanceConfig
	HTTPClient  *http.Client
	cache       map[string]string
	lastFetched time.Time
	cacheTTL    time.Duration
}

type MarketDataBinance struct {
	ID           string  `json:"id"`
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	CurrentPrice float64 `json:"current_price"`
	MarketCap    float64 `json:"market_cap"`
	TotalVolume  float64 `json:"total_volume"`
	Image        string  `json:"image"`
	Sparkline    struct {
		Price []float64 `json:"price"`
	} `json:"sparkline_in_7d"`
}

func NewClient(cfg config.BinanceConfig) *ClientBinance {
	return &ClientBinance{
		Config:      cfg,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
		cache:       make(map[string]string),
		lastFetched: time.Time{},
		cacheTTL:    1 * time.Minute,
	}
}

func (c *ClientBinance) fetchAllPrices() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := c.Config.BaseURL + c.Config.MarketURL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var prices []PriceBinance
	if err := json.NewDecoder(resp.Body).Decode(&prices); err != nil {
		return err
	}
	c.cache = make(map[string]string)
	for _, p := range prices {
		c.cache[p.Symbol] = p.Price
	}
	return nil
}

func (c *ClientBinance) GetPrice(symbol string) (string, error) {
	if len(c.cache) == 0 {
		if err := c.fetchAllPrices(); err != nil {
			return "", err
		}
	}
	price, ok := c.cache[symbol]
	if !ok {
		return "", fmt.Errorf("symbol %s not found", symbol)
	}
	return price, nil
}

func (c *ClientBinance) GetAllPrices() (map[string]string, error) {
	if len(c.cache) == 0 {
		if err := c.fetchAllPrices(); err != nil {
			return nil, err
		}
	}
	return c.cache, nil
}
