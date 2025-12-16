package market

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ParserFunc func(baseURL, marketURL string, params map[string]string, client *http.Client) (map[string]CryptoData, error)

var (
	BinanceParser       = parseBinance
	GeckoParser         = parseCoinGecko
	CryptoCompareParser = parseCryptoCompare
	CoinMarketCapParser = parseCoinMarketCap
)

type ResponseParser func([]byte) (map[string]float64, error)

type Client struct {
	BaseURL     string
	MarketURL   string
	Params      map[string]string
	Parser      ParserFunc
	HTTPClient  *http.Client
	CacheTTL    time.Duration
	LastFetched time.Time
	Cache       map[string]CryptoData
}

func NewClient(baseURL, marketURL string, params map[string]string, parser ParserFunc, timeout, cacheTTL time.Duration) *Client {
	return &Client{
		BaseURL:    baseURL,
		MarketURL:  marketURL,
		Params:     params,
		Parser:     parser,
		HTTPClient: &http.Client{Timeout: timeout},
		CacheTTL:   cacheTTL,
		Cache:      make(map[string]CryptoData),
	}
}

func (c *Client) ShouldFetch() bool {
	return time.Since(c.LastFetched) > c.CacheTTL
}

func (c *Client) GetAllCryptoData(extraParams map[string]string) (map[string]CryptoData, error) {
	if time.Since(c.LastFetched) < c.CacheTTL && len(c.Cache) > 0 {
		return c.Cache, nil
	}
	params := map[string]string{}
	for k, v := range c.Params {
		params[k] = v
	}
	for k, v := range extraParams {
		params[k] = v
	}
	data, err := c.Parser(c.BaseURL, c.MarketURL, params, c.HTTPClient)
	if err != nil {
		return nil, err
	}
	c.Cache = data
	c.LastFetched = time.Now()
	return data, nil
}

func (c *Client) GetPrice(symbol string) (float64, error) {
	data, err := c.GetAllCryptoData(nil)
	if err != nil {
		return 0, err
	}
	md, ok := data[symbol]
	if !ok {
		return 0, fmt.Errorf("symbol %s not found", symbol)
	}
	return md.Price, nil
}

func parseBinance(baseURL, marketURL string, params map[string]string, client *http.Client) (map[string]CryptoData, error) {
	resp, err := client.Get(baseURL + marketURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var raw []struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	result := make(map[string]CryptoData)
	for _, r := range raw {
		price, _ := strconv.ParseFloat(r.Price, 64)
		result[r.Symbol] = CryptoData{
			Symbol: r.Symbol,
			Price:  price,
		}
	}
	return result, nil
}

func parseCoinGecko(baseURL, marketURL string, params map[string]string, client *http.Client) (map[string]CryptoData, error) {
	url := baseURL + marketURL + "?"
	for k, v := range params {
		url += fmt.Sprintf("%s=%s&", k, v)
	}
	resp, err := client.Get(strings.TrimSuffix(url, "&"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var raw []struct {
		Symbol       string  `json:"symbol"`
		CurrentPrice float64 `json:"current_price"`
		MarketCap    float64 `json:"market_cap"`
		Volume24h    float64 `json:"total_volume"`
		Sparkline    struct {
			Price []float64 `json:"price"`
		} `json:"sparkline_in_7d"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	result := make(map[string]CryptoData)
	for _, r := range raw {
		sym := strings.ToUpper(r.Symbol)
		result[sym] = CryptoData{
			Symbol:    sym,
			Price:     r.CurrentPrice,
			MarketCap: r.MarketCap,
			Volume24h: r.Volume24h,
			Sparkline: r.Sparkline.Price,
		}
	}
	return result, nil
}

func parseCryptoCompare(baseURL, marketURL string, params map[string]string, client *http.Client) (map[string]CryptoData, error) {
	qs := "?"
	for k, v := range params {
		qs += fmt.Sprintf("%s=%s&", k, v)
	}
	url := baseURL + marketURL + strings.TrimSuffix(qs, "&")
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	result := make(map[string]CryptoData)
	rawRAW, ok := raw["RAW"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("RAW field missing")
	}
	for fromSym, toMapRaw := range rawRAW {
		toMap, ok := toMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		for _, dataRaw := range toMap {
			data, ok := dataRaw.(map[string]interface{})
			if !ok {
				continue
			}
			md := CryptoData{
				Symbol: strings.ToUpper(fromSym),
			}
			if v, ok := data["PRICE"].(float64); ok {
				md.Price = v
			}
			if v, ok := data["VOLUME24H"].(float64); ok {
				md.Volume24h = v
			}
			if v, ok := data["MKTCAP"].(float64); ok {
				md.MarketCap = v
			}
			result[md.Symbol] = md
		}
	}
	return result, nil
}
func parseCoinMarketCap(baseURL, marketURL string, params map[string]string, client *http.Client) (map[string]CryptoData, error) {
	qs := "?"
	for k, v := range params {
		qs += fmt.Sprintf("%s=%s&", k, v)
	}
	url := baseURL + marketURL + strings.TrimSuffix(qs, "&")
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-CMC_PRO_API_KEY", params["api_key"])
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}
	var raw struct {
		Data []struct {
			Symbol string `json:"symbol"`
			Quote  map[string]struct {
				Price     float64 `json:"price"`
				Volume24h float64 `json:"volume_24h"`
				MarketCap float64 `json:"market_cap"`
			} `json:"quote"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}
	if len(raw.Data) == 0 {
		return nil, fmt.Errorf("no data found in API response")
	}
	result := make(map[string]CryptoData)
	for _, c := range raw.Data {
		for _, q := range c.Quote {
			result[c.Symbol] = CryptoData{
				Symbol:    c.Symbol,
				Price:     q.Price,
				Volume24h: q.Volume24h,
				MarketCap: q.MarketCap,
			}
		}
	}
	return result, nil
}
