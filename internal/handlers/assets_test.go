package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cecep-azhar/jurnalumi/internal/handlers"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func TestAssetPOST_Validation(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name       string
		form       url.Values
		expectCode int
	}{
		{
			name: "Negative weight rejected",
			form: url.Values{
				"name":      {"Emas Antam"},
				"type":      {"gold_bar"},
				"weight":    {"-10"},
				"karatage":  {"24"},
				"buy_price": {"10000000"},
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Zero weight rejected",
			form: url.Values{
				"name":      {"Emas Antam"},
				"type":      {"gold_bar"},
				"weight":    {"0"},
				"karatage":  {"24"},
				"buy_price": {"10000000"},
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Negative karatage rejected",
			form: url.Values{
				"name":      {"Emas Perhiasan"},
				"type":      {"gold_bar"},
				"weight":    {"5"},
				"karatage":  {"-1"},
				"buy_price": {"4000000"},
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Negative buy price rejected",
			form: url.Values{
				"name":      {"Emas Antam"},
				"type":      {"gold_bar"},
				"weight":    {"5"},
				"karatage":  {"24"},
				"buy_price": {"-500000"},
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Invalid commodity type rejected",
			form: url.Values{
				"name":      {"Crypto Bitcoin"},
				"type":      {"crypto"},
				"weight":    {"1"},
				"karatage":  {"24"},
				"buy_price": {"500000000"},
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Empty name rejected",
			form: url.Values{
				"name":      {""},
				"type":      {"gold_bar"},
				"weight":    {"5"},
				"karatage":  {"24"},
				"buy_price": {"5000000"},
			},
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/assets", strings.NewReader(tt.form.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set("user_context", middleware.UserContext{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				Role:     "user",
			})

			err := handlers.AssetPOST(c)
			if err != nil {
				t.Fatalf("unexpected handler error: %v", err)
			}

			if rec.Code != tt.expectCode {
				t.Fatalf("expected status %d, got %d: %s", tt.expectCode, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAccountRBAC_NonOwnerBlocked(t *testing.T) {
	e := echo.New()

	t.Run("Non-owner cannot export account data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/account/export", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user_context", middleware.UserContext{
			TenantID: uuid.New(),
			UserID:   uuid.New(),
			Role:     "user", // Not owner
		})

		err := handlers.AccountExportGET(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403 Forbidden for non-owner, got: %d", rec.Code)
		}
	})

	t.Run("Non-owner cannot delete tenant account", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/account/delete", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user_context", middleware.UserContext{
			TenantID: uuid.New(),
			UserID:   uuid.New(),
			Role:     "member", // Not owner
		})

		err := handlers.AccountDeletePOST(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403 Forbidden for non-owner, got: %d", rec.Code)
		}
	})
}
