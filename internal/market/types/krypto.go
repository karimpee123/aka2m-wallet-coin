package types

type Sparkline struct {
	Price []float64 `json:"price"`
}

type Info struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	Price       float64   `json:"current_price"`
	MarketCap   float64   `json:"market_cap"`
	TotalVolume float64   `json:"total_volume"`
	Sparkline   Sparkline `json:"sparkline_in_7d"`
}

type InfoCap struct {
	ID                int      `json:"id"`
	Name              string   `json:"name"`
	Symbol            string   `json:"symbol"`
	Slug              string   `json:"slug"`
	NumMarketPairs    int      `json:"num_market_pairs"`
	DateAdded         string   `json:"date_added"`
	Tags              []string `json:"tags"`
	MaxSupply         *float64 `json:"max_supply"`
	CirculatingSupply float64  `json:"circulating_supply"`
	TotalSupply       float64  `json:"total_supply"`
	InfiniteSupply    bool     `json:"infinite_supply"`
	MintedMarketCap   float64  `json:"minted_market_cap"`
	Platform          *struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Symbol       string `json:"symbol"`
		Slug         string `json:"slug"`
		TokenAddress string `json:"token_address"`
	} `json:"platform"`
	CmcRank int `json:"cmc_rank"`
	Quote   struct {
		USD struct {
			Price            float64 `json:"price"`
			Volume24h        float64 `json:"volume_24h"`
			PercentChange1h  float64 `json:"percent_change_1h"`
			PercentChange24h float64 `json:"percent_change_24h"`
			MarketCap        float64 `json:"market_cap"`
		} `json:"USD"`
	} `json:"quote"`
}
