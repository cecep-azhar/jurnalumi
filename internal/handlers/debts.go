package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/web/views"
	"github.com/google/uuid"
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
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Find(&debts)

	var wallets []models.Wallet
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Find(&wallets)

	var totalDebt int64 = 0
	var totalReceivable int64 = 0
	var snowballDebt *models.Debt = nil

	for i, d := range debts {
		if d.Type == "debt" {
			totalDebt += d.RemainingAmount
			if d.RemainingAmount > 0 {
				if snowballDebt == nil || d.RemainingAmount < snowballDebt.RemainingAmount {
					snowballDebt = &debts[i]
				}
			}
		} else if d.Type == "receivable" {
			totalReceivable += d.RemainingAmount
		}
	}

	return Render(c, views.DebtManagement(tenant, user, debts, wallets, totalDebt, totalReceivable, snowballDebt))
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

	totalAmount, _ := strconv.ParseInt(totalAmountStr, 10, 64)
	interestRate, _ := strconv.ParseInt(interestRateStr, 10, 64)

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

	if title == "" || counterparty == "" || totalAmount <= 0 || interestRate < 0 {
		return c.String(http.StatusBadRequest, "Input tidak valid")
	}

	if err := db.DB.Create(&debt).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Gagal menyimpan utang/piutang")
	}

	return c.Redirect(http.StatusFound, "/debts")
}

// DebtPayPOST handles partial or full payment of a debt (mock implementation for phase 5)
func DebtPayPOST(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)
	debtID := c.FormValue("debt_id")
	amountStr := c.FormValue("amount")
	walletID := c.FormValue("wallet_id")

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amount <= 0 {
		return c.Redirect(http.StatusFound, "/debts")
	}

	xr := db.DB.Begin()
	if xr.Error != nil {
		return c.Redirect(http.StatusFound, "/debts")
	}
	defer xr.Rollback()

	var debt models.Debt
	if err := xr.Scopes(db.Scoped(userCtx.TenantID)).Where("id = ?", debtID).First(&debt).Error; err != nil {
		return c.Redirect(http.StatusFound, "/debts")
	}

	var wallet models.Wallet
	if err := xr.Scopes(db.Scoped(userCtx.TenantID)).Where("id = ?", walletID).First(&wallet).Error; err != nil {
		return c.Redirect(http.StatusFound, "/debts")
	}

	walletIDParsed, _ := uuid.Parse(walletID)

	trx := models.Transaction{
		TenantID:        userCtx.TenantID,
		UserID:          userCtx.UserID,
		WalletID:        walletIDParsed,
		Type:            "expense", // Default to expense
		CategoryName:    "Pelunasan " + debt.Title,
		Amount:          amount,
		Description:     "Pembayaran utang/piutang ke " + debt.Counterparty,
		TransactionDate: time.Now(),
	}

	if debt.Type == "receivable" {
		trx.Type = "income"
		wallet.Balance += amount
	} else {
		wallet.Balance -= amount
	}

	debt.RemainingAmount -= amount
	if debt.RemainingAmount < 0 {
		debt.RemainingAmount = 0
	}
	if debt.RemainingAmount == 0 {
		debt.Status = "paid"
	}

	if err := xr.Create(&trx).Error; err != nil {
		return c.Redirect(http.StatusFound, "/debts")
	}
	if err := xr.Save(&wallet).Error; err != nil {
		return c.Redirect(http.StatusFound, "/debts")
	}
	if err := xr.Save(&debt).Error; err != nil {
		return c.Redirect(http.StatusFound, "/debts")
	}
	xr.Commit()

	return c.Redirect(http.StatusFound, "/debts")
}
