package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
)

// ActivateVoucherPOST handles voucher activation from landing page
func ActivateVoucherPOST(c echo.Context) error {
	voucherCode := c.FormValue("voucher_code")
	
	if voucherCode == "" {
		return c.String(http.StatusBadRequest, "Kode voucher diperlukan")
	}

	userCtxRaw := c.Get("user_context")
	if userCtxRaw == nil {
        // Option 1: Save voucher code to cookie/session and redirect to login/register, then activate after.
        // For simplicity now, require login first if they use the form from landing while logged in, 
        // or just return an error telling them to login. Let's redirect to login with the code.
		return c.Redirect(http.StatusFound, "/login?next=/dashboard&voucher="+voucherCode)
	}
	
	userCtx := userCtxRaw.(middleware.UserContext)

	// In a real flow, if they are already logged in, they can submit this.
	// But the form is on landing.html. If they are logged in and submit from landing, they have user_context?
	// The landing page doesn't have auth middleware natively on `/` but `/activate-voucher` could require auth,
	// or it handles both. Let's assume it requires auth to apply to a tenant.
	
	tx := db.DB.Begin()
	if tx.Error != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	var voucher models.Voucher
	if err := tx.Where("code = ? AND is_used = ?", voucherCode, false).First(&voucher).Error; err != nil {
		tx.Rollback()
		return c.String(http.StatusBadRequest, "Voucher tidak valid atau sudah digunakan")
	}

	var tenant models.Tenant
	if err := tx.First(&tenant, "id = ?", userCtx.TenantID).Error; err != nil {
		tx.Rollback()
		return c.String(http.StatusInternalServerError, "Tenant tidak ditemukan")
	}

	// Update voucher
	voucher.IsUsed = true
	voucher.UsedBy = &userCtx.TenantID
	now := time.Now()
	/* voucher.UsedAt = &now */
	if err := tx.Save(&voucher).Error; err != nil {
		tx.Rollback()
		return c.String(http.StatusInternalServerError, "Gagal mengupdate voucher")
	}

	// Update tenant plan
	tenant.Plan = "premium"
	
	// Extend expiration if already premium, otherwise set from now
	if tenant.PlanExpiresAt != nil && tenant.PlanExpiresAt.After(now) {
		newExpiry := tenant.PlanExpiresAt.AddDate(0, 0, voucher.Duration)
		tenant.PlanExpiresAt = &newExpiry
	} else {
		newExpiry := now.AddDate(0, 0, voucher.Duration)
		tenant.PlanExpiresAt = &newExpiry
	}

	if err := tx.Save(&tenant).Error; err != nil {
		tx.Rollback()
		return c.String(http.StatusInternalServerError, "Gagal mengupdate tenant")
	}

	tx.Commit()

	return c.Redirect(http.StatusFound, "/dashboard?success=voucher_activated")
}
