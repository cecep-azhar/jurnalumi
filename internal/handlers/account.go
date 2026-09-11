package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/web/views"
)

// AccountGET renders the account settings and data management page
func AccountGET(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	var tenant models.Tenant
	if err := db.DB.First(&tenant, "id = ?", userCtx.TenantID).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Data tenant tidak ditemukan")
	}

	user := models.User{
		Name: userCtx.Name,
		Role: userCtx.Role,
	}

	return Render(c, views.AccountSettings(tenant, user))
}

// AccountExportGET generates a JSON export of all tenant data for GDPR/PDP compliance
func AccountExportGET(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	if userCtx.Role != "owner" {
		return c.String(http.StatusForbidden, "Hanya owner yang dapat mengekspor data")
	}

	// Gather all tenant data
	var tenant models.Tenant
	db.DB.First(&tenant, "id = ?", userCtx.TenantID)

	var users []models.User
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Find(&users)

	var wallets []models.Wallet
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Find(&wallets)

	var categories []models.Category
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Find(&categories)

	var transactions []models.Transaction
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Find(&transactions)

	var assets []models.CommodityAsset
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Find(&assets)

	var debts []models.Debt
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Find(&debts)

	exportData := map[string]interface{}{
		"export_date":  time.Now(),
		"tenant":       tenant,
		"users":        users,
		"wallets":      wallets,
		"categories":   categories,
		"transactions": transactions,
		"assets":       assets,
		"debts":        debts,
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename=\"jurnalumi_export.json\"")
	c.Response().Header().Set("Content-Type", "application/json")
	return json.NewEncoder(c.Response().Writer).Encode(exportData)
}

// AccountDeletePOST permanently deletes the tenant account and all related data (soft delete)
func AccountDeletePOST(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	if userCtx.Role != "owner" {
		return c.String(http.StatusForbidden, "Hanya owner yang dapat menghapus akun")
	}

	// Soft delete the tenant (cascade will not happen automatically unless configured, 
	// but standard soft delete of tenant will lock users out since login checks tenant).
	// To be safer according to PDP, we will soft-delete all child records.
	
	tx := db.DB.Begin()
	if tx.Error != nil {
		return c.String(http.StatusInternalServerError, "Gagal memulai transaksi")
	}

	tenantID := userCtx.TenantID

	if err := tx.Where("tenant_id = ?", tenantID).Delete(&models.Transaction{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("tenant_id = ?", tenantID).Delete(&models.CommodityAsset{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("tenant_id = ?", tenantID).Delete(&models.Debt{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("tenant_id = ?", tenantID).Delete(&models.Category{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("tenant_id = ?", tenantID).Delete(&models.Wallet{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("tenant_id = ?", tenantID).Delete(&models.User{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("id = ?", tenantID).Delete(&models.Tenant{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return c.String(http.StatusInternalServerError, "Gagal menghapus data")
	}

	return c.Redirect(http.StatusFound, "/logout")
}
