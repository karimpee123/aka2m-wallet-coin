package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type BinanceConfig struct {
	BaseURL   string `json:"baseUrl"`
	MarketURL string `json:"marketUrl"`
}

type CoinGeckoConfig struct {
	BaseURL   string `json:"BaseURL"`
	MarketURL string `json:"marketUrl"`
}

type CryptoCompareConfig struct {
	BaseURL   string `json:"BaseURL"`
	MarketURL string `json:"marketUrl"`
	ApiKey    string `json:"apiKey"`
}

type CoinMarketCapConfig struct {
	BaseURL   string `json:"BaseURL"`
	MarketURL string `json:"marketUrl"`
	ApiKey    string `json:"apiKey"`
}

type APIConfig struct {
	Binance       BinanceConfig       `json:"binance"`
	CoinGecko     CoinGeckoConfig     `json:"CoinGecko"`
	CryptoCompare CryptoCompareConfig `json:"CryptoCompare"`
	CoinMarketCap CoinMarketCapConfig `json:"CoinMarketCap"`
}

func LoadApiConfig(path string) (*APIConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var cfg APIConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}
