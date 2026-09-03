package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/cecep-azhar/jurnalumi/web/views"
)

// LoginGET renders the login page
func LoginGET(c echo.Context) error {
	return Render(c, views.Login())
}

// LoginPOST handles the login form submission
func LoginPOST(c echo.Context) error {
	// Dummy bypass redirect ke dashboard
	return c.Redirect(http.StatusFound, "/dashboard")
}

// RegisterGET renders the registration page
func RegisterGET(c echo.Context) error {
	return Render(c, views.Register())
}

// RegisterPOST handles the registration form submission
func RegisterPOST(c echo.Context) error {
	// Dummy bypass redirect ke dashboard
	return c.Redirect(http.StatusFound, "/dashboard")
}
