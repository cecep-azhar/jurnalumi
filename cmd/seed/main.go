package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run cmd/seed/main.go <email> <password>")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]

	_ = godotenv.Load()
	
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=jurnalumi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}
	db.InitDB(dsn)
	database := db.DB

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	
	// Create a default tenant for the superadmin
	tenant := models.Tenant{
		Name: "Superadmin Tenant",
		Plan: "premium",
	}
	if err := database.Create(&tenant).Error; err != nil {
		log.Fatalf("Failed to create tenant: %v", err)
	}

	user := models.User{
		TenantID:     tenant.ID,
		Name:         "Superadmin",
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "superadmin",
	}

	if err := database.Create(&user).Error; err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("User %s created with role superadmin\n", email)
}