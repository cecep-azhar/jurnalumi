package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/cecep-azhar/jurnalumi/internal/handlers"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
)

func TestDebtPayPOST_InputValidation(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name       string
		amount     string
		walletID   string
		debtID     string
		expectCode int
	}{
		{
			name:       "Negative payment amount redirects safely",
			amount:     "-50000",
			walletID:   uuid.New().String(),
			debtID:     uuid.New().String(),
			expectCode: http.StatusFound,
		},
		{
			name:       "Zero payment amount redirects safely",
			amount:     "0",
			walletID:   uuid.New().String(),
			debtID:     uuid.New().String(),
			expectCode: http.StatusFound,
		},
		{
			name:       "Non-numeric payment amount redirects safely",
			amount:     "abc",
			walletID:   uuid.New().String(),
			debtID:     uuid.New().String(),
			expectCode: http.StatusFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := make(url.Values)
			f.Set("amount", tc.amount)
			f.Set("wallet_id", tc.walletID)
			f.Set("debt_id", tc.debtID)

			req := httptest.NewRequest(http.MethodPost, "/debts/pay", strings.NewReader(f.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_context", middleware.UserContext{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				Role:     "owner",
			})

			err := handlers.DebtPayPOST(c)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectCode, rec.Code)
			assert.Equal(t, "/debts", rec.Header().Get("Location"))
		})
	}
}

func TestAdminGenerateVoucher_DurationFallback(t *testing.T) {
	// Notice: Test admin voucher parameters validation logic
	e := echo.New()

	t.Run("Generate voucher requires superadmin role check", func(t *testing.T) {
		f := make(url.Values)
		f.Set("prefix", "DISC50_")
		f.Set("duration", "30")

		req := httptest.NewRequest(http.MethodPost, "/admin/vouchers/generate", strings.NewReader(f.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Context without superadmin role
		c.Set("user_context", middleware.UserContext{
			TenantID: uuid.New(),
			UserID:   uuid.New(),
			Role:     "member",
		})

		handler := middleware.RequireRole("superadmin")(handlers.AdminGenerateVoucherPOST)
		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}
