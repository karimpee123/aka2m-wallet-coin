package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Coin struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	ImgURL string `json:"img_url"`
}

type CoinConfig struct {
	Coins []Coin `json:"coins"`
}

func LoadCoinConfig(path string) (*CoinConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read coin config: %w", err)
	}

	var cfg CoinConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse coin config: %w", err)
	}

	return &cfg, nil
}
