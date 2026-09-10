package main

import (
	"log"
	"os"
	"github.com/joho/godotenv"
	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/models"
)

func main() {
    _ = godotenv.Load()
	db.InitDB(os.Getenv("DATABASE_URL"))
	log.Println("Running daily reconciliation...")

	var tenants []models.Tenant
	db.DB.Find(&tenants)

	for _, tenant := range tenants {
		var wallets []models.Wallet
		db.DB.Where("tenant_id = ?", tenant.ID).Find(&wallets)

		for _, w := range wallets {
			var totalIncome int64
			var totalExpense int64

			db.DB.Model(&models.Transaction{}).Where("wallet_id = ? AND type = ?", w.ID, "income").Select("COALESCE(SUM(amount), 0)").Scan(&totalIncome)
			db.DB.Model(&models.Transaction{}).Where("wallet_id = ? AND type = ?", w.ID, "expense").Select("COALESCE(SUM(amount), 0)").Scan(&totalExpense)

			expected := totalIncome - totalExpense
			if w.Balance != expected {
				log.Printf("[WARN] Tenant %s Wallet %s: balance divergence! DB %d vs History %d", tenant.ID, w.Name, w.Balance, expected)
			}
		}
	}
	log.Println("Reconciliation complete.")
}
