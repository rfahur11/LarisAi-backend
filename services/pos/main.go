package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/larisai/pos-service/internal/config"
	"github.com/larisai/pos-service/internal/handlers/http"
	"github.com/larisai/pos-service/internal/repositories"
	"github.com/larisai/pos-service/internal/services"
)

func main() {
	// 0. Load .env file (abaikan error jika di docker/production env sudah ada)
	_ = godotenv.Load()

	// 1. Load config & DB
	config.InitMongoDB()

	// 2. Init Repositories
	productRepo := repositories.NewProductRepository()
	txRepo := repositories.NewTransactionRepository()

	// 3. Init Services
	productSvc := services.NewProductService(productRepo)
	checkoutSvc := services.NewCheckoutService(productRepo, txRepo)

	// 4. Init Handlers
	posHandler := http.NewPOSHandler(productSvc, checkoutSvc)

	// 5. Setup Fiber HTTP Server
	app := fiber.New(fiber.Config{
		AppName: "LarisAI POS Service v1.0",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "OK",
			"service": "LarisAI POS Service",
		})
	})

	// Register Routes
	posHandler.RegisterRoutes(app)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 LarisAI POS Backend berjalan di port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
