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

func TestTransactionPOST_BoundaryValidation(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name          string
		form          url.Values
		expectedQuery string
		description   string
	}{
		{
			name: "Negative amount rejected",
			form: url.Values{
				"type":        {"income"},
				"amount":      {"-50000"},
				"wallet_id":   {uuid.New().String()},
				"category_id": {uuid.New().String()},
			},
			expectedQuery: "error=invalid_amount",
			description:   "Amount < 0 must redirect with invalid_amount",
		},
		{
			name: "Zero amount rejected",
			form: url.Values{
				"type":        {"expense"},
				"amount":      {"0"},
				"wallet_id":   {uuid.New().String()},
				"category_id": {uuid.New().String()},
			},
			expectedQuery: "error=invalid_amount",
			description:   "Amount == 0 must redirect with invalid_amount",
		},
		{
			name: "Non-numeric amount rejected",
			form: url.Values{
				"type":        {"expense"},
				"amount":      {"abcde"},
				"wallet_id":   {uuid.New().String()},
				"category_id": {uuid.New().String()},
			},
			expectedQuery: "error=invalid_amount",
			description:   "Non-numeric amount must redirect with invalid_amount",
		},
		{
			name: "Malformed wallet UUID rejected",
			form: url.Values{
				"type":        {"expense"},
				"amount":      {"50000"},
				"wallet_id":   {"not-a-valid-uuid"},
				"category_id": {uuid.New().String()},
			},
			expectedQuery: "error=invalid_wallet",
			description:   "Malformed wallet UUID must redirect with invalid_wallet",
		},
		{
			name: "Malformed category UUID rejected",
			form: url.Values{
				"type":        {"income"},
				"amount":      {"50000"},
				"wallet_id":   {uuid.New().String()},
				"category_id": {"12345-invalid"},
			},
			expectedQuery: "error=invalid_category",
			description:   "Malformed category UUID must redirect with invalid_category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(tt.form.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set("user_context", middleware.UserContext{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				Role:     "owner",
			})

			err := handlers.TransactionPOST(c)
			if err != nil {
				t.Fatalf("unexpected handler error: %v", err)
			}

			if rec.Code != http.StatusFound {
				t.Fatalf("expected status 302 Found, got %d: %s", rec.Code, rec.Body.String())
			}

			loc := rec.Header().Get("Location")
			if !strings.Contains(loc, tt.expectedQuery) {
				t.Fatalf("expected redirect to contain %q, got %q", tt.expectedQuery, loc)
			}
		})
	}
}

func TestWalletPOST_FreeTierAndValidation(t *testing.T) {
	e := echo.New()

	t.Run("Free tier cannot create sinking or emergency fund", func(t *testing.T) {
		form := url.Values{
			"name":          {"Emergency Fund"},
			"type":          {"emergency"},
			"balance":       {"1000000"},
			"target_amount": {"5000000"},
		}
		req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user_context", middleware.UserContext{
			TenantID: uuid.New(),
			UserID:   uuid.New(),
			Role:     "user",
		})
		c.Set("is_premium", false)

		err := handlers.WalletPOST(c)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403 Forbidden for free tier sinking/emergency, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("Invalid wallet inputs rejected", func(t *testing.T) {
		invalidCases := []struct {
			name string
			form url.Values
		}{
			{
				name: "Empty wallet name",
				form: url.Values{
					"name":          {""},
					"type":          {"cash"},
					"balance":       {"100000"},
					"target_amount": {"0"},
				},
			},
			{
				name: "Negative balance",
				form: url.Values{
					"name":          {"Dompet"},
					"type":          {"cash"},
					"balance":       {"-10000"},
					"target_amount": {"0"},
				},
			},
			{
				name: "Negative target amount",
				form: url.Values{
					"name":          {"Tabungan"},
					"type":          {"savings"},
					"balance":       {"100000"},
					"target_amount": {"-50000"},
				},
			},
		}

		for _, tc := range invalidCases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(tc.form.Encode()))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				c.Set("user_context", middleware.UserContext{
					TenantID: uuid.New(),
					UserID:   uuid.New(),
					Role:     "owner",
				})
				c.Set("is_premium", true)

				err := handlers.WalletPOST(c)
				if err != nil {
					t.Fatalf("unexpected handler error: %v", err)
				}

				if rec.Code != http.StatusBadRequest {
					t.Fatalf("expected status 400 Bad Request, got %d: %s", rec.Code, rec.Body.String())
				}
			})
		}
	})
}

func TestCategoryPOST_Validation(t *testing.T) {
	e := echo.New()

	invalidCases := []struct {
		name string
		form url.Values
	}{
		{
			name: "Empty category name rejected",
			form: url.Values{
				"name":         {""},
				"type":         {"expense"},
				"budget_limit": {"100000"},
			},
		},
		{
			name: "Negative budget limit rejected",
			form: url.Values{
				"name":         {"Makan"},
				"type":         {"expense"},
				"budget_limit": {"-50000"},
			},
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(tc.form.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set("user_context", middleware.UserContext{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				Role:     "owner",
			})

			err := handlers.CategoryPOST(c)
			if err != nil {
				t.Fatalf("unexpected handler error: %v", err)
			}

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400 Bad Request, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestTransactionDelete_Validation(t *testing.T) {
	e := echo.New()

	t.Run("Malformed transaction UUID rejected", func(t *testing.T) {
		form := url.Values{
			"id": {"invalid-uuid-string"},
		}
		req := httptest.NewRequest(http.MethodPost, "/transactions/delete", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user_context", middleware.UserContext{
			TenantID: uuid.New(),
			UserID:   uuid.New(),
			Role:     "owner",
		})

		err := handlers.TransactionDelete(c)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}

		if rec.Code != http.StatusFound {
			t.Fatalf("expected status 302 Found, got %d", rec.Code)
		}

		loc := rec.Header().Get("Location")
		if !strings.Contains(loc, "error=invalid_transaction") {
			t.Fatalf("expected redirect to contain error=invalid_transaction, got %q", loc)
		}
	})
}
