package middleware

import (
	"net/http"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/labstack/echo/v4"
)

// CheckPlan injects tenant plan status to context
func CheckPlan(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userCtxRaw := c.Get("user_context")
		if userCtxRaw == nil {
			return next(c)
		}
		userCtx, ok := userCtxRaw.(UserContext)
		if !ok {
			return next(c)
		}

		var tenant models.Tenant
		if err := db.DB.First(&tenant, "id = ?", userCtx.TenantID).Error; err == nil {
			c.Set("is_premium", tenant.Plan == "premium")
			c.Set("tenant_plan", tenant.Plan)
		} else {
			c.Set("is_premium", false)
			c.Set("tenant_plan", "free")
		}

		return next(c)
	}
}

// RequirePremiumFeature blocks access if user is not premium
func RequirePremiumFeature(featureName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			isPremiumRaw := c.Get("is_premium")
			isPremium := false
			if isPremiumRaw != nil {
				isPremium = isPremiumRaw.(bool)
			}
			
			if !isPremium {
				return c.String(http.StatusForbidden, "Fitur "+featureName+" hanya tersedia di paket Premium")
			}
			return next(c)
		}
	}
}
