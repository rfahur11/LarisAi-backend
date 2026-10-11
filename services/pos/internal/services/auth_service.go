package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/models/entity"
	"github.com/larisai/pos-service/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repositories.UserRepository
	storeRepo *repositories.StoreRepository
	jwtSecret []byte
}

func NewAuthService(userRepo *repositories.UserRepository, storeRepo *repositories.StoreRepository) *AuthService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "super_secret_larisai_key_2026"
	}
	return &AuthService{
		userRepo:  userRepo,
		storeRepo: storeRepo,
		jwtSecret: []byte(secret),
	}
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.userRepo.FindByEmail(ctx, email)
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

	// Store Metadata
	storeID := user.StoreID
	if storeID == "" {
		storeID = "STORE_BERKAH_01"
	}
	storeName := user.StoreName
	if storeName == "" {
		storeName = "Toko Berkah Jaya"
	}

	// Fetch subscription info from storeRepo if exists
	var subInfo *dto.SubscriptionInfo
	if s.storeRepo != nil {
		store, _ := s.storeRepo.FindByStoreID(ctx, storeID)
		if store != nil {
			now := time.Now()
			remainingDays := 0
			status := store.SubscriptionStatus
			isGrace := false

			var targetDate *time.Time
			if store.ExpiresAt != nil {
				targetDate = store.ExpiresAt
			} else if store.TrialEndsAt != nil {
				targetDate = store.TrialEndsAt
			}

			if targetDate != nil {
				hoursLeft := targetDate.Sub(now).Hours()
				diffDays := int(hoursLeft / 24)
				if hoursLeft > 0 {
					remainingDays = diffDays
					if remainingDays == 0 {
						remainingDays = 1
					}
				} else {
					if diffDays >= -3 {
						status = entity.SubscriptionStatusGracePeriod
						isGrace = true
						remainingDays = 3 + diffDays
						if remainingDays < 0 {
							remainingDays = 0
						}
					} else {
						status = entity.SubscriptionStatusExpired
						isGrace = false
						remainingDays = 0
					}
				}
			}

			subInfo = &dto.SubscriptionInfo{
				PlanTier:      store.PlanTier,
				Status:        status,
				TrialEndsAt:   store.TrialEndsAt,
				ExpiresAt:     store.ExpiresAt,
				RemainingDays: remainingDays,
				MaxOutlets:    store.MaxOutlets,
				IsGracePeriod: isGrace,
			}
		}
	}

	// Fallback jika belum ada record store
	if subInfo == nil {
		trialEnd := time.Now().Add(14 * 24 * time.Hour)
		subInfo = &dto.SubscriptionInfo{
			PlanTier:      entity.PlanTierTrial,
			Status:        entity.SubscriptionStatusTrial,
			TrialEndsAt:   &trialEnd,
			ExpiresAt:     nil,
			RemainingDays: 14,
			MaxOutlets:    1,
			IsGracePeriod: false,
		}
	}

	// Generate JWT
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
		Token:        tokenString,
		StoreID:      storeID,
		StoreName:    storeName,
		Phone:        user.Phone,
		Subscription: subInfo,
		User:         *user,
	}, nil
}

// RegisterMerchant handles merchant onboarding with a 14-day free trial
func (s *AuthService) RegisterMerchant(ctx context.Context, req *dto.RegisterMerchantRequest) (*dto.RegisterMerchantResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.StoreName) == "" {
		return nil, errors.New("name, store_name, email, and password are required")
	}

	if len(strings.TrimSpace(req.Password)) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// 1. Cek email unik
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email is already registered")
	}

	// 2. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(req.Password)), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. Generate Unique StoreID
	now := time.Now()
	uniqueHex := primitive.NewObjectID().Hex()
	storeID := fmt.Sprintf("STORE_%d_%s", now.Unix(), strings.ToUpper(uniqueHex[len(uniqueHex)-4:]))
	storeName := strings.TrimSpace(req.StoreName)
	userName := strings.TrimSpace(req.Name)
	phone := strings.TrimSpace(req.Phone)
	category := strings.TrimSpace(req.Category)
	if category == "" {
		category = "Retail"
	}

	// 4. Inisialisasi 14-day Free Trial
	trialEnd := now.Add(14 * 24 * time.Hour)

	// Buat Store
	newStore := &entity.Store{
		ID:                 primitive.NewObjectID().Hex(),
		StoreID:            storeID,
		StoreName:          storeName,
		OwnerEmail:         email,
		OwnerPhone:         phone,
		Category:           category,
		PlanTier:           entity.PlanTierTrial,
		SubscriptionStatus: entity.SubscriptionStatusTrial,
		TrialEndsAt:        &trialEnd,
		ExpiresAt:          nil,
		MaxOutlets:         1,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if s.storeRepo != nil {
		if err := s.storeRepo.Create(ctx, newStore); err != nil {
			return nil, fmt.Errorf("failed to create store: %w", err)
		}
	}

	// Buat User Owner (Admin)
	newUser := &entity.User{
		ID:           primitive.NewObjectID().Hex(),
		StoreID:      storeID,
		StoreName:    storeName,
		Name:         userName,
		Email:        email,
		Phone:        phone,
		PasswordHash: string(hashedPassword),
		Role:         entity.RoleAdmin,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 5. Generate JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    newUser.ID,
		"email":      newUser.Email,
		"role":       newUser.Role,
		"store_id":   storeID,
		"store_name": storeName,
		"exp":        now.Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &dto.RegisterMerchantResponse{
		Token:     tokenString,
		StoreID:   storeID,
		StoreName: storeName,
		Phone:     phone,
		Subscription: dto.SubscriptionInfo{
			PlanTier:      entity.PlanTierTrial,
			Status:        entity.SubscriptionStatusTrial,
			TrialEndsAt:   &trialEnd,
			ExpiresAt:     nil,
			RemainingDays: 14,
			MaxOutlets:    1,
			IsGracePeriod: false,
		},
		User: *newUser,
	}, nil
}

// SeedAdmin seeds the default multi-tenant demo stores if they don't exist
func (s *AuthService) SeedAdmin(ctx context.Context) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	now := time.Now()
	trialEnd := now.Add(30 * 24 * time.Hour) // Demo store 30 days

	// Toko 1: Toko Berkah Jaya
	admin1, _ := s.userRepo.FindByEmail(ctx, "admin@larisai.com")
	if admin1 == nil {
		admin1 = &entity.User{
			StoreID:      "STORE_BERKAH_01",
			StoreName:    "Toko Berkah Jaya",
			Name:         "Pemilik Toko Berkah",
			Email:        "admin@larisai.com",
			Phone:        "081234567890",
			PasswordHash: string(hash),
			Role:         entity.RoleAdmin,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		_ = s.userRepo.Create(ctx, admin1)
	}

	if s.storeRepo != nil {
		store1, _ := s.storeRepo.FindByStoreID(ctx, "STORE_BERKAH_01")
		if store1 == nil {
			store1 = &entity.Store{
				ID:                 primitive.NewObjectID().Hex(),
				StoreID:            "STORE_BERKAH_01",
				StoreName:          "Toko Berkah Jaya",
				OwnerEmail:         "admin@larisai.com",
				OwnerPhone:         "081234567890",
				Category:           "Kelontong & Sembako",
				PlanTier:           entity.PlanTierTrial,
				SubscriptionStatus: entity.SubscriptionStatusTrial,
				TrialEndsAt:        &trialEnd,
				MaxOutlets:         1,
				CreatedAt:          now,
				UpdatedAt:          now,
			}
			_ = s.storeRepo.Create(ctx, store1)
		}
	}

	// Toko 2: Warung Sejahtera
	admin2, _ := s.userRepo.FindByEmail(ctx, "toko2@larisai.com")
	if admin2 == nil {
		admin2 = &entity.User{
			StoreID:      "STORE_SEJAHTERA_02",
			StoreName:    "Warung Sejahtera UMKM",
			Name:         "Pemilik Warung Sejahtera",
			Email:        "toko2@larisai.com",
			Phone:        "081298765432",
			PasswordHash: string(hash),
			Role:         entity.RoleAdmin,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		_ = s.userRepo.Create(ctx, admin2)
	}

	if s.storeRepo != nil {
		store2, _ := s.storeRepo.FindByStoreID(ctx, "STORE_SEJAHTERA_02")
		if store2 == nil {
			store2 = &entity.Store{
				ID:                 primitive.NewObjectID().Hex(),
				StoreID:            "STORE_SEJAHTERA_02",
				StoreName:          "Warung Sejahtera UMKM",
				OwnerEmail:         "toko2@larisai.com",
				OwnerPhone:         "081298765432",
				Category:           "Kuliner & Warung Makan",
				PlanTier:           entity.PlanTierTrial,
				SubscriptionStatus: entity.SubscriptionStatusTrial,
				TrialEndsAt:        &trialEnd,
				MaxOutlets:         1,
				CreatedAt:          now,
				UpdatedAt:          now,
			}
			_ = s.storeRepo.Create(ctx, store2)
		}
	}

	return nil
}
