package handlers_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/cecep-azhar/jurnalumi/internal/handlers"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func TestMayarWebhookSignatureVerification(t *testing.T) {
	e := echo.New()
	secret := "super-secure-secret"
	os.Setenv("MAYAR_WEBHOOK_SECRET", secret)
	defer os.Unsetenv("MAYAR_WEBHOOK_SECRET")

	body := []byte(`{"event":"payment.success","data":{"id":"trx_123","status":"SUCCESS","amount":9000}}`)

	t.Run("Invalid signature rejected with 401 Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks/mayar", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("X-Mayar-Signature", "tampered-bad-signature")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handlers.MayarWebhookPOST(c)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 Unauthorized, got %d", rec.Code)
		}
	})

	t.Run("Missing signature when secret configured rejected with 401 Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhooks/mayar", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handlers.MayarWebhookPOST(c)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 Unauthorized, got %d", rec.Code)
		}
	})

	t.Run("Valid HMAC-SHA256 signature accepted past auth check", func(t *testing.T) {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))

		req := httptest.NewRequest(http.MethodPost, "/webhooks/mayar", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("X-Mayar-Signature", sig)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handlers.MayarWebhookPOST(c)

		// Must pass signature check (not 401). Might be 400 or 500 depending on DB state, but signature is valid.
		if rec.Code == http.StatusUnauthorized {
			t.Fatalf("valid signature should not return 401 Unauthorized")
		}
	})
}

func TestFamilyPOST_ValidationAndRBAC(t *testing.T) {
	e := echo.New()

	t.Run("Free tier user blocked from adding family member", func(t *testing.T) {
		form := url.Values{
			"name":     {"Istri Tercinta"},
			"email":    {"istri@example.com"},
			"role":     {"member"},
			"password": {"securepass123"},
		}

		req := httptest.NewRequest(http.MethodPost, "/family", strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user_context", middleware.UserContext{
			TenantID: uuid.New(),
			UserID:   uuid.New(),
			Role:     "owner",
		})
		c.Set("is_premium", false)

		err := handlers.FamilyPOST(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for free user, got %d", rec.Code)
		}
	})

	t.Run("Invalid inputs rejected with 400 Bad Request for premium tenant", func(t *testing.T) {
		invalidCases := []struct {
			name string
			form url.Values
		}{
			{
				name: "Empty name",
				form: url.Values{
					"name":     {""},
					"email":    {"istri@example.com"},
					"role":     {"member"},
					"password": {"password123"},
				},
			},
			{
				name: "Empty email",
				form: url.Values{
					"name":     {"Istri"},
					"email":    {""},
					"role":     {"member"},
					"password": {"password123"},
				},
			},
			{
				name: "Empty password",
				form: url.Values{
					"name":     {"Istri"},
					"email":    {"istri@example.com"},
					"role":     {"member"},
					"password": {""},
				},
			},
			{
				name: "Illegal role escalation",
				form: url.Values{
					"name":     {"Istri"},
					"email":    {"istri@example.com"},
					"role":     {"superadmin"}, // Invalid: only admin/member allowed
					"password": {"password123"},
				},
			},
		}

		for _, tc := range invalidCases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, "/family", strings.NewReader(tc.form.Encode()))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				c.Set("user_context", middleware.UserContext{
					TenantID: uuid.New(),
					UserID:   uuid.New(),
					Role:     "owner",
				})
				c.Set("is_premium", true)

				err := handlers.FamilyPOST(c)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if rec.Code != http.StatusBadRequest {
					t.Fatalf("expected 400 Bad Request, got %d: %s", rec.Code, rec.Body.String())
				}
			})
		}
	})
}
