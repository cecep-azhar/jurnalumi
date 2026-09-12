package services

// Fixed prices for soft-launch (manual update required).
// Update these periodically.
const (
	FallbackGoldPricePerGram = 1450000
	FixedSilverPricePerGram  = 16500
)

// GetGoldPricePerGram returns the current gold price for cron snapshots
// Note: Currently uses manual periodic update, not real-time API.
func GetGoldPricePerGram() (int64, error) {
	return FallbackGoldPricePerGram, nil
}

// GetSilverPricePerGramFallback returns the fixed silver price
func GetSilverPricePerGramFallback() int64 {
	return FixedSilverPricePerGram
}

// CalculateCommodityValue calculates the IDR value of gold/dinar based on manual pricing
func CalculateCommodityValue(commodityType string, weightGram int64, karatage int64) int64 {
	liveGoldPrice := int64(FallbackGoldPricePerGram)
	karatRatio := float64(karatage) / 24.0

	if commodityType == "dinar" {
		// weightGram represents "keping" for dinar
		// 1 Dinar = 4.25 Gram Emas 22K (91.6%)
		return int64(float64(weightGram) * 4.25 * float64(liveGoldPrice) * (22.0 / 24.0))
	} else if commodityType == "silver" {
		// Perak (harga manual berkala)
		return weightGram * FixedSilverPricePerGram
	}

	// Gold Bar (Default)
	return int64(float64(weightGram) * float64(liveGoldPrice) * karatRatio)
}
