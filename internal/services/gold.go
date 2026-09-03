package services

import (
	"encoding/json"
	"net/http"
	"time"
)

type GoldPriceResponse struct {
	PricePerGram float64 `json:"price_per_gram"`
	Currency     string  `json:"currency"`
}

// FetchLiveGoldPrice fetches real-time gold price (Mocked / Fallback to Antam standard)
func FetchLiveGoldPrice() float64 {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.logammulia.com/v1/price")
	if err == nil && resp.StatusCode == http.StatusOK {
		var res GoldPriceResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err == nil && res.PricePerGram > 0 {
			return res.PricePerGram
		}
	}

	// Standard Fallback Price (Antam 24K September 2026 ~ Rp 1.450.000 / gram)
	return 1450000.00
}

// CalculateCommodityValue calculates real-time IDR value of gold/dinar
func CalculateCommodityValue(commodityType string, weightGram float64, karatage float64) float64 {
	liveGoldPrice := FetchLiveGoldPrice()
	karatRatio := karatage / 24.0

	if commodityType == "dinar" {
		// 1 Dinar = 4.25 Gram Emas 22K (91.6%)
		return weightGram * liveGoldPrice * (22.0 / 24.0)
	} else if commodityType == "silver" {
		// Perak ~ Rp 16.500 / gram
		return weightGram * 16500.00
	}

	// Gold Bar (Default)
	return weightGram * liveGoldPrice * karatRatio
}
