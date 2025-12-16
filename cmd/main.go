package main

import (
	"coin/internal/market"
	"encoding/json"
	"log"
	"net/http"
)

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	http.HandleFunc("/price", priceHandler)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func priceHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	provider := r.URL.Query().Get("provider")
	currency := r.URL.Query().Get("currency")
	if provider == "" {
		provider = "gecko"
	}
	if currency == "" {
		currency = "usd"
	}
	prices, err := market.CryptoInstant.Fetch(provider, currency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prices)
}
