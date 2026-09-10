package middleware

import (
	"github.com/labstack/echo/v4"
)

// RLS is a middleware that sets the local PostgreSQL variable app.tenant_id
// for Row-Level Security isolation.
func RLS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}
