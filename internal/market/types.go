package market

type MarketData struct {
	Symbol    string
	Price     float64
	Volume24h float64
	MarketCap float64
	Sparkline []float64
}
