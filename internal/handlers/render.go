package handlers

import (
	"context"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// Render is a custom Echo wrapper to render Templ components globally for handlers
func Render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	ctx := c.Request().Context()
	if token, ok := c.Get("csrf").(string); ok {
		ctx = context.WithValue(ctx, "csrf", token)
	} else {
		ctx = context.WithValue(ctx, "csrf", "")
	}
	return component.Render(ctx, c.Response().Writer)
}
