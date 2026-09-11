package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/web/views"
)

// ActivateVoucherGET renders the voucher activation page
func ActivateVoucherGET(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	var tenant models.Tenant
	db.DB.First(&tenant, "id = ?", userCtx.TenantID)

	user := models.User{
		Name: userCtx.Name,
		Role: userCtx.Role,
	}

	return Render(c, views.ActivateVoucherGET(tenant, user, "", false))
}

// ActivateVoucherPOST validates and redeems a voucher code
func ActivateVoucherPOST(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	var tenant models.Tenant
	if err := db.DB.First(&tenant, "id = ?", userCtx.TenantID).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Tenant tidak ditemukan")
	}

	user := models.User{
		Name: userCtx.Name,
		Role: userCtx.Role,
	}

	code := strings.TrimSpace(strings.ToUpper(c.FormValue("voucher_code")))
	if code == "" {
		return Render(c, views.ActivateVoucherGET(tenant, user, "Kode voucher tidak boleh kosong.", true))
	}

	// Already premium?
	if tenant.Plan == "premium" && tenant.PlanExpiresAt != nil && tenant.PlanExpiresAt.After(time.Now()) {
		return Render(c, views.ActivateVoucherGET(tenant, user, "Akun Anda sudah premium.", true))
	}

	var voucher models.Voucher
	if err := db.DB.Where("code = ?", code).First(&voucher).Error; err != nil {
		return Render(c, views.ActivateVoucherGET(tenant, user, "Kode voucher tidak ditemukan.", true))
	}

	if voucher.IsUsed {
		return Render(c, views.ActivateVoucherGET(tenant, user, "Voucher ini sudah digunakan.", true))
	}

	// Redeem: upgrade tenant + mark voucher used, in one transaction
	tx := db.DB.Begin()
	if tx.Error != nil {
		return Render(c, views.ActivateVoucherGET(tenant, user, "Gagal memproses, coba lagi.", true))
	}

	now := time.Now()
	expiry := now.AddDate(0, 0, voucher.Duration)
	tenant.Plan = "premium"
	tenant.PlanExpiresAt = &expiry

	if err := tx.Save(&tenant).Error; err != nil {
		tx.Rollback()
		return Render(c, views.ActivateVoucherGET(tenant, user, "Gagal mengupgrade paket.", true))
	}

	voucher.IsUsed = true
	voucher.UsedBy = &userCtx.TenantID
	voucher.UsedAt = &now
	if err := tx.Save(&voucher).Error; err != nil {
		tx.Rollback()
		return Render(c, views.ActivateVoucherGET(tenant, user, "Gagal menyimpan data voucher.", true))
	}

	if err := tx.Commit().Error; err != nil {
		return Render(c, views.ActivateVoucherGET(tenant, user, "Gagal memproses, coba lagi.", true))
	}

	return Render(c, views.ActivateVoucherGET(tenant, user, "Voucher berhasil diaktifkan! Akun Anda sekarang Premium.", false))
}
