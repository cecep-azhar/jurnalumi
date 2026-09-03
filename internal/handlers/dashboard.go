package handlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/web/views"
)

// Render is a custom Echo wrapper to render Templ components
func Render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// DashboardHandler handles the rendering of the authenticated dashboard
func DashboardHandler(c echo.Context) error {
	// Extract secure User Context from Middleware Session
	userCtx := c.Get("user_context").(middleware.UserContext)

	// Fetch Real Tenant Data from DB
	var tenant models.Tenant
	db.DB.First(&tenant, "id = ?", userCtx.TenantID)

	// User Object for UI
	user := models.User{
		Name: userCtx.Name,
		Role: userCtx.Role,
	}

	// Fetch Wallets belongs to this Tenant ONLY (Multi-Tenancy Isolation)
	var wallets []models.Wallet
	db.DB.Where("tenant_id = ?", userCtx.TenantID).Find(&wallets)

	// Render the Templ view with Real Data
	return Render(c, views.Dashboard(tenant, user, wallets))
}
