package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/labstack/echo/v4"
)

func TestRequireRole(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	c.Set("user_context", UserContext{Role: "user"})

	mw := RequireRole("superadmin")
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	_ = handler(c)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
