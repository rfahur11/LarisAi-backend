package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/models/entity"
	"github.com/larisai/pos-service/internal/repositories"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CheckoutService interface {
	ProcessCheckout(ctx context.Context, req dto.CheckoutRequest) (*dto.TransactionResponse, error)
	GetTransactions(ctx context.Context) ([]dto.TransactionResponse, error)
}

type checkoutSvc struct {
	productRepo repositories.ProductRepository
	txRepo      repositories.TransactionRepository
}

func NewCheckoutService(productRepo repositories.ProductRepository, txRepo repositories.TransactionRepository) CheckoutService {
	return &checkoutSvc{
		productRepo: productRepo,
		txRepo:      txRepo,
	}
}

func (s *checkoutSvc) ProcessCheckout(ctx context.Context, req dto.CheckoutRequest) (*dto.TransactionResponse, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("keranjang belanja tidak boleh kosong")
	}
	if req.PaymentType == "" {
		req.PaymentType = "TUNAI"
	}

	var totalAmount int64
	var txItems []entity.TransactionItem
	var respItems []dto.TransactionItemResponse

	// Validasi & kurangi stok untuk setiap item
	for _, itemReq := range req.Items {
		if itemReq.Quantity <= 0 {
			return nil, errors.New("kuantitas barang harus lebih dari 0")
		}
		oid, err := primitive.ObjectIDFromHex(itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("ID produk tidak valid: %s", itemReq.ProductID)
		}

		product, err := s.productRepo.FindByID(ctx, oid)
		if err != nil || product == nil {
			return nil, fmt.Errorf("produk dengan ID %s tidak ditemukan", itemReq.ProductID)
		}

		if product.Stock < int64(itemReq.Quantity) {
			return nil, fmt.Errorf("stok produk %s tidak mencukupi (tersisa: %d)", product.Name, product.Stock)
		}

		// Kurangi stok secara atomik
		err = s.productRepo.UpdateStock(ctx, oid, -int64(itemReq.Quantity))
		if err != nil {
			return nil, fmt.Errorf("gagal mengurangi stok untuk produk %s: %v", product.Name, err)
		}

		subtotal := product.Price * int64(itemReq.Quantity)
		totalAmount += subtotal

		txItems = append(txItems, entity.TransactionItem{
			ProductID: oid,
			Barcode:   product.Barcode,
			Name:      product.Name,
			Price:     product.Price,
			Quantity:  itemReq.Quantity,
			Subtotal:  subtotal,
		})

		respItems = append(respItems, dto.TransactionItemResponse{
			ProductID: oid.Hex(),
			Barcode:   product.Barcode,
			Name:      product.Name,
			Price:     product.Price,
			Quantity:  itemReq.Quantity,
			Subtotal:  subtotal,
		})
	}

	// Generate Invoice Number unik
	invoiceNo := fmt.Sprintf("INV-%s-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)

	var custID *primitive.ObjectID
	if req.CustomerID != "" {
		if oid, err := primitive.ObjectIDFromHex(req.CustomerID); err == nil {
			custID = &oid
		}
	}

	status := "paid" // Default untuk tunai
	var paymentUrl string
	if req.PaymentType == "QRIS" || req.PaymentType == "qris" {
		status = "pending"
		// Coba minta Snap URL dari Midtrans
		serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
		if serverKey != "" {
			midtrans.ServerKey = serverKey
			midtrans.Environment = midtrans.Sandbox

			reqSnap := &snap.Request{
				TransactionDetails: midtrans.TransactionDetails{
					OrderID:  invoiceNo,
					GrossAmt: totalAmount,
				},
				CreditCard: &snap.CreditCardDetails{
					Secure: true,
				},
			}
			snapResp, _ := snap.CreateTransaction(reqSnap)
			if snapResp != nil {
				paymentUrl = snapResp.RedirectURL
			}
		}
	}

	tx := &entity.Transaction{
		InvoiceNo:   invoiceNo,
		CustomerID:  custID,
		TotalAmount: totalAmount,
		PaymentType: req.PaymentType,
		PaymentURL:  paymentUrl,
		Status:      status,
		Items:       txItems,
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("gagal mencatat transaksi kasir: %v", err)
	}

	return &dto.TransactionResponse{
		ID:          tx.ID.Hex(),
		InvoiceNo:   tx.InvoiceNo,
		CustomerID:  req.CustomerID,
		TotalAmount: tx.TotalAmount,
		PaymentType: tx.PaymentType,
		PaymentURL:  tx.PaymentURL,
		Status:      tx.Status,
		Items:       respItems,
		CreatedAt:   tx.CreatedAt,
	}, nil
}

func (s *checkoutSvc) GetTransactions(ctx context.Context) ([]dto.TransactionResponse, error) {
	txs, err := s.txRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]dto.TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		var custIDStr string
		if tx.CustomerID != nil {
			custIDStr = tx.CustomerID.Hex()
		}
		var respItems []dto.TransactionItemResponse
		for _, item := range tx.Items {
			respItems = append(respItems, dto.TransactionItemResponse{
				ProductID: item.ProductID.Hex(),
				Barcode:   item.Barcode,
				Name:      item.Name,
				Price:     item.Price,
				Quantity:  item.Quantity,
				Subtotal:  item.Subtotal,
			})
		}
		res = append(res, dto.TransactionResponse{
			ID:          tx.ID.Hex(),
			InvoiceNo:   tx.InvoiceNo,
			CustomerID:  custIDStr,
			TotalAmount: tx.TotalAmount,
			PaymentType: tx.PaymentType,
			PaymentURL:  tx.PaymentURL,
			Status:      tx.Status,
			Items:       respItems,
			CreatedAt:   tx.CreatedAt,
		})
	}
	return res, nil
}
