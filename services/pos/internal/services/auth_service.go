package services

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/models/entity"
	"github.com/larisai/pos-service/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repositories.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "super_secret_larisai_key_2026"
	}
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(secret),
	}
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT
	storeID := user.StoreID
	if storeID == "" {
		storeID = "STORE_BERKAH_01"
	}
	storeName := user.StoreName
	if storeName == "" {
		storeName = "Toko Berkah Jaya"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"store_id":   storeID,
		"store_name": storeName,
		"exp":        time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days expiration for POS terminal
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token:     tokenString,
		StoreID:   storeID,
		StoreName: storeName,
		User:      *user,
	}, nil
}

// SeedAdmin seeds the default multi-tenant demo stores if they don't exist
func (s *AuthService) SeedAdmin(ctx context.Context) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	// Toko 1: Toko Berkah Jaya
	admin1, _ := s.userRepo.FindByEmail(ctx, "admin@larisai.com")
	if admin1 == nil {
		admin1 = &entity.User{
			StoreID:      "STORE_BERKAH_01",
			StoreName:    "Toko Berkah Jaya",
			Name:         "Pemilik Toko Berkah",
			Email:        "admin@larisai.com",
			PasswordHash: string(hash),
			Role:         entity.RoleAdmin,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		_ = s.userRepo.Create(ctx, admin1)
	}

	// Toko 2: Warung Sejahtera
	admin2, _ := s.userRepo.FindByEmail(ctx, "toko2@larisai.com")
	if admin2 == nil {
		admin2 = &entity.User{
			StoreID:      "STORE_SEJAHTERA_02",
			StoreName:    "Warung Sejahtera UMKM",
			Name:         "Pemilik Warung Sejahtera",
			Email:        "toko2@larisai.com",
			PasswordHash: string(hash),
			Role:         entity.RoleAdmin,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		_ = s.userRepo.Create(ctx, admin2)
	}

	return nil
}
