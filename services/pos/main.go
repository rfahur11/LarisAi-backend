package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/larisai/pos-service/internal/config"
	"github.com/larisai/pos-service/internal/handlers/http"
	"github.com/larisai/pos-service/internal/middleware"
	"github.com/larisai/pos-service/internal/repositories"
	"github.com/larisai/pos-service/internal/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	// 0. Load .env file (abaikan error jika di docker/production env sudah ada)
	_ = godotenv.Load()

	// 1. Load config & DB
	config.InitMongoDB()

	// 1.1 Ensure Multi-Tenant Compound Indexes
	ctxIndex, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_, _ = config.DB.Collection("products").Indexes().CreateMany(ctxIndex, []mongo.IndexModel{
		{Keys: bson.D{{Key: "store_id", Value: 1}, {Key: "barcode", Value: 1}}},
		{Keys: bson.D{{Key: "store_id", Value: 1}, {Key: "is_archived", Value: 1}}},
	})
	_, _ = config.DB.Collection("transactions").Indexes().CreateMany(ctxIndex, []mongo.IndexModel{
		{Keys: bson.D{{Key: "store_id", Value: 1}, {Key: "created_at", Value: -1}}},
	})
	_, _ = config.DB.Collection("stores").Indexes().CreateMany(ctxIndex, []mongo.IndexModel{
		{Keys: bson.D{{Key: "store_id", Value: 1}}},
		{Keys: bson.D{{Key: "owner_email", Value: 1}}},
	})
	cancel()

	// 2. Init Repositories
	productRepo := repositories.NewProductRepository()
	txRepo := repositories.NewTransactionRepository()
	analyticsRepo := repositories.NewAnalyticsRepository()
	userRepo := repositories.NewUserRepository()
	storeRepo := repositories.NewStoreRepository()

	// 3. Init Services
	storageSvc := services.NewStorageService()
	productSvc := services.NewProductService(productRepo)
	checkoutSvc := services.NewCheckoutService(productRepo, txRepo)
	analyticsSvc := services.NewAnalyticsService(analyticsRepo)
	authSvc := services.NewAuthService(userRepo, storeRepo)
	userSvc := services.NewUserService(userRepo)
	subscriptionSvc := services.NewSubscriptionService(storeRepo)

	// AI Engine URL
	aiEngineURL := os.Getenv("AI_ENGINE_URL")
	if aiEngineURL == "" {
		aiEngineURL = "http://localhost:8001"
	}

	// Seed Admin & Demo Stores
	if err := authSvc.SeedAdmin(context.Background()); err != nil {
		log.Printf("⚠️ Gagal melakukan seeding admin: %v", err)
	}

	// 4. Init Handlers
	uploadHandler := http.NewUploadHandler(storageSvc)
	posHandler := http.NewPOSHandler(productSvc, checkoutSvc)
	analyticsHandler := http.NewAnalyticsHandler(analyticsSvc)
	authHandler := http.NewAuthHandler(authSvc)
	userHandler := http.NewUserHandler(userSvc)
	aiHandler := http.NewAIHandler(aiEngineURL)
	paymentHandler := http.NewPaymentHandler(txRepo)
	subscriptionHandler := http.NewSubscriptionHandler(subscriptionSvc)

	// 5. Setup Fiber HTTP Server
	app := fiber.New(fiber.Config{
		AppName: "LarisAI POS Service v1.0",
		BodyLimit: 15 * 1024 * 1024, // 15MB for image uploads
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Static images route
	app.Static("/api/v1/images", "./uploads")

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "OK",
			"service": "LarisAI POS Service",
		})
	})

	// Public Auth Routes
	app.Post("/api/v1/auth/login", authHandler.Login)
	app.Post("/api/v1/auth/register-merchant", authHandler.RegisterMerchant)

	// Public Payment & Webhook Routes
	paymentApi := app.Group("/api/v1/payment")
	paymentHandler.RegisterRoutes(paymentApi)

	subscriptionPublicApi := app.Group("/api/v1/subscription")
	subscriptionHandler.RegisterPublicRoutes(subscriptionPublicApi)

	// Protected Routes (POS, Analytics, Uploads, Subscription Status)
	protectedApi := app.Group("/api/v1", middleware.Protected())
	uploadHandler.RegisterRoutes(protectedApi)
	posHandler.RegisterRoutes(protectedApi)
	analyticsHandler.RegisterRoutes(protectedApi)

	subscriptionProtectedApi := protectedApi.Group("/subscription")
	subscriptionHandler.RegisterProtectedRoutes(subscriptionProtectedApi)

	// AI Routes
	aiApi := protectedApi.Group("/ai")
	aiApi.Get("/forecasting/stockouts", aiHandler.GetStockoutPredictions)
	aiApi.Get("/clustering/customers", aiHandler.GetCustomerClusters)
	aiApi.Post("/promo/send", aiHandler.SendPromo)

	// Admin Only Routes
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
