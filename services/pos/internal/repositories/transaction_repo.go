package repositories

import (
	"context"
	"time"

	"github.com/larisai/pos-service/internal/config"
	"github.com/larisai/pos-service/internal/models/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction) error
	FindAll(ctx context.Context) ([]entity.Transaction, error)
	UpdateStatus(ctx context.Context, invoiceNo string, status string) error
}

type transactionRepo struct {
	col *mongo.Collection
}

func NewTransactionRepository() TransactionRepository {
	return &transactionRepo{
		col: config.DB.Collection("transactions"),
	}
}

func (r *transactionRepo) Create(ctx context.Context, tx *entity.Transaction) error {
	tx.IsArchived = false
	tx.CreatedAt = time.Now()

	res, err := r.col.InsertOne(ctx, tx)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		tx.ID = oid
	}
	return nil
}

func (r *transactionRepo) FindAll(ctx context.Context) ([]entity.Transaction, error) {
	filter := bson.M{"is_archived": false}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.col.Find(ctx, filter, opts)
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

func (r *transactionRepo) UpdateStatus(ctx context.Context, invoiceNo string, status string) error {
	filter := bson.M{"invoice_no": invoiceNo}
	update := bson.M{"$set": bson.M{"status": status}}
	
	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
