package dto

import "time"

type SubscriptionInfo struct {
	PlanTier      string     `json:"plan_tier"`
	Status        string     `json:"status"` // "trial", "active", "grace_period", "expired"
	TrialEndsAt   *time.Time `json:"trial_ends_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	RemainingDays int        `json:"remaining_days"`
	MaxOutlets    int        `json:"max_outlets"`
	IsGracePeriod bool       `json:"is_grace_period"`
}

type SubscriptionStatusResponse struct {
	StoreID       string     `json:"store_id"`
	StoreName     string     `json:"store_name"`
	PlanTier      string     `json:"plan_tier"`
	Status        string     `json:"status"`
	RemainingDays int        `json:"remaining_days"`
	TrialEndsAt   *time.Time `json:"trial_ends_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	MaxOutlets    int        `json:"max_outlets"`
	IsGracePeriod bool       `json:"is_grace_period"`
}

type SubscriptionSyncRequest struct {
	StoreID        string `json:"store_id"`
	PlanTier       string `json:"plan_tier"`       // "subs_monthly", "subs_yearly", "lan_multidevice"
	DurationMonths int    `json:"duration_months"` // 1, 12, or 999 for lifetime
	OrderID        string `json:"order_id"`
}
