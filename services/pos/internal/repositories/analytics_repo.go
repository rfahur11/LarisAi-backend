package repositories

import (
	"context"

	"github.com/larisai/pos-service/internal/config"
	"github.com/larisai/pos-service/internal/models/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AnalyticsRepository interface {
	GetValidTransactions(ctx context.Context) ([]entity.Transaction, error)
	GetActiveProducts(ctx context.Context) ([]entity.Product, error)
}

type analyticsRepo struct {
	txCol   *mongo.Collection
	prodCol *mongo.Collection
}

func NewAnalyticsRepository() AnalyticsRepository {
	return &analyticsRepo{
		txCol:   config.DB.Collection("transactions"),
		prodCol: config.DB.Collection("products"),
	}
}

func (r *analyticsRepo) GetValidTransactions(ctx context.Context) ([]entity.Transaction, error) {
	filter := bson.M{"is_archived": false}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := r.txCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var txs []entity.Transaction
	if err := cursor.All(ctx, &txs); err != nil {
		return nil, err
	}
	return txs, nil
}

func (r *analyticsRepo) GetActiveProducts(ctx context.Context) ([]entity.Product, error) {
	filter := bson.M{"is_archived": false}
	cursor, err := r.prodCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var prods []entity.Product
	if err := cursor.All(ctx, &prods); err != nil {
		return nil, err
	}
	return prods, nil
}
