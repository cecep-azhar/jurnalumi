package handlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

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
	// Hardcoded dummy data for PRD UI preview
	tenant := models.Tenant{
		Name: "Keluarga Prof. Cecep",
	}

	user := models.User{
		Name: "Cecep Azhar",
		Role: "Owner",
	}

	wallets := []models.Wallet{
		{Name: "BCA Suami", Type: "bank", Balance: 2500000.00},
		{Name: "Uang Tunai Istri", Type: "cash", Balance: 1500000.00},
		{Name: "Gopay Belanja", Type: "ewallet", Balance: 500000.00},
	}

	// Render the Templ view
	return Render(c, views.Dashboard(tenant, user, wallets))
}
