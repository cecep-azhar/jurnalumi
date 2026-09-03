package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/web/views"
)

// DashboardHandler handles the rendering of the authenticated dashboard
func DashboardHandler(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	var tenant models.Tenant
	db.DB.First(&tenant, "id = ?", userCtx.TenantID)

	user := models.User{
		Name: userCtx.Name,
		Role: userCtx.Role,
	}

	var wallets []models.Wallet
	db.DB.Where("tenant_id = ?", userCtx.TenantID).Find(&wallets)

	var categories []models.Category
	db.DB.Where("tenant_id = ?", userCtx.TenantID).Find(&categories)

	var transactions []models.Transaction
	db.DB.Where("tenant_id = ?", userCtx.TenantID).Order("transaction_date desc").Find(&transactions)

	// Calculate Metrics
	var liquidBalance float64 = 0.0
	for _, w := range wallets {
		liquidBalance += w.Balance
	}

	var totalIncome float64 = 0.0
	var totalExpense float64 = 0.0
	for _, t := range transactions {
		if t.Type == "income" {
			totalIncome += t.Amount
		} else if t.Type == "expense" {
			totalExpense += t.Amount
		}
	}

	return Render(c, views.Dashboard(tenant, user, wallets, categories, transactions, liquidBalance, totalIncome, totalExpense))
}

// TransactionPOST handles inserting a new income/expense into DB
func TransactionPOST(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	transType := c.FormValue("type")
	amountStr := c.FormValue("amount")
	walletIDStr := c.FormValue("wallet_id")
	category := c.FormValue("category")
	description := c.FormValue("description")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return c.Redirect(http.StatusFound, "/dashboard?error=invalid_amount")
	}

	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		return c.Redirect(http.StatusFound, "/dashboard?error=invalid_wallet")
	}

	tx := db.DB.Begin()

	var wallet models.Wallet
	if err := tx.Where("id = ? AND tenant_id = ?", walletID, userCtx.TenantID).First(&wallet).Error; err != nil {
		tx.Rollback()
		return c.Redirect(http.StatusFound, "/dashboard?error=unauthorized_wallet")
	}

	transaction := models.Transaction{
		TenantID:        userCtx.TenantID,
		UserID:          userCtx.UserID,
		WalletID:        walletID,
		Type:            transType,
		CategoryName:    category,
		Amount:          amount,
		Description:     description,
		TransactionDate: time.Now(),
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return c.Redirect(http.StatusFound, "/dashboard?error=failed_transaction")
	}

	if transType == "income" {
		wallet.Balance += amount
	} else if transType == "expense" {
		wallet.Balance -= amount
	}
	tx.Save(&wallet)

	tx.Commit()

	return c.Redirect(http.StatusFound, "/dashboard")
}

// WalletPOST handles adding a new wallet
func WalletPOST(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	name := c.FormValue("name")
	walletType := c.FormValue("type")
	balanceStr := c.FormValue("balance")

	balance, _ := strconv.ParseFloat(balanceStr, 64)

	wallet := models.Wallet{
		TenantID: userCtx.TenantID,
		Name:     name,
		Type:     walletType,
		Balance:  balance,
	}

	db.DB.Create(&wallet)

	return c.Redirect(http.StatusFound, "/dashboard")
}

// CategoryPOST handles adding a new master category
func CategoryPOST(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	name := c.FormValue("name")
	categoryType := c.FormValue("type")

	category := models.Category{
		TenantID: userCtx.TenantID,
		Type:     categoryType,
		Name:     name,
	}

	db.DB.Create(&category)

	return c.Redirect(http.StatusFound, "/dashboard")
}
