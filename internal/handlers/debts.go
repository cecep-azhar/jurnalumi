package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/web/views"
	"github.com/labstack/echo/v4"
)

// DebtGET handles displaying the Debt & Receivable Dashboard
func DebtGET(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	var tenant models.Tenant
	db.DB.First(&tenant, "id = ?", userCtx.TenantID)

	var user models.User
	db.DB.First(&user, "id = ?", userCtx.UserID)

	// Fetch Debts & Receivables
	var debts []models.Debt
	db.DB.Where("tenant_id = ?", userCtx.TenantID).Find(&debts)

	var totalDebt float64 = 0
	var totalReceivable float64 = 0

	for _, d := range debts {
		if d.Type == "debt" {
			totalDebt += d.RemainingAmount
		} else if d.Type == "receivable" {
			totalReceivable += d.RemainingAmount
		}
	}

	return Render(c, views.DebtManagement(tenant, user, debts, totalDebt, totalReceivable))
}

// DebtPOST handles adding a new debt or receivable
func DebtPOST(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	debtType := c.FormValue("type")
	title := c.FormValue("title")
	counterparty := c.FormValue("counterparty")
	totalAmountStr := c.FormValue("total_amount")
	interestRateStr := c.FormValue("interest_rate")
	dueDateStr := c.FormValue("due_date")

	totalAmount, _ := strconv.ParseFloat(totalAmountStr, 64)
	interestRate, _ := strconv.ParseFloat(interestRateStr, 64)

	var dueDate *time.Time
	if dueDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dueDateStr)
		if err == nil {
			dueDate = &parsedDate
		}
	}

	debt := models.Debt{
		TenantID:        userCtx.TenantID,
		Type:            debtType,
		Title:           title,
		Counterparty:    counterparty,
		TotalAmount:     totalAmount,
		RemainingAmount: totalAmount, // Initial remaining is total
		InterestRate:    interestRate,
		DueDate:         dueDate,
	}

	db.DB.Create(&debt)

	return c.Redirect(http.StatusFound, "/debts")
}

// DebtPayPOST handles partial or full payment of a debt (mock implementation for phase 5)
func DebtPayPOST(c echo.Context) error {
	// In real implementation, this should insert a Transaction, reduce Wallet balance, and reduce RemainingAmount
	userCtx := c.Get("user_context").(middleware.UserContext)
	debtID := c.FormValue("debt_id")

	var debt models.Debt
	if err := db.DB.Where("tenant_id = ? AND id = ?", userCtx.TenantID, debtID).First(&debt).Error; err == nil {
		// Mock payment: half the amount for demo
		debt.RemainingAmount = debt.RemainingAmount / 2
		if debt.RemainingAmount < 1000 {
			debt.RemainingAmount = 0
		}
		db.DB.Save(&debt)
	}

	return c.Redirect(http.StatusFound, "/debts")
}
