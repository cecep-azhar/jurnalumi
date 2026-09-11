with open("internal/services/gold.go", "r") as f:
    content = f.read()

new_func = """// CalculateCommodityValueWithPrice calculates real-time IDR value based on a given base price
func CalculateCommodityValueWithPrice(commodityType string, weightGram int64, karatage int64, baseGoldPrice int64) int64 {
	karatRatio := float64(karatage) / 24.0

	if commodityType == "dinar" {
		// 1 Dinar = 4.25 Gram Emas 22K (91.6%)
		return int64(float64(weightGram) * float64(baseGoldPrice) * (22.0 / 24.0))
	} else if commodityType == "silver" || commodityType == "perak" {
		// Perak ~ Rp 16.500 / gram
		return weightGram * 16500
	}

	// Gold Bar (Default)
	return int64(float64(weightGram) * float64(baseGoldPrice) * karatRatio)
}
"""

if "CalculateCommodityValueWithPrice" not in content:
    content += "\n" + new_func

with open("internal/services/gold.go", "w") as f:
    f.write(content)
