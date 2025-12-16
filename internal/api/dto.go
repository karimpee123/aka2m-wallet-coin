package api

type PriceDTO struct {
	ID            string  `json:"id"`
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Image         string  `json:"image"`
	CurrentPrice  float64 `json:"current_price"`
	TotalVolume   float64 `json:"total_volume"`
	MarketCap     float64 `json:"market_cap"`
	SparklineIn7D struct {
		Price []float64 `json:"price"`
	} `json:"sparkline_in_7d"`
}
