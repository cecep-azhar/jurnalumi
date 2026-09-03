package db

import (
	"log"

	"github.com/cecep-azhar/jurnalumi/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Enable UUID extension in PostgreSQL
	DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	DB.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;")

	log.Println("Database connection established. Running migrations...")
	
	err = DB.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.Category{},
		&models.Wallet{},
		&models.CommodityAsset{},
		&models.Debt{},
		&models.Transaction{},
		&models.Voucher{},
	)
	
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	
	log.Println("Database migrated successfully.")
}