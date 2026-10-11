package services

import (
	"context"
	"time"

	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/models/entity"
	"github.com/larisai/pos-service/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionService struct {
	storeRepo *repositories.StoreRepository
}

func NewSubscriptionService(storeRepo *repositories.StoreRepository) *SubscriptionService {
	return &SubscriptionService{storeRepo: storeRepo}
}

func (s *SubscriptionService) GetStatus(ctx context.Context, storeID string) (*dto.SubscriptionStatusResponse, error) {
	store, err := s.storeRepo.FindByStoreID(ctx, storeID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Fallback jika store belum tercatat (misal store demo lama)
	if store == nil {
		trialEnd := now.Add(14 * 24 * time.Hour)
		return &dto.SubscriptionStatusResponse{
			StoreID:       storeID,
			StoreName:     "Toko Laris UMKM",
			PlanTier:      entity.PlanTierTrial,
			Status:        entity.SubscriptionStatusTrial,
			RemainingDays: 14,
			TrialEndsAt:   &trialEnd,
			ExpiresAt:     nil,
			MaxOutlets:    1,
			IsGracePeriod: false,
		}, nil
	}

	remainingDays := 0
	status := store.SubscriptionStatus
	isGracePeriod := false

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
			if remainingDays == 0 && hoursLeft > 0 {
				remainingDays = 1 // Hari terakhir
			}
		} else {
			// Sudah lewat masa aktif -> cek grace period 3 hari
			if diffDays >= -3 {
				status = entity.SubscriptionStatusGracePeriod
				isGracePeriod = true
				remainingDays = 3 + diffDays
				if remainingDays < 0 {
					remainingDays = 0
				}
			} else {
				status = entity.SubscriptionStatusExpired
				isGracePeriod = false
				remainingDays = 0
			}
		}
	}

	return &dto.SubscriptionStatusResponse{
		StoreID:       store.StoreID,
		StoreName:     store.StoreName,
		PlanTier:      store.PlanTier,
		Status:        status,
		RemainingDays: remainingDays,
		TrialEndsAt:   store.TrialEndsAt,
		ExpiresAt:     store.ExpiresAt,
		MaxOutlets:    store.MaxOutlets,
		IsGracePeriod: isGracePeriod,
	}, nil
}

func (s *SubscriptionService) SyncSubscription(ctx context.Context, req *dto.SubscriptionSyncRequest) (*dto.SubscriptionStatusResponse, error) {
	store, err := s.storeRepo.FindByStoreID(ctx, req.StoreID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var newExpiresAt *time.Time

	// 1. Cek jika paket Lifetime (LAN Resto atau duration >= 999 bulan)
	if req.PlanTier == entity.PlanTierLanMulti || req.DurationMonths >= 999 {
		lifetime := now.AddDate(100, 0, 0)
		newExpiresAt = &lifetime
	} else {
		// 2. Paket Bulanan / Tahunan
		months := req.DurationMonths
		if months <= 0 {
			if req.PlanTier == entity.PlanTierYearly {
				months = 12
			} else {
				months = 1
			}
		}

		baseTime := now
		if store != nil && store.ExpiresAt != nil && store.ExpiresAt.After(now) {
			baseTime = *store.ExpiresAt
		}
		exp := baseTime.AddDate(0, months, 0)
		newExpiresAt = &exp
	}

	maxOutlets := 1
	if req.PlanTier == entity.PlanTierLanMulti {
		maxOutlets = 4 // 1 PC Host + 3 Waiter
	}

	if store == nil {
		store = &entity.Store{
			ID:                 primitive.NewObjectID().Hex(),
			StoreID:            req.StoreID,
			StoreName:          "Toko " + req.StoreID,
			PlanTier:           req.PlanTier,
			SubscriptionStatus: entity.SubscriptionStatusActive,
			ExpiresAt:          newExpiresAt,
			MaxOutlets:         maxOutlets,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := s.storeRepo.Create(ctx, store); err != nil {
			return nil, err
		}
	} else {
		if err := s.storeRepo.UpdateSubscription(ctx, req.StoreID, req.PlanTier, entity.SubscriptionStatusActive, newExpiresAt); err != nil {
			return nil, err
		}
	}

	remainingDays := 0
	if newExpiresAt != nil {
		diff := int(newExpiresAt.Sub(now).Hours() / 24)
		if diff > 0 {
			remainingDays = diff
		}
	}

	return &dto.SubscriptionStatusResponse{
		StoreID:       req.StoreID,
		StoreName:     store.StoreName,
		PlanTier:      req.PlanTier,
		Status:        entity.SubscriptionStatusActive,
		RemainingDays: remainingDays,
		TrialEndsAt:   store.TrialEndsAt,
		ExpiresAt:     newExpiresAt,
		MaxOutlets:    maxOutlets,
		IsGracePeriod: false,
	}, nil
}
