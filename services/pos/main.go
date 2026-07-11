package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/larisai/pos-service/internal/config"
	"github.com/larisai/pos-service/internal/handlers/http"
	"github.com/larisai/pos-service/internal/middleware"
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
	analyticsRepo := repositories.NewAnalyticsRepository()
	userRepo := repositories.NewUserRepository()

	// 3. Init Services
	productSvc := services.NewProductService(productRepo)
	checkoutSvc := services.NewCheckoutService(productRepo, txRepo)
	analyticsSvc := services.NewAnalyticsService(analyticsRepo)
	authSvc := services.NewAuthService(userRepo)
	userSvc := services.NewUserService(userRepo)

	// AI Engine URL
	aiEngineURL := os.Getenv("AI_ENGINE_URL")
	if aiEngineURL == "" {
		aiEngineURL = "http://localhost:8001"
	}

	// Seed Admin
	if err := authSvc.SeedAdmin(context.Background()); err != nil {
		log.Printf("⚠️ Gagal melakukan seeding admin: %v", err)
	}

	// 4. Init Handlers
	posHandler := http.NewPOSHandler(productSvc, checkoutSvc)
	analyticsHandler := http.NewAnalyticsHandler(analyticsSvc)
	authHandler := http.NewAuthHandler(authSvc)
	userHandler := http.NewUserHandler(userSvc)
	aiHandler := http.NewAIHandler(aiEngineURL)

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
	app.Post("/api/v1/auth/login", authHandler.Login)
	
	// PROTECTED ROUTES
	protectedApi := app.Group("/api/v1", middleware.Protected())
	posHandler.RegisterRoutes(protectedApi)
	analyticsHandler.RegisterRoutes(protectedApi)

	// AI ROUTES
	aiApi := protectedApi.Group("/ai")
	aiApi.Get("/forecasting/stockouts", aiHandler.GetStockoutPredictions)
	aiApi.Get("/clustering/customers", aiHandler.GetCustomerClusters)
	aiApi.Post("/promo/send", aiHandler.SendPromo)

	// ADMIN ONLY ROUTES
	adminApi := protectedApi.Group("/users", middleware.AdminOnly())
	adminApi.Get("/", userHandler.GetUsers)
	adminApi.Post("/", userHandler.CreateUser)
	adminApi.Delete("/:id", userHandler.DeleteUser)

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
