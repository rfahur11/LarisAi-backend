package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/larisai/pos-service/internal/config"
	"github.com/larisai/pos-service/internal/models/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProductRepository interface {
	FindAll(ctx context.Context, search string) ([]entity.Product, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*entity.Product, error)
	FindByBarcode(ctx context.Context, barcode string) (*entity.Product, error)
	Create(ctx context.Context, product *entity.Product) error
	UpdateStock(ctx context.Context, id primitive.ObjectID, quantityDelta int64) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error
}

type productRepo struct {
	col *mongo.Collection
}

func NewProductRepository() ProductRepository {
	return &productRepo{
		col: config.DB.Collection("products"),
	}
}

// FindAll mengambil daftar produk aktif (is_archived: false).
func (r *productRepo) FindAll(ctx context.Context, search string) ([]entity.Product, error) {
	filter := bson.M{"is_archived": false}
	if search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"barcode": bson.M{"$regex": search, "$options": "i"}},
			{"category": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []entity.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*entity.Product, error) {
	var product entity.Product
	err := r.col.FindOne(ctx, bson.M{"_id": id, "is_archived": false}).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

// FindByBarcode untuk fitur Scanner di kasir (<10ms lookup).
func (r *productRepo) FindByBarcode(ctx context.Context, barcode string) (*entity.Product, error) {
	var product entity.Product
	err := r.col.FindOne(ctx, bson.M{"barcode": barcode, "is_archived": false}).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *productRepo) Create(ctx context.Context, product *entity.Product) error {
	product.IsArchived = false
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	res, err := r.col.InsertOne(ctx, product)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		product.ID = oid
	}
	return nil
}

// UpdateStock mengurangi atau menambah stok secara atomik.
func (r *productRepo) UpdateStock(ctx context.Context, id primitive.ObjectID, quantityDelta int64) error {
	filter := bson.M{
		"_id":         id,
		"is_archived": false,
	}
	// Jika mengurangi stok, pastikan stok >= kuantitas yang diminta
	if quantityDelta < 0 {
		filter["stock"] = bson.M{"$gte": -quantityDelta}
	}

	update := bson.M{
		"$inc": bson.M{"stock": quantityDelta},
		"$set": bson.M{"updated_at": time.Now()},
	}

	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("stok produk tidak mencukupi atau produk tidak ditemukan/diarsipkan")
	}
	return nil
}

// SoftDelete mengarsipkan produk dengan mengatur is_archived = true.
func (r *productRepo) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id, "is_archived": false}
	update := bson.M{
		"$set": bson.M{
			"is_archived": true,
			"updated_at":  time.Now(),
		},
	}
	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("produk tidak ditemukan atau sudah diarsipkan")
	}
	return nil
}
