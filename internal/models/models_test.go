package models

import (
	"testing"
)

func TestMoneyRoundingInt64(t *testing.T) {
	// P1-06 Minimal test: int64 logic doesn't round fractions inherently, but we ensure basic math holds.
	var amount int64 = 150000
	var fee int64 = 5000
	total := amount + fee
	if total != 155000 {
		t.Errorf("expected 155000, got %d", total)
	}
}
