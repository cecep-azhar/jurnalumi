import re

with open("internal/handlers/debts.go", "r") as f:
    content = f.read()

content = content.replace('"github.com/labstack/echo/v4"', '"github.com/google/uuid"\n\t"github.com/labstack/echo/v4"')

content = content.replace('''	var debts []models.Debt
	db.DB.Where("tenant_id = ?", userCtx.TenantID).Find(&debts)''', '''	var debts []models.Debt
	db.DB.Where("tenant_id = ?", userCtx.TenantID).Find(&debts)

	var wallets []models.Wallet
	db.DB.Where("tenant_id = ?", userCtx.TenantID).Find(&wallets)''')

content = content.replace('''	return Render(c, views.DebtManagement(tenant, user, debts, totalDebt, totalReceivable))''', '''	return Render(c, views.DebtManagement(tenant, user, debts, wallets, totalDebt, totalReceivable))''')

new_post = '''func DebtPayPOST(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)
	debtID := c.FormValue("debt_id")
	amountStr := c.FormValue("amount")
	walletID := c.FormValue("wallet_id")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return c.Redirect(http.StatusFound, "/debts")
	}

	xr := db.DB.Begin()
	if xr.Error != nil {
		return c.Redirect(http.StatusFound, "/debts")
	}
	defer xr.Rollback()

	var debt models.Debt
	if err := xr.Where("tenant_id = ? AND id = ?", userCtx.TenantID, debtID).First(&debt).Error; err != nil {
		return c.Redirect(http.StatusFound, "/debts")
	}

	var wallet models.Wallet
	if err := xr.Where("tenant_id = ? AND id = ?", userCtx.TenantID, walletID).First(&wallet).Error; err != nil {
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
}'''

content = re.sub(r'func DebtPayPOST\(c echo\.Context\) error \{[\s\S]*?^\}', new_post, content, flags=re.MULTILINE)

with open("internal/handlers/debts.go", "w") as f:
    f.write(content)
