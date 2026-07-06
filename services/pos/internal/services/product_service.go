package services

import (
	"context"
	"errors"

	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/models/entity"
	"github.com/larisai/pos-service/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductService interface {
	GetProducts(ctx context.Context, search string) ([]dto.ProductResponse, error)
	GetProductByBarcode(ctx context.Context, barcode string) (*dto.ProductResponse, error)
	CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, idStr string) error
}

type productSvc struct {
	repo repositories.ProductRepository
}

func NewProductService(repo repositories.ProductRepository) ProductService {
	return &productSvc{repo: repo}
}

func (s *productSvc) GetProducts(ctx context.Context, search string) ([]dto.ProductResponse, error) {
	products, err := s.repo.FindAll(ctx, search)
	if err != nil {
		return nil, err
	}

	res := make([]dto.ProductResponse, 0, len(products))
	for _, p := range products {
		res = append(res, toProductResponse(p))
	}
	return res, nil
}

func (s *productSvc) GetProductByBarcode(ctx context.Context, barcode string) (*dto.ProductResponse, error) {
	if barcode == "" {
		return nil, errors.New("kode barcode tidak boleh kosong")
	}
	product, err := s.repo.FindByBarcode(ctx, barcode)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("produk dengan barcode tersebut tidak ditemukan atau sudah diarsipkan")
	}
	resp := toProductResponse(*product)
	return &resp, nil
}

func (s *productSvc) CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	if req.Barcode == "" || req.Name == "" || req.Price <= 0 {
		return nil, errors.New("barcode, nama produk, dan harga valid wajib diisi")
	}

	// Cek apakah barcode sudah ada
	existing, _ := s.repo.FindByBarcode(ctx, req.Barcode)
	if existing != nil {
		return nil, errors.New("produk dengan barcode tersebut sudah ada")
	}

	product := &entity.Product{
		Barcode:  req.Barcode,
		Name:     req.Name,
		Category: req.Category,
		Price:    req.Price,
		Stock:    req.Stock,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	resp := toProductResponse(*product)
	return &resp, nil
}

func (s *productSvc) DeleteProduct(ctx context.Context, idStr string) error {
	oid, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return errors.New("format ID produk tidak valid")
	}
	return s.repo.SoftDelete(ctx, oid)
}

func toProductResponse(p entity.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:         p.ID.Hex(),
		Barcode:    p.Barcode,
		Name:       p.Name,
		Category:   p.Category,
		Price:      p.Price,
		Stock:      p.Stock,
		IsArchived: p.IsArchived,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}
