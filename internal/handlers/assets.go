package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/internal/services"
	"github.com/cecep-azhar/jurnalumi/web/views"
)

// AssetGET renders the Asset Management page
func AssetGET(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	var tenant models.Tenant
	db.DB.First(&tenant, "id = ?", userCtx.TenantID)

	user := models.User{
		Name: userCtx.Name,
		Role: userCtx.Role,
	}

	var assets []models.CommodityAsset
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Find(&assets)

	// QA-P1-14: Read prices from snapshots instead of live HTTP per asset
	goldPrice := getSnapshotPrice("antam", services.FallbackGoldPricePerGram)
	silverPrice := getSnapshotPrice("perak", services.FixedSilverPricePerGram)
	dinarPrice := getSnapshotPrice("dinar", int64(float64(goldPrice)*4.25))

	var totalAssetValue int64 = 0
	for i, a := range assets {
		assets[i].CurrentValue = calculateFromSnapshot(a.Type, a.WeightGram, a.Karatage, goldPrice, silverPrice, dinarPrice)
		totalAssetValue += assets[i].CurrentValue
	}

	return Render(c, views.AssetManagement(tenant, user, assets, totalAssetValue))
}

// AssetPOST handles adding a new gold/dinar asset
func AssetPOST(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	name := c.FormValue("name")
	assetType := c.FormValue("type")
	weightStr := c.FormValue("weight")
	karatageStr := c.FormValue("karatage")
	buyPriceStr := c.FormValue("buy_price")

	weight, _ := strconv.ParseInt(weightStr, 10, 64)
	karatage, _ := strconv.ParseInt(karatageStr, 10, 64)
	buyPrice, _ := strconv.ParseInt(buyPriceStr, 10, 64)

	currentValue := services.CalculateCommodityValue(assetType, weight, karatage)

	asset := models.CommodityAsset{
		TenantID:     userCtx.TenantID,
		Type:         assetType,
		Name:         name,
		WeightGram:   weight, // for dinar this acts as pieces/keping
		Karatage:     karatage,
		BuyPrice:     buyPrice,
		CurrentValue: currentValue,
	}

	if name == "" || weight <= 0 || karatage < 0 || buyPrice < 0 || (assetType != "gold_bar" && assetType != "silver" && assetType != "dinar") {
		return c.String(http.StatusBadRequest, "Input tidak valid")
	}

	if err := db.DB.Create(&asset).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Gagal menambahkan aset")
	}

	return c.Redirect(http.StatusFound, "/assets")
}

func getSnapshotPrice(assetType string, fallback int64) int64 {
	var snap models.PriceSnapshot
	if err := db.DB.Order("snapshot_date DESC").First(&snap, "commodity_type = ?", assetType).Error; err == nil && snap.PricePerGram > 0 {
		return snap.PricePerGram
	}
	return fallback
}

func calculateFromSnapshot(assetType string, weightGram int64, karatage int64, goldPrice int64, silverPrice int64, dinarPrice int64) int64 {
	karatRatio := float64(karatage) / 24.0

	if assetType == "dinar" {
		// weightGram represents "keping" for dinar, 1 Dinar = 4.25 Gram Emas 22K (91.6%)
		return int64(float64(weightGram) * float64(dinarPrice) * (22.0 / 24.0))
	} else if assetType == "silver" {
		return weightGram * silverPrice
	}

	return int64(float64(weightGram) * float64(goldPrice) * karatRatio)
}
