package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/web/views"
)

// AdminDashboardGET renders the Super Admin Dashboard
func AdminDashboardGET(c echo.Context) error {
	var tenants []models.Tenant
	db.DB.Order("created_at desc").Find(&tenants)

	var activeVouchersCount int64
	db.DB.Model(&models.Voucher{}).Where("is_used = ?", false).Count(&activeVouchersCount)

	return Render(c, views.AdminDashboard(tenants, len(tenants), int(activeVouchersCount)))
}

// AdminUpgradeTenantPOST manually approves/upgrades a Tenant plan
func AdminUpgradeTenantPOST(c echo.Context) error {
	tenantIDStr := c.FormValue("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return c.Redirect(http.StatusFound, "/admin/dashboard?error=invalid_id")
	}

	var tenant models.Tenant
	if err := db.DB.First(&tenant, "id = ?", tenantID).Error; err == nil {
		tenant.Plan = "premium"
		now := time.Now().AddDate(0, 1, 0)
		tenant.PlanExpiresAt = &now
		db.DB.Save(&tenant)
	}

	return c.Redirect(http.StatusFound, "/admin/dashboard")
}

// AdminGenerateVoucherPOST generates new activation vouchers
func AdminGenerateVoucherPOST(c echo.Context) error {
	prefix := c.FormValue("prefix")
	durationStr := c.FormValue("duration")

	duration, _ := strconv.Atoi(durationStr)
	if duration <= 0 {
		duration = 30
	}

	// Generate 16 character random code
	randomCode := fmt.Sprintf("%s%s", prefix, uuid.New().String()[:8])

	voucher := models.Voucher{
		Code:     randomCode,
		Duration: duration,
		IsUsed:   false,
	}

	db.DB.Create(&voucher)

	return c.Redirect(http.StatusFound, "/admin/dashboard")
}
