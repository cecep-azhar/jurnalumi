import re

with open("internal/handlers/assets.go", "r") as f:
    content = f.read()

# Fix AssetGET to read from PriceSnapshot instead of CalculateCommodityValue for each row
old_asset_get = """	// Phase 4: API Provider Integration (Live Gold Pricing)
	var totalAssetValue int64 = 0.0
	for i, a := range assets {
		assets[i].CurrentValue = services.CalculateCommodityValue(a.Type, a.WeightGram, a.Karatage)
		totalAssetValue += assets[i].CurrentValue
	}"""

new_asset_get = """	// QA-P1-14: Baca dari snapshot terbaru
	var latestSnapshot models.PriceSnapshot
	if err := db.DB.Order("snapshot_date desc").First(&latestSnapshot).Error; err != nil {
		// Fallback if no snapshot found yet
		latestSnapshot.PricePerGram = services.FetchLiveGoldPrice()
	}

	var totalAssetValue int64 = 0
	for i, a := range assets {
		assets[i].CurrentValue = services.CalculateCommodityValueWithPrice(a.Type, a.WeightGram, a.Karatage, latestSnapshot.PricePerGram)
		totalAssetValue += assets[i].CurrentValue
	}"""
content = content.replace(old_asset_get, new_asset_get)

# Fix AssetPOST to read from snapshot for initial value
old_asset_post = """	// In real world, we fetch current value from API
	currentValue := services.CalculateCommodityValue(assetType, weight, karatage)"""

new_asset_post = """	// Read latest snapshot
	var latestSnapshot models.PriceSnapshot
	if err := db.DB.Order("snapshot_date desc").First(&latestSnapshot).Error; err != nil {
		latestSnapshot.PricePerGram = services.FetchLiveGoldPrice()
	}
	currentValue := services.CalculateCommodityValueWithPrice(assetType, weight, karatage, latestSnapshot.PricePerGram)"""
content = content.replace(old_asset_post, new_asset_post)

with open("internal/handlers/assets.go", "w") as f:
    f.write(content)
