package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func TestMayarWebhookPOST(t *testing.T) {
	if db.DB == nil {
		t.Skip("Database not initialized, skipping test")
	}

	e := echo.New()

	// Setup tenant
	tenant := models.Tenant{
		Name: "Test Family",
		Plan: "free",
	}
	if err := db.DB.Create(&tenant).Error; err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	secret := "test-secret"
	os.Setenv("MAYAR_WEBHOOK_SECRET", secret)
	defer os.Unsetenv("MAYAR_WEBHOOK_SECRET")

	extID := uuid.New().String()
	payload := MayarWebhookPayload{
		Event: "payment.success",
	}
	payload.Data.ID = extID
	payload.Data.Status = "SUCCESS"
	payload.Data.Amount = 9000
	payload.Data.Metadata.TenantID = tenant.ID.String()
	payload.Data.PaymentMethod = "qris"

	body, _ := json.Marshal(payload)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/webhooks/mayar", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Mayar-Signature", sig)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := MayarWebhookPOST(c); err != nil {
		t.Fatalf("MayarWebhookPOST returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify tenant upgraded
	var updated models.Tenant
	db.DB.First(&updated, "id = ?", tenant.ID)
	if updated.Plan != "premium" {
		t.Errorf("expected plan premium, got %s", updated.Plan)
	}
	if updated.PlanExpiresAt == nil || updated.PlanExpiresAt.Before(time.Now()) {
		t.Errorf("expected valid plan_expires_at, got %v", updated.PlanExpiresAt)
	}

	// Idempotency check: send again with same ID
	req2 := httptest.NewRequest(http.MethodPost, "/webhooks/mayar", bytes.NewReader(body))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req2.Header.Set("X-Mayar-Signature", sig)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	if err := MayarWebhookPOST(c2); err != nil {
		t.Fatalf("MayarWebhookPOST 2nd call failed: %v", err)
	}
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected status 200 on retry, got %d", rec2.Code)
	}
}
