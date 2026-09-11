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
	"github.com/cecep-azhar/jurnalumi/internal/scheduler"
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

	// Start background scheduler (QA-P1-19 + QA-P1-14)
	scheduler.Start()
	defer scheduler.Stop()

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
	e.GET("/privacy", func(c echo.Context) error {
		return c.File("web/views/privacy.html")
	})
	e.GET("/terms", func(c echo.Context) error {
		return c.File("web/views/terms.html")
	})

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

	appGroup := e.Group("")
	appGroup.Use(appMiddleware.RequireAuth)
	appGroup.Use(appMiddleware.CheckPlan)

	// App Dashboard Route (Protected by Auth Middleware)
	appGroup.GET("/dashboard", handlers.DashboardHandler)
	appGroup.POST("/transactions", handlers.TransactionPOST, appMiddleware.RequireRole("owner", "spouse", "member"))
	appGroup.POST("/transactions/delete", handlers.TransactionDelete, appMiddleware.RequireRole("owner", "spouse", "member"))
	appGroup.POST("/wallets", handlers.WalletPOST, appMiddleware.RequireRole("owner", "spouse"))
	appGroup.POST("/categories", handlers.CategoryPOST, appMiddleware.RequireRole("owner", "spouse"))

	// Extended Features Routes (Protected)
	appGroup.GET("/assets", handlers.AssetGET, appMiddleware.RequirePremiumFeature("Aset & Logam Mulia"))
	appGroup.POST("/assets", handlers.AssetPOST, appMiddleware.RequireRole("owner", "spouse"), appMiddleware.RequirePremiumFeature("Aset & Logam Mulia"))

	// Phase 5: Debt & Protection Routes (Protected)
	appGroup.GET("/debts", handlers.DebtGET)
	appGroup.POST("/debts", handlers.DebtPOST, appMiddleware.RequireRole("owner", "spouse"))
	appGroup.POST("/debts/pay", handlers.DebtPayPOST, appMiddleware.RequireRole("owner", "spouse"))

	appGroup.GET("/reports", handlers.ReportGET)
	appGroup.GET("/reports/export", handlers.ReportExportCSV, appMiddleware.RequireRole("owner", "spouse", "member", "auditor"))

	appGroup.GET("/family", handlers.FamilyGET)
	appGroup.POST("/family", handlers.FamilyPOST, appMiddleware.RequireRole("owner", "spouse"))

	// Account settings & data management (UU PDP compliance)
	appGroup.GET("/account", handlers.AccountGET)
	appGroup.GET("/account/export", handlers.AccountExportGET, appMiddleware.RequireRole("owner"))
	appGroup.POST("/account/delete", handlers.AccountDeletePOST, appMiddleware.RequireRole("owner"))

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
