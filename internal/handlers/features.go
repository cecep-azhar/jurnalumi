package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/web/views"
)

// ReportGET renders the Financial Audit & Report page
func ReportGET(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	var tenant models.Tenant
	db.DB.First(&tenant, "id = ?", userCtx.TenantID)

	user := models.User{
		Name: userCtx.Name,
		Role: userCtx.Role,
	}

	var transactions []models.Transaction
	db.DB.Where("tenant_id = ?", userCtx.TenantID).Order("transaction_date desc").Find(&transactions)

	var totalIncome float64 = 0.0
	var totalExpense float64 = 0.0
	for _, t := range transactions {
		if t.Type == "income" {
			totalIncome += t.Amount
		} else if t.Type == "expense" {
			totalExpense += t.Amount
		}
	}

	return Render(c, views.ReportView(tenant, user, transactions, totalIncome, totalExpense))
}

// ReportExportCSV exports transactions to downloadable CSV format
func ReportExportCSV(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	var transactions []models.Transaction
	db.DB.Where("tenant_id = ?", userCtx.TenantID).Order("transaction_date desc").Find(&transactions)

	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=Laporan_Keuangan_JurnalUmi_%s.csv", time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(c.Response().Writer)
	defer writer.Flush()

	// Write CSV Header
	writer.Write([]string{"ID", "Tanggal", "Tipe", "Kategori", "Keterangan", "Nominal (Rp)"})

	for _, t := range transactions {
		writer.Write([]string{
			t.ID.String(),
			t.TransactionDate.Format("2006-01-02 15:04:05"),
			t.Type,
			t.CategoryName,
			t.Description,
			fmt.Sprintf("%.2f", t.Amount),
		})
	}

	return nil
}

// FamilyGET renders Family User Management
func FamilyGET(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	var tenant models.Tenant
	db.DB.First(&tenant, "id = ?", userCtx.TenantID)

	user := models.User{
		Name: userCtx.Name,
		Role: userCtx.Role,
	}

	var members []models.User
	db.DB.Where("tenant_id = ?", userCtx.TenantID).Find(&members)

	return Render(c, views.FamilyManagement(tenant, user, members))
}

// FamilyPOST creates a new family member (Spouse / Member)
func FamilyPOST(c echo.Context) error {
	userCtx := c.Get("user_context").(middleware.UserContext)

	name := c.FormValue("name")
	email := c.FormValue("email")
	role := c.FormValue("role")
	password := c.FormValue("password")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	member := models.User{
		TenantID:     userCtx.TenantID,
		Name:         name,
		Email:        email,
		Role:         role,
		PasswordHash: string(hashedPassword),
	}

	db.DB.Create(&member)

	return c.Redirect(http.StatusFound, "/family")
}
