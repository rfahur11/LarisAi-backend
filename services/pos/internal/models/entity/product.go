package entity

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product merepresentasikan skema koleksi barang di MongoDB dengan Soft Delete (IsArchived).
type Product struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Barcode    string             `bson:"barcode" json:"barcode"`
	Name       string             `bson:"name" json:"name"`
	Category   string             `bson:"category" json:"category"`
	Price      int64              `bson:"price" json:"price"`
	Stock      int64              `bson:"stock" json:"stock"`
	IsArchived bool               `bson:"is_archived" json:"is_archived"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}
