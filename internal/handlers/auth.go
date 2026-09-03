package handlers

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/web/views"
)

// LoginGET renders the login page
func LoginGET(c echo.Context) error {
	return Render(c, views.Login())
}

// LoginPOST handles the real login authentication against DB
func LoginPOST(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return c.Redirect(http.StatusFound, "/login?error=invalid_credentials")
	}

	// Check password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return c.Redirect(http.StatusFound, "/login?error=invalid_credentials")
	}

	// Save Session
	sess, _ := session.Get("jurnalumi_session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 Days
		HttpOnly: true,
	}
	sess.Values["user_id"] = user.ID.String()
	sess.Values["tenant_id"] = user.TenantID.String()
	sess.Values["email"] = user.Email
	sess.Values["name"] = user.Name
	sess.Values["role"] = user.Role
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusFound, "/dashboard")
}

// RegisterGET renders the registration page
func RegisterGET(c echo.Context) error {
	return Render(c, views.Register())
}

// RegisterPOST handles creating Tenant & User in DB
func RegisterPOST(c echo.Context) error {
	familyName := c.FormValue("family_name")
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.Redirect(http.StatusFound, "/register?error=server_error")
	}

	// Transaction to create Tenant & Owner User
	tx := db.DB.Begin()

	tenant := models.Tenant{
		Name: familyName,
		Plan: "free",
	}
	if err := tx.Create(&tenant).Error; err != nil {
		tx.Rollback()
		return c.Redirect(http.StatusFound, "/register?error=tenant_exists")
	}

	user := models.User{
		TenantID:     tenant.ID,
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "owner", // Primary Family Admin
	}
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return c.Redirect(http.StatusFound, "/register?error=email_exists")
	}

	// Create Default Wallet for Family
	defaultWallet := models.Wallet{
		TenantID: tenant.ID,
		Name:     "Dompet Utama (Cash)",
		Type:     "cash",
		Balance:  0.00,
	}
	tx.Create(&defaultWallet)

	// Seed Default Categories for Family
	defaultCategories := []models.Category{
		{TenantID: tenant.ID, Type: "expense", Name: "Belanja Dapur", Color: "red", BudgetLimit: 2000000},
		{TenantID: tenant.ID, Type: "expense", Name: "Listrik & Air", Color: "yellow", BudgetLimit: 500000},
		{TenantID: tenant.ID, Type: "expense", Name: "Transport & Bensin", Color: "blue", BudgetLimit: 500000},
		{TenantID: tenant.ID, Type: "expense", Name: "Jajan & Hiburan", Color: "purple", BudgetLimit: 500000},
		{TenantID: tenant.ID, Type: "expense", Name: "Sosial & Zakat", Color: "emerald", BudgetLimit: 250000},
		{TenantID: tenant.ID, Type: "income", Name: "Gaji Utama", Color: "emerald", BudgetLimit: 0},
		{TenantID: tenant.ID, Type: "income", Name: "Bonus / THR", Color: "blue", BudgetLimit: 0},
		{TenantID: tenant.ID, Type: "income", Name: "Hasil Sampingan", Color: "purple", BudgetLimit: 0},
	}
	for _, cat := range defaultCategories {
		tx.Create(&cat)
	}

	tx.Commit()

	// Auto-login session after register
	sess, _ := session.Get("jurnalumi_session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
	}
	sess.Values["user_id"] = user.ID.String()
	sess.Values["tenant_id"] = user.TenantID.String()
	sess.Values["email"] = user.Email
	sess.Values["name"] = user.Name
	sess.Values["role"] = user.Role
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusFound, "/dashboard")
}

// LogoutGET handles destroying the session
func LogoutGET(c echo.Context) error {
	sess, _ := session.Get("jurnalumi_session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusFound, "/login")
}
