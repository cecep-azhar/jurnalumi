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

func TestDebtPOST_Validation(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name        string
		form        url.Values
		expectCode  int
		description string
	}{
		{
			name: "Negative amount rejected",
			form: url.Values{
				"type":         {"debt"},
				"title":        {"Hutang Warung"},
				"counterparty": {"Pak Budi"},
				"total_amount": {"-50000"},
				"interest_rate": {"0"},
			},
			expectCode:  http.StatusBadRequest,
			description: "Amount < 0 must be 400 Bad Request",
		},
		{
			name: "Zero amount rejected",
			form: url.Values{
				"type":         {"debt"},
				"title":        {"Hutang Warung"},
				"counterparty": {"Pak Budi"},
				"total_amount": {"0"},
				"interest_rate": {"0"},
			},
			expectCode:  http.StatusBadRequest,
			description: "Amount == 0 must be 400 Bad Request",
		},
		{
			name: "Empty title rejected",
			form: url.Values{
				"type":         {"debt"},
				"title":        {""},
				"counterparty": {"Pak Budi"},
				"total_amount": {"100000"},
				"interest_rate": {"0"},
			},
			expectCode:  http.StatusBadRequest,
			description: "Empty title must be rejected",
		},
		{
			name: "Negative interest rate rejected",
			form: url.Values{
				"type":         {"debt"},
				"title":        {"Hutang Modal"},
				"counterparty": {"Bank"},
				"total_amount": {"5000000"},
				"interest_rate": {"-5"},
			},
			expectCode:  http.StatusBadRequest,
			description: "Interest rate < 0 must be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/debts", strings.NewReader(tt.form.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set("user_context", middleware.UserContext{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				Role:     "user",
			})

			err := handlers.DebtPOST(c)
			if err != nil {
				t.Fatalf("unexpected handler error: %v", err)
			}

			if rec.Code != tt.expectCode {
				t.Fatalf("expected status %d, got %d: %s", tt.expectCode, rec.Code, rec.Body.String())
			}
		})
	}
}
