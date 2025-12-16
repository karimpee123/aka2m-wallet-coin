package config

import (
	"encoding/json"
	"os"
	"time"

	"coin/internal/market"
)

type ClientJSONConfig struct {
	Name         string            `json:"name"`
	BaseURL      string            `json:"base_url"`
	MarketURL    string            `json:"market_url"`
	Params       map[string]string `json:"params"`
	TimeoutSecs  int               `json:"timeout_seconds"`
	CacheTTLSecs int               `json:"cache_ttl_seconds"`
}

type ClientsConfig struct {
	Clients []ClientJSONConfig `json:"clients"`
}

func LoadClientsFromJSON(path string) (map[string]*market.Client, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var cfg ClientsConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	clients := make(map[string]*market.Client)
	for _, c := range cfg.Clients {
		var parser market.ParserFunc
		switch c.Name {
		case "binance":
			parser = market.BinanceParser
		case "gecko":
			parser = market.GeckoParser
		case "cryptocompare":
			parser = market.CryptoCompareParser
		case "coinmarketcap":
			parser = market.CoinMarketCapParser
		}
		client := market.NewClient(
			c.BaseURL,
			c.MarketURL,
			c.Params,
			parser,
			time.Duration(c.TimeoutSecs)*time.Second,
			time.Duration(c.CacheTTLSecs)*time.Second,
		)
		clients[c.Name] = client
	}
	return clients, nil
}
