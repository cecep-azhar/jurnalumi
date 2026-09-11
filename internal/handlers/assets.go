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

	// QA-P1-14: Baca dari snapshot terbaru
	var latestSnapshot models.PriceSnapshot
	if err := db.DB.Order("snapshot_date desc").First(&latestSnapshot).Error; err != nil {
		// Fallback if no snapshot found yet
		latestSnapshot.PricePerGram = services.FetchLiveGoldPrice()
	}

	var totalAssetValue int64 = 0
	for i, a := range assets {
		assets[i].CurrentValue = services.CalculateCommodityValueWithPrice(a.Type, a.WeightGram, a.Karatage, latestSnapshot.PricePerGram)
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

	// Read latest snapshot
	var latestSnapshot models.PriceSnapshot
	if err := db.DB.Order("snapshot_date desc").First(&latestSnapshot).Error; err != nil {
		latestSnapshot.PricePerGram = services.FetchLiveGoldPrice()
	}
	currentValue := services.CalculateCommodityValueWithPrice(assetType, weight, karatage, latestSnapshot.PricePerGram)

	asset := models.CommodityAsset{
		TenantID:     userCtx.TenantID,
		Type:         assetType,
		Name:         name,
		WeightGram:   weight,
		Karatage:     karatage,
		BuyPrice:     buyPrice,
		CurrentValue: currentValue,
	}

	if name == "" || weight <= 0 || karatage < 0 || buyPrice < 0 || (assetType != "emas" && assetType != "perak" && assetType != "dinar") {
		return c.String(http.StatusBadRequest, "Input tidak valid")
	}

	if err := db.DB.Create(&asset).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Gagal menambahkan aset")
	}

	return c.Redirect(http.StatusFound, "/assets")
}
