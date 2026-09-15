package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/labstack/echo/v4"
)

func TestRequireRole_AdminBypassPrevention(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name       string
		userRole   string
		expectCode int
	}{
		{
			name:       "Standard user cannot access superadmin",
			userRole:   "user",
			expectCode: http.StatusForbidden,
		},
		{
			name:       "Tenant owner cannot access superadmin",
			userRole:   "tenant_owner",
			expectCode: http.StatusForbidden,
		},
		{
			name:       "Superadmin allowed",
			userRole:   "superadmin",
			expectCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set("user_context", middleware.UserContext{Role: tt.userRole})

			mw := middleware.RequireRole("superadmin")
			handler := mw(func(c echo.Context) error {
				return c.String(http.StatusOK, "welcome")
			})

			_ = handler(c)
			if rec.Code != tt.expectCode {
				t.Fatalf("expected status %d, got %d", tt.expectCode, rec.Code)
			}
		})
	}
}

func TestRequireRole_MissingContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := middleware.RequireRole("superadmin")
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "welcome")
	})

	_ = handler(c)
	// Missing context redirects to login
	if rec.Code != http.StatusFound {
		t.Fatalf("expected status 302, got %d", rec.Code)
	}
}

func TestRequirePremiumFeature(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name       string
		isPremium  interface{}
		expectCode int
	}{
		{
			name:       "Free user blocked from premium feature",
			isPremium:  false,
			expectCode: http.StatusForbidden,
		},
		{
			name:       "Unset context blocked from premium feature",
			isPremium:  nil,
			expectCode: http.StatusForbidden,
		},
		{
			name:       "Premium user allowed",
			isPremium:  true,
			expectCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/gold-calculator", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.isPremium != nil {
				c.Set("is_premium", tt.isPremium.(bool))
			}

			mw := middleware.RequirePremiumFeature("Simulasi Emas")
			handler := mw(func(c echo.Context) error {
				return c.String(http.StatusOK, "welcome to feature")
			})

			_ = handler(c)
			if rec.Code != tt.expectCode {
				t.Fatalf("expected status %d, got %d", tt.expectCode, rec.Code)
			}
			if rec.Code == http.StatusForbidden && !strings.Contains(rec.Body.String(), "Simulasi Emas") {
				t.Errorf("expected rejection body to mention feature name")
			}
		})
	}
}
