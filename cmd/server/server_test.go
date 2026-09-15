package main_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func TestCSRFProtection(t *testing.T) {
	e := echo.New()

	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf_token",
		CookieName:  "_csrf",
	}))

	e.POST("/test-form", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// 1. Missing CSRF Token should be 400 Bad Request / 403 Forbidden
	formNoToken := url.Values{"action": {"transfer"}}
	reqNoToken := httptest.NewRequest(http.MethodPost, "/test-form", strings.NewReader(formNoToken.Encode()))
	reqNoToken.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recNoToken := httptest.NewRecorder()

	e.ServeHTTP(recNoToken, reqNoToken)
	if recNoToken.Code != http.StatusBadRequest && recNoToken.Code != http.StatusForbidden {
		t.Fatalf("expected 400/403 for missing CSRF token, got %d", recNoToken.Code)
	}

	// 2. Tampered / Mismatched CSRF Token
	formBadToken := url.Values{
		"action":     {"transfer"},
		"csrf_token": {"forged-invalid-token"},
	}
	reqBadToken := httptest.NewRequest(http.MethodPost, "/test-form", strings.NewReader(formBadToken.Encode()))
	reqBadToken.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	reqBadToken.AddCookie(&http.Cookie{
		Name:  "_csrf",
		Value: "legitimate-cookie-token",
	})
	recBadToken := httptest.NewRecorder()

	e.ServeHTTP(recBadToken, reqBadToken)
	if recBadToken.Code != http.StatusBadRequest && recBadToken.Code != http.StatusForbidden {
		t.Fatalf("expected 400/403 for forged CSRF token, got %d", recBadToken.Code)
	}
}

func TestPanicRecoveryIsolation(t *testing.T) {
	e := echo.New()
	e.Use(middleware.Recover())

	e.GET("/panic-route", func(c echo.Context) error {
		panic("simulated database crash or nil pointer")
	})

	e.GET("/healthy-route", func(c echo.Context) error {
		return c.String(http.StatusOK, "healthy")
	})

	// Route that panics must return 500 Internal Server Error, not crash the process
	reqPanic := httptest.NewRequest(http.MethodGet, "/panic-route", nil)
	recPanic := httptest.NewRecorder()
	e.ServeHTTP(recPanic, reqPanic)

	if recPanic.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on panic, got %d", recPanic.Code)
	}

	// Server continues handling subsequent requests normally
	reqHealthy := httptest.NewRequest(http.MethodGet, "/healthy-route", nil)
	recHealthy := httptest.NewRecorder()
	e.ServeHTTP(recHealthy, reqHealthy)

	if recHealthy.Code != http.StatusOK || recHealthy.Body.String() != "healthy" {
		t.Fatalf("expected 200 healthy after panic recovery, got %d: %s", recHealthy.Code, recHealthy.Body.String())
	}
}
