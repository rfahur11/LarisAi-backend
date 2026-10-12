package dto

import "time"

// CreateProductRequest adalah kontrak input dari client untuk membuat produk baru.
type CreateProductRequest struct {
	StoreID   string `json:"store_id,omitempty"`
	Barcode   string `json:"barcode"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Price     int64  `json:"price"`
	Stock     int64  `json:"stock"`
	ImagePath string `json:"image_path,omitempty"`
}

// ProductResponse adalah kontrak output ke client untuk data produk.
type ProductResponse struct {
	ID         string    `json:"id"`
	StoreID    string    `json:"store_id,omitempty"`
	Barcode    string    `json:"barcode"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	Price      int64     `json:"price"`
	Stock      int64     `json:"stock"`
	ImagePath  string    `json:"image_path,omitempty"`
	IsArchived bool      `json:"is_archived"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CheckoutItemRequest merepresentasikan item di dalam keranjang belanja saat checkout.
type CheckoutItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// CheckoutRequest adalah kontrak input untuk memproses transaksi kasir.
type CheckoutRequest struct {
	CustomerID  string                `json:"customer_id,omitempty"`
	PaymentType string                `json:"payment_type"` // TUNAI, QRIS, DEBIT
	Items       []CheckoutItemRequest `json:"items"`
}

// TransactionItemResponse merepresentasikan item pada struk/invoice transaksi.
type TransactionItemResponse struct {
	ProductID string `json:"product_id"`
	Barcode   string `json:"barcode"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
	Quantity  int    `json:"quantity"`
	Subtotal  int64  `json:"subtotal"`
}

// TransactionResponse adalah kontrak output untuk struk transaksi.
type TransactionResponse struct {
	ID          string                    `json:"id"`
	InvoiceNo   string                    `json:"invoice_no"`
	CustomerID  string                    `json:"customer_id,omitempty"`
	TotalAmount int64                     `json:"total_amount"`
	PaymentType string                    `json:"payment_type"`
	PaymentURL  string                    `json:"payment_url,omitempty"`
	Status      string                    `json:"status"`
	Items       []TransactionItemResponse `json:"items"`
	CreatedAt   time.Time                 `json:"created_at"`
}

// BulkCreateProductsRequest adalah kontrak input untuk batch insert produk massal.
type BulkCreateProductsRequest struct {
	Products []CreateProductRequest `json:"products"`
}

// BulkCreateProductsResponse adalah kontrak output ringkasan batch insert produk massal.
type BulkCreateProductsResponse struct {
	TotalReceived int      `json:"total_received"`
	TotalInserted int      `json:"total_inserted"`
	TotalSkipped  int      `json:"total_skipped"`
	SkippedErrors []string `json:"skipped_errors,omitempty"`
}
