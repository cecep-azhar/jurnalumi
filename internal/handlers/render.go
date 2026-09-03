package handlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// Render is a custom Echo wrapper to render Templ components globally for handlers
func Render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return component.Render(c.Request().Context(), c.Response().Writer)
}
