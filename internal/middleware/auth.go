package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type UserContext struct {
	UserID   uuid.UUID
	TenantID uuid.UUID
	Role     string
	Email    string
	Name     string
}

// RequireAuth ensures that the user is logged in
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("jurnalumi_session", c)
		if err != nil || sess.Values["user_id"] == nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		userIDStr, ok := sess.Values["user_id"].(string)
		tenantIDStr, _ := sess.Values["tenant_id"].(string)
		
		if !ok || userIDStr == "" {
			return c.Redirect(http.StatusFound, "/login")
		}

		userID, _ := uuid.Parse(userIDStr)
		tenantID, _ := uuid.Parse(tenantIDStr)
		role, _ := sess.Values["role"].(string)
		email, _ := sess.Values["email"].(string)
		name, _ := sess.Values["name"].(string)

		userCtx := UserContext{
			UserID:   userID,
			TenantID: tenantID,
			Role:     role,
			Email:    email,
			Name:     name,
		}

		// Inject to Context
		c.Set("user_context", userCtx)

		return next(c)
	}
}
