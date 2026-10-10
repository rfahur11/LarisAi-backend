package dto

import "github.com/larisai/pos-service/internal/models/entity"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string      `json:"token"`
	StoreID   string      `json:"store_id"`
	StoreName string      `json:"store_name"`
	User      entity.User `json:"user"`
}
