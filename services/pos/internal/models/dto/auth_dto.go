package dto

import "github.com/larisai/pos-service/internal/models/entity"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token        string            `json:"token"`
	StoreID      string            `json:"store_id"`
	StoreName    string            `json:"store_name"`
	Phone        string            `json:"phone,omitempty"`
	Subscription *SubscriptionInfo `json:"subscription,omitempty"`
	User         entity.User       `json:"user"`
}

type RegisterMerchantRequest struct {
	Name      string `json:"name"`
	StoreName string `json:"store_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Phone     string `json:"phone"`
	Category  string `json:"category,omitempty"`
}

type RegisterMerchantResponse struct {
	Token        string           `json:"token"`
	StoreID      string           `json:"store_id"`
	StoreName    string           `json:"store_name"`
	Phone        string           `json:"phone,omitempty"`
	Subscription SubscriptionInfo `json:"subscription"`
	User         entity.User      `json:"user"`
}
