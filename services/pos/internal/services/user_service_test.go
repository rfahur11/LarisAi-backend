package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/larisai/pos-service/internal/config"
	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/repositories"
	"github.com/larisai/pos-service/internal/services"
)

func TestUserService(t *testing.T) {
	// 1. Setup DB for test
	config.InitMongoDB()

	// 2. Setup repo and service
	userRepo := repositories.NewUserRepository()
	userSvc := services.NewUserService(userRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 3. Test Create Cashier
	req := dto.CreateUserRequest{
		Name:     "Test Kasir 1",
		Email:    "kasir1_test@larisai.com",
		Password: "password123",
		Role:     "cashier",
	}

	err := userSvc.CreateUser(ctx, req)
	if err != nil {
		if err.Error() == "email already exists" {
			t.Log("Warning: Test user already exists, skipping creation.")
		} else {
			t.Fatalf("Failed to create user: %v", err)
		}
	} else {
		t.Log("Successfully created user")
	}

	// 4. Clean up (Optional, but let's find the user and delete)
	existingUser, _ := userRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		userRepo.Delete(ctx, existingUser.ID.Hex())
		t.Log("Successfully cleaned up test user")
	}
}
