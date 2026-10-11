package entity

import "time"

const (
	PlanTierTrial         = "trial"
	PlanTierMonthly       = "subs_monthly"
	PlanTierYearly        = "subs_yearly"
	PlanTierLanMulti      = "lan_multidevice"

	SubscriptionStatusTrial       = "trial"
	SubscriptionStatusActive      = "active"
	SubscriptionStatusGracePeriod = "grace_period"
	SubscriptionStatusExpired     = "expired"
)

type Store struct {
	ID                 string     `json:"id" bson:"_id,omitempty"`
	StoreID            string     `json:"store_id" bson:"store_id"`
	StoreName          string     `json:"store_name" bson:"store_name"`
	OwnerEmail         string     `json:"owner_email" bson:"owner_email"`
	OwnerPhone         string     `json:"owner_phone" bson:"owner_phone"`
	Category           string     `json:"category" bson:"category"`
	PlanTier           string     `json:"plan_tier" bson:"plan_tier"`
	SubscriptionStatus string     `json:"subscription_status" bson:"subscription_status"`
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty" bson:"trial_ends_at,omitempty"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	MaxOutlets         int        `json:"max_outlets" bson:"max_outlets"`
	CreatedAt          time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" bson:"updated_at"`
}
