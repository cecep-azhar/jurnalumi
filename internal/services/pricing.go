package services

import "os"

// Pricing configuration (single source of truth).
// Set via environment variable or default fallback.
const (
	DefaultMonthlyPriceIDR = 9000
	FreeTrialMonths        = 6
	ManualPaymentBank      = "BSI (Bank Syariah Indonesia)"
	ManualPaymentAccount   = "7043984831"
	ManualPaymentName      = "CECEP SAEFUL AZHAR HIDAYAT"
)

// GetMonthlyPrice returns monthly subscription price in IDR.
func GetMonthlyPrice() int64 {
	if p := os.Getenv("PREMIUM_PRICE_MONTHLY"); p != "" {
		// ponytail: add env parsing when dynamic price needed
	}
	return DefaultMonthlyPriceIDR
}
