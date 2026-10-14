package repositories

import (
	"context"
	"time"

	"github.com/larisai/pos-service/internal/config"
	"github.com/larisai/pos-service/internal/models/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StoreRepository struct {
	collection *mongo.Collection
}

func NewStoreRepository() *StoreRepository {
	return &StoreRepository{
		collection: config.DB.Collection("stores"),
	}
}

func (r *StoreRepository) FindByStoreID(ctx context.Context, storeID string) (*entity.Store, error) {
	var store entity.Store
	err := r.collection.FindOne(ctx, bson.M{"store_id": storeID}).Decode(&store)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &store, nil
}

func (r *StoreRepository) FindByOwnerEmail(ctx context.Context, email string) (*entity.Store, error) {
	var store entity.Store
	err := r.collection.FindOne(ctx, bson.M{"owner_email": email}).Decode(&store)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &store, nil
}

func (r *StoreRepository) Create(ctx context.Context, store *entity.Store) error {
	_, err := r.collection.InsertOne(ctx, store)
	return err
}

func (r *StoreRepository) Upsert(ctx context.Context, store *entity.Store) error {
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"store_id": store.StoreID}
	update := bson.M{
		"$set": bson.M{
			"store_name":          store.StoreName,
			"owner_email":         store.OwnerEmail,
			"owner_phone":         store.OwnerPhone,
			"category":            store.Category,
			"plan_tier":           store.PlanTier,
			"subscription_status": store.SubscriptionStatus,
			"trial_ends_at":       store.TrialEndsAt,
			"expires_at":          store.ExpiresAt,
			"max_outlets":         store.MaxOutlets,
			"updated_at":          time.Now(),
		},
		"$setOnInsert": bson.M{
			"_id":        store.ID,
			"created_at": store.CreatedAt,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *StoreRepository) UpdateSubscription(ctx context.Context, storeID string, planTier string, status string, expiresAt *time.Time) error {
	filter := bson.M{"store_id": storeID}
	update := bson.M{
		"$set": bson.M{
			"plan_tier":           planTier,
			"subscription_status": status,
			"expires_at":          expiresAt,
			"updated_at":          time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *StoreRepository) UpdateStoreProfile(ctx context.Context, storeID string, storeName string, phone string, category string) error {
	filter := bson.M{"store_id": storeID}
	setFields := bson.M{
		"updated_at": time.Now(),
	}
	if storeName != "" {
		setFields["store_name"] = storeName
	}
	if phone != "" {
		setFields["owner_phone"] = phone
	}
	if category != "" {
		setFields["category"] = category
	}
	_, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": setFields})
	return err
}
