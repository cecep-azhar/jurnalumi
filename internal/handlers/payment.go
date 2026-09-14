package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MayarWebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		ID            string `json:"id"`
		TransactionID string `json:"transaction_id"`
		Status        string `json:"status"`
		Amount        int64  `json:"amount"`
		Customer      struct {
			Email string `json:"email"`
		} `json:"customer"`
		Metadata struct {
			TenantID string `json:"tenant_id"`
		} `json:"metadata"`
		PaymentMethod string `json:"payment_method"`
	} `json:"data"`
}

// MayarWebhookPOST handles incoming webhooks from Mayar.id
func MayarWebhookPOST(c echo.Context) error {
	secret := os.Getenv("MAYAR_WEBHOOK_SECRET")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid request body")
	}

	// Verify signature if secret configured
	if secret != "" {
		signature := c.Request().Header.Get("X-Mayar-Signature")
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			return c.String(http.StatusUnauthorized, "Invalid signature")
		}
	}

	var payload MayarWebhookPayload
	if err := c.Echo().JSONSerializer.Deserialize(c, &payload); err != nil {
		return c.String(http.StatusBadRequest, "Invalid JSON payload")
	}

	if payload.Event != "payment.success" && payload.Data.Status != "SUCCESS" {
		return c.JSON(http.StatusOK, echo.Map{"status": "ignored", "reason": "not a successful payment"})
	}

	externalID := payload.Data.ID
	if externalID == "" {
		externalID = payload.Data.TransactionID
	}
	if externalID == "" {
		return c.String(http.StatusBadRequest, "Missing external ID")
	}

	// Idempotency check: check if payment already recorded
	var existing models.Payment
	if err := db.DB.Where("external_id = ?", externalID).First(&existing).Error; err == nil {
		return c.JSON(http.StatusOK, echo.Map{"status": "already_processed"})
	}

	// Resolve tenant ID
	var tenant models.Tenant
	var tenantID uuid.UUID

	if payload.Data.Metadata.TenantID != "" {
		tenantID, _ = uuid.Parse(payload.Data.Metadata.TenantID)
	}

	if tenantID != uuid.Nil {
		db.DB.First(&tenant, "id = ?", tenantID)
	} else if payload.Data.Customer.Email != "" {
		var user models.User
		if err := db.DB.Where("email = ?", payload.Data.Customer.Email).First(&user).Error; err == nil && user.TenantID != uuid.Nil {
			tenantID = user.TenantID
			db.DB.First(&tenant, "id = ?", tenantID)
		}
	}

	if tenant.ID == uuid.Nil {
		return c.String(http.StatusNotFound, "Tenant not found for payment")
	}

	// Transaction: record payment + upgrade tenant plan
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	paymentMethod := payload.Data.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "mayar"
	}

	payment := models.Payment{
		TenantID:      tenant.ID,
		ExternalID:    externalID,
		Amount:        payload.Data.Amount,
		Status:        "success",
		PaymentMethod: paymentMethod,
		PaidAt:        &now,
	}

	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		return c.String(http.StatusInternalServerError, "Failed to record payment")
	}

	tenant.Plan = "premium"
	expiry := now.AddDate(1, 0, 0) // default 1 year
	tenant.PlanExpiresAt = &expiry
	if err := tx.Save(&tenant).Error; err != nil {
		tx.Rollback()
		return c.String(http.StatusInternalServerError, "Failed to upgrade tenant")
	}

	if err := tx.Commit().Error; err != nil {
		return c.String(http.StatusInternalServerError, "Failed to commit transaction")
	}

	return c.JSON(http.StatusOK, echo.Map{"status": "success"})
}
