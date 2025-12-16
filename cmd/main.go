package main

import (
	"coin/internal/api"
	"coin/internal/market"
	"coin/pkg/config"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	clients, err := config.LoadClientsFromJSON("config/clients.json")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/price", PriceHandler(clients))
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func PriceHandler(clients map[string]*market.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		provider := r.URL.Query().Get("provider")
		if provider == "" {
			provider = "gecko"
		}
		client, ok := clients[provider]
		if !ok {
			http.Error(w, "provider not found", http.StatusBadRequest)
			return
		}
		data, err := client.GetAllCryptoData(nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result := make([]api.PriceDTO, 0, len(data))
		for symbol, md := range data {
			dto := api.PriceDTO{
				ID:           strings.ToLower(symbol),
				Symbol:       strings.ToLower(symbol),
				Name:         symbol,
				Image:        "/media/images/coins/" + strings.ToLower(symbol) + ".png",
				CurrentPrice: md.Price,
				TotalVolume:  md.Volume24h,
				MarketCap:    md.MarketCap,
			}
			dto.SparklineIn7D.Price = md.Sparkline
			result = append(result, dto)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
