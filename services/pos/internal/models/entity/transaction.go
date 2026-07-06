package entity

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TransactionItem merepresentasikan item produk yang dibeli di dalam transaksi kasir.
type TransactionItem struct {
	ProductID primitive.ObjectID `bson:"product_id" json:"product_id"`
	Barcode   string             `bson:"barcode" json:"barcode"`
	Name      string             `bson:"name" json:"name"`
	Price     int64              `bson:"price" json:"price"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Subtotal  int64              `bson:"subtotal" json:"subtotal"`
}

// Transaction merepresentasikan invoice kasir yang tercatat di MongoDB.
type Transaction struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	InvoiceNo   string              `bson:"invoice_no" json:"invoice_no"`
	CustomerID  *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty"`
	TotalAmount int64               `bson:"total_amount" json:"total_amount"`
	PaymentType string              `bson:"payment_type" json:"payment_type"`
	Items       []TransactionItem   `bson:"items" json:"items"`
	IsArchived  bool                `bson:"is_archived" json:"is_archived"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
}
