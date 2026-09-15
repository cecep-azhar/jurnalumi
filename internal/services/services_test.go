package services_test

import (
	"math"
	"testing"

	"github.com/cecep-azhar/jurnalumi/internal/services"
)

func TestPricingConstants(t *testing.T) {
	price := services.GetMonthlyPrice()
	if price <= 0 {
		t.Fatalf("expected monthly price > 0, got %d", price)
	}
	if services.FreeTrialMonths <= 0 {
		t.Fatalf("expected free trial months > 0, got %d", services.FreeTrialMonths)
	}
}

func TestCommodityCalculations(t *testing.T) {
	// Zero weight check
	valZero := services.CalculateCommodityValue("gold", 0, 24)
	if valZero != 0 {
		t.Errorf("expected 0 for 0 weight, got %d", valZero)
	}

	// 1 gram 24k gold check
	valGold24 := services.CalculateCommodityValue("gold", 1, 24)
	if valGold24 != int64(services.FallbackGoldPricePerGram) {
		t.Errorf("expected %d, got %d", int64(services.FallbackGoldPricePerGram), valGold24)
	}

	// 1 keping dinar
	valDinar := services.CalculateCommodityValue("dinar", 1, 22)
	goldPrice := float64(services.FallbackGoldPricePerGram)
	expectedDinar := int64(goldPrice * 4.25 * (22.0 / 24.0))
	if valDinar != expectedDinar {
		t.Errorf("expected dinar value %d, got %d", expectedDinar, valDinar)
	}

	// 100 gram silver
	valSilver := services.CalculateCommodityValue("silver", 100, 0)
	expectedSilver := int64(100 * services.FixedSilverPricePerGram)
	if valSilver != expectedSilver {
		t.Errorf("expected silver value %d, got %d", expectedSilver, valSilver)
	}

	// Large calculation
	largeWeight := int64(1_000_000_000)
	valLarge := services.CalculateCommodityValue("gold", largeWeight, 24)
	if valLarge < 0 || valLarge > math.MaxInt64 {
		t.Errorf("potential int64 overflow detected in large commodity calculation: %d", valLarge)
	}
}
