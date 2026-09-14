package handlers

import (
	"net/http"
	"os"
	"time"

	"crypto/rand"
	"encoding/hex"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/models"
	"github.com/cecep-azhar/jurnalumi/internal/services"
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

	if !user.IsVerified {
		return c.Redirect(http.StatusFound, "/login?error=email_not_verified")
	}

	// Check password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return c.Redirect(http.StatusFound, "/login?error=invalid_credentials")
	}

	// Session Rotation: destroy old session completely, create new one
	sess, _ := session.Get("jurnalumi_session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())

	sess, _ = session.Get("jurnalumi_session", c)
	isProd := os.Getenv("APP_ENV") == "production"

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 Days
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
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

	// Generate verification token
	bytes := make([]byte, 32)
	rand.Read(bytes)
	verifyToken := hex.EncodeToString(bytes)

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
		IsVerified:   false,
		VerifyToken:  verifyToken,
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
	
	// Send verification email
	go func() {
		_ = services.SendVerificationEmail(email, verifyToken)
	}()

	return c.Redirect(http.StatusFound, "/login?success=registered")
}

// VerifyEmailGET handles email verification link
func VerifyEmailGET(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.Redirect(http.StatusFound, "/login?error=invalid_token")
	}

	var user models.User
	if err := db.DB.Where("verify_token = ?", token).First(&user).Error; err != nil {
		return c.Redirect(http.StatusFound, "/login?error=invalid_token")
	}

	if user.IsVerified {
		return c.Redirect(http.StatusFound, "/login?success=already_verified")
	}

	user.IsVerified = true
	user.VerifyToken = ""
	if err := db.DB.Save(&user).Error; err != nil {
		return c.Redirect(http.StatusFound, "/login?error=server_error")
	}

	return c.Redirect(http.StatusFound, "/login?success=verified")
}

// LogoutGET handles destroying the session
func LogoutGET(c echo.Context) error {
	sess, _ := session.Get("jurnalumi_session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusFound, "/login")
}

// ForgotPasswordGET renders forgot password request page
func ForgotPasswordGET(c echo.Context) error {
	return Render(c, views.ForgotPassword())
}

// ForgotPasswordPOST generates reset token and sends email
func ForgotPasswordPOST(c echo.Context) error {
	email := c.FormValue("email")
	if email == "" {
		return c.Redirect(http.StatusFound, "/forgot-password")
	}

	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err == nil {
		bytes := make([]byte, 32)
		rand.Read(bytes)
		token := hex.EncodeToString(bytes)
		expires := time.Now().Add(1 * time.Hour)

		user.ResetToken = token
		user.ResetExpires = &expires
		if err := db.DB.Save(&user).Error; err == nil {
			go func() {
				_ = services.SendResetPasswordEmail(user.Email, token)
			}()
		}
	}

	// Always redirect with success message to prevent user enumeration
	return c.Redirect(http.StatusFound, "/forgot-password?success=sent")
}

// ResetPasswordGET renders reset password page
func ResetPasswordGET(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.Redirect(http.StatusFound, "/login?error=invalid_token")
	}

	var user models.User
	if err := db.DB.Where("reset_token = ?", token).First(&user).Error; err != nil {
		return c.Redirect(http.StatusFound, "/login?error=invalid_token")
	}

	if user.ResetExpires == nil || user.ResetExpires.Before(time.Now()) {
		return c.Redirect(http.StatusFound, "/reset-password?token="+token+"&error=invalid_or_expired")
	}

	return Render(c, views.ResetPassword(token))
}

// ResetPasswordPOST updates user's password with provided token
func ResetPasswordPOST(c echo.Context) error {
	token := c.FormValue("token")
	password := c.FormValue("password")

	if token == "" {
		return c.Redirect(http.StatusFound, "/login?error=invalid_token")
	}

	if password == "" {
		return c.Redirect(http.StatusFound, "/reset-password?token="+token+"&error=empty_password")
	}

	var user models.User
	if err := db.DB.Where("reset_token = ?", token).First(&user).Error; err != nil {
		return c.Redirect(http.StatusFound, "/reset-password?token="+token+"&error=invalid_or_expired")
	}

	if user.ResetExpires == nil || user.ResetExpires.Before(time.Now()) {
		return c.Redirect(http.StatusFound, "/reset-password?token="+token+"&error=invalid_or_expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.Redirect(http.StatusFound, "/reset-password?token="+token+"&error=server_error")
	}

	user.PasswordHash = string(hashedPassword)
	user.ResetToken = ""
	user.ResetExpires = nil
	// Resetting password verifies email if not yet verified
	user.IsVerified = true
	user.VerifyToken = ""

	if err := db.DB.Save(&user).Error; err != nil {
		return c.Redirect(http.StatusFound, "/reset-password?token="+token+"&error=server_error")
	}

	return c.Redirect(http.StatusFound, "/login?success=password_reset")
}
