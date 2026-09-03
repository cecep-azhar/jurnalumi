package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/cecep-azhar/jurnalumi/internal/db"
	"github.com/cecep-azhar/jurnalumi/internal/handlers"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Port 5433 untuk container pg16
		dsn = "host=127.0.0.1 user=postgres password=[REDACTED] dbname=jurnalumi port=5433 sslmode=disable"
	}

	// Connect Database & Migrate
	db.InitDB(dsn)

	e := echo.New()

	// Global Middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Static Files
	e.Static("/static", "web/static")

	// Landing Page Route
	e.GET("/", func(c echo.Context) error {
		return c.File("web/views/landing.html")
	})

	// App Dashboard Route (Renders Templ via Handler)
	e.GET("/dashboard", handlers.DashboardHandler)

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
