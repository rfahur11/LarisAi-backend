package services

import (
	"context"
	"errors"
	"time"

	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/models/entity"
	"github.com/larisai/pos-service/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	return s.userRepo.FindAll(ctx)
}

func (s *UserService) CreateUser(ctx context.Context, req dto.CreateUserRequest) error {
	// Check if email exists
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("email already exists")
	}

	// Validate role (fallback to cashier if not admin)
	role := "cashier"
	if req.Role == "admin" {
		role = "admin"
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	newUser := &entity.User{
		ID:           primitive.NewObjectID().Hex(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.userRepo.Create(ctx, newUser)
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepo.Delete(ctx, id)
}
