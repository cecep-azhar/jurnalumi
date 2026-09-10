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
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Find(&wallets)

	var categories []models.Category
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Find(&categories)

	var transactions []models.Transaction
	db.DB.Scopes(db.Scoped(userCtx.TenantID)).Order("transaction_date desc").Find(&transactions)

	// Calculate Metrics
	var liquidBalance int64 = 0.0
	for _, w := range wallets {
		liquidBalance += w.Balance
	}

	var totalIncome int64 = 0.0
	var totalExpense int64 = 0.0
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
	categoryID := c.FormValue("category_id")
	description := c.FormValue("description")

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amount <= 0 {
		return c.Redirect(http.StatusFound, "/dashboard?error=invalid_amount")
	}

	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		return c.Redirect(http.StatusFound, "/dashboard?error=invalid_wallet")
	}

	catID, err := uuid.Parse(categoryID)
	if err != nil {
		return c.Redirect(http.StatusFound, "/dashboard?error=invalid_category")
	}

	tx := db.DB.Begin()

	var wallet models.Wallet
	if err := tx.Scopes(db.Scoped(userCtx.TenantID)).Where("id = ?", walletID).First(&wallet).Error; err != nil {
		tx.Rollback()
		return c.Redirect(http.StatusFound, "/dashboard?error=unauthorized_wallet")
	}

	var category models.Category
	if err := tx.Where("id = ? AND tenant_id = ?", catID, userCtx.TenantID).First(&category).Error; err != nil {
		tx.Rollback()
		return c.Redirect(http.StatusFound, "/dashboard?error=unauthorized_category")
	}

	transaction := models.Transaction{
		TenantID:        userCtx.TenantID,
		UserID:          userCtx.UserID,
		WalletID:        walletID,
		Type:            transType,
		CategoryID:      &catID,
		CategoryName:    category.Name,
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
	targetStr := c.FormValue("target_amount") // Added for Phase 5 (Sinking Funds)

	balance, _ := strconv.ParseInt(balanceStr, 10, 64)
	targetAmount, _ := strconv.ParseInt(targetStr, 10, 64)

	wallet := models.Wallet{
		TenantID:     userCtx.TenantID,
		Name:         name,
		Type:         walletType,
		Balance:      balance,
		TargetAmount: targetAmount,
	}

	if name == "" || balance < 0 || targetAmount < 0 {
		return c.String(http.StatusBadRequest, "Input tidak valid")
	}

	tx := db.DB.Begin()

	if err := tx.Create(&wallet).Error; err != nil {
		tx.Rollback()
		return c.Redirect(http.StatusFound, "/dashboard?error=failed_create_wallet")
	}

	if balance > 0 {
		var initialCat models.Category
		if err := tx.Where("tenant_id = ? AND name = ?", userCtx.TenantID, "Saldo Awal").First(&initialCat).Error; err != nil {
			initialCat = models.Category{
				TenantID: userCtx.TenantID,
				Name:     "Saldo Awal",
				Type:     "income",
			}
			tx.Create(&initialCat)
		}

		transaction := models.Transaction{
			TenantID:        userCtx.TenantID,
			UserID:          userCtx.UserID,
			WalletID:        wallet.ID,
			Type:            "income",
			CategoryID:      &initialCat.ID,
			CategoryName:    initialCat.Name,
			Amount:          balance,
			Description:     "Saldo Awal Dompet",
			TransactionDate: time.Now(),
		}
		if err := tx.Create(&transaction).Error; err != nil {
			tx.Rollback()
			return c.Redirect(http.StatusFound, "/dashboard?error=failed_create_wallet_tx")
		}
	}

	tx.Commit()

	return c.Redirect(http.StatusFound, "/dashboard")
}

// CategoryPOST handles adding a new master category
func CategoryPOST(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	name := c.FormValue("name")
	categoryType := c.FormValue("type")
	budgetLimitStr := c.FormValue("budget_limit")

	budgetLimit, _ := strconv.ParseInt(budgetLimitStr, 10, 64)

	category := models.Category{
		TenantID:    userCtx.TenantID,
		Type:        categoryType,
		Name:        name,
		BudgetLimit: budgetLimit,
	}

	if name == "" || budgetLimit < 0 {
		return c.String(http.StatusBadRequest, "Input tidak valid")
	}

	if err := db.DB.Create(&category).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Gagal menambahkan kategori")
	}

	return c.Redirect(http.StatusFound, "/dashboard")
}
