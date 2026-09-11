package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/handlers"
	appMiddleware "github.com/cecep-azhar/jurnalumi/internal/middleware"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Port 5432 untuk container utama pg5432
		dsn = "host=127.0.0.1 user=postgres password=postgres dbname=jurnalumi port=5432 sslmode=disable"
	}

	// Connect Database & Migrate
	db.InitDB(dsn)

	e := echo.New()

	// Setup Sessions (using secure cookie store)
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		if os.Getenv("APP_ENV") == "production" {
			log.Fatal("SESSION_SECRET is required in production")
		}
		secret = "jurnalumi-super-secret-key"
	}
	store := sessions.NewCookieStore([]byte(secret))
	e.Use(session.Middleware(store))

	// Global Middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
		ContentSecurityPolicy: "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://unpkg.com; img-src 'self' data: https:; font-src 'self' data:; frame-ancestors 'self';",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf_token",
	}))



	rateLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(5))
	
	// Landing Page Route
	e.GET("/", func(c echo.Context) error {
		return c.File("web/views/landing.html")
	})

	// Auth Routes
	e.GET("/login", handlers.LoginGET)
	e.POST("/login", handlers.LoginPOST, rateLimiter)
	e.GET("/register", handlers.RegisterGET)
	e.POST("/register", handlers.RegisterPOST, rateLimiter)
	e.GET("/logout", handlers.LogoutGET)

	adminGroup := e.Group("/admin")
	adminGroup.Use(appMiddleware.RequireAuth)
	adminGroup.Use(appMiddleware.RequireRole("superadmin"))
	
	// Super Admin Control Panel Routes
	adminGroup.GET("/dashboard", handlers.AdminDashboardGET)
	adminGroup.POST("/tenant/upgrade", handlers.AdminUpgradeTenantPOST)
	adminGroup.POST("/vouchers/generate", handlers.AdminGenerateVoucherPOST)

	e.POST("/activate-voucher", handlers.ActivateVoucherPOST, appMiddleware.RequireAuth)

	// App Dashboard Route (Protected by Auth Middleware)
	e.GET("/dashboard", handlers.DashboardHandler, appMiddleware.RequireAuth)
	e.POST("/transactions", handlers.TransactionPOST, appMiddleware.RequireAuth, appMiddleware.RequireRole("owner", "spouse", "member"))
	e.POST("/wallets", handlers.WalletPOST, appMiddleware.RequireAuth, appMiddleware.RequireRole("owner", "spouse"))
	e.POST("/categories", handlers.CategoryPOST, appMiddleware.RequireAuth, appMiddleware.RequireRole("owner", "spouse"))

	// Extended Features Routes (Protected)
	e.GET("/assets", handlers.AssetGET, appMiddleware.RequireAuth)
	e.POST("/assets", handlers.AssetPOST, appMiddleware.RequireAuth, appMiddleware.RequireRole("owner", "spouse"))

	// Phase 5: Debt & Protection Routes (Protected)
	e.GET("/debts", handlers.DebtGET, appMiddleware.RequireAuth)
	e.POST("/debts", handlers.DebtPOST, appMiddleware.RequireAuth, appMiddleware.RequireRole("owner", "spouse"))
	e.POST("/debts/pay", handlers.DebtPayPOST, appMiddleware.RequireAuth, appMiddleware.RequireRole("owner", "spouse"))

	e.GET("/reports", handlers.ReportGET, appMiddleware.RequireAuth)
	e.GET("/reports/export", handlers.ReportExportCSV, appMiddleware.RequireAuth, appMiddleware.RequireRole("owner", "spouse", "member", "auditor"))

	e.GET("/family", handlers.FamilyGET, appMiddleware.RequireAuth)
	e.POST("/family", handlers.FamilyPOST, appMiddleware.RequireAuth, appMiddleware.RequireRole("owner", "spouse"))

	// Static files for PWA (Phase 6)
	e.Static("/static", "web/static")

	// Health Check API
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{
			"status": "healthy",
			"app":    "JurnalUmi Go Server",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("Starting JurnalUmi Go Server on port %s...", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Server shutdown with error: %v", err)
	}
}
