package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/models/entity"
	"github.com/larisai/pos-service/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mockProductRepo implements repositories.ProductRepository for testing
type mockProductRepo struct {
	products map[string]*entity.Product
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{
		products: make(map[string]*entity.Product),
	}
}

func (m *mockProductRepo) FindAll(ctx context.Context, search string) ([]entity.Product, error) {
	return nil, nil
}

func (m *mockProductRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*entity.Product, error) {
	if p, ok := m.products[id.Hex()]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockProductRepo) FindByBarcode(ctx context.Context, barcode string) (*entity.Product, error) {
	for _, p := range m.products {
		if p.Barcode == barcode {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockProductRepo) Create(ctx context.Context, product *entity.Product) error {
	product.ID = primitive.NewObjectID()
	m.products[product.ID.Hex()] = product
	return nil
}

func (m *mockProductRepo) Update(ctx context.Context, id primitive.ObjectID, product *entity.Product) error {
	if _, ok := m.products[id.Hex()]; !ok {
		return errors.New("not found")
	}
	// simulate update
	existing := m.products[id.Hex()]
	existing.Name = product.Name
	existing.Barcode = product.Barcode
	existing.Price = product.Price
	existing.Stock = product.Stock
	return nil
}

func (m *mockProductRepo) UpdateStock(ctx context.Context, id primitive.ObjectID, quantityDelta int64) error {
	return nil
}

func (m *mockProductRepo) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	if _, ok := m.products[id.Hex()]; !ok {
		return errors.New("not found")
	}
	delete(m.products, id.Hex())
	return nil
}

func TestProductService_UpdateProduct_Success(t *testing.T) {
	repo := newMockProductRepo()
	svc := services.NewProductService(repo)
	ctx := context.Background()

	// Seed data
	id := primitive.NewObjectID()
	repo.products[id.Hex()] = &entity.Product{
		ID:        id,
		Name:      "Beras 5kg",
		Barcode:   "12345",
		Price:     60000,
		Stock:     10,
		CreatedAt: time.Now(),
	}

	req := dto.CreateProductRequest{
		Barcode:  "12345",
		Name:     "Beras Premium 5kg",
		Category: "Sembako",
		Price:    65000,
		Stock:    15,
	}

	res, err := svc.UpdateProduct(ctx, id.Hex(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if res.Name != "Beras Premium 5kg" {
		t.Errorf("Expected Name 'Beras Premium 5kg', got %s", res.Name)
	}
	if res.Price != 65000 {
		t.Errorf("Expected Price 65000, got %d", res.Price)
	}
	if res.Stock != 15 {
		t.Errorf("Expected Stock 15, got %d", res.Stock)
	}
}

func TestProductService_UpdateProduct_InvalidInput(t *testing.T) {
	repo := newMockProductRepo()
	svc := services.NewProductService(repo)
	ctx := context.Background()

	req := dto.CreateProductRequest{
		Barcode: "",
		Name:    "Invalid Product",
		Price:   -100,
	}

	_, err := svc.UpdateProduct(ctx, primitive.NewObjectID().Hex(), req)
	if err == nil {
		t.Fatal("Expected error for invalid input, got nil")
	}
}

func TestProductService_UpdateProduct_DuplicateBarcode(t *testing.T) {
	repo := newMockProductRepo()
	svc := services.NewProductService(repo)
	ctx := context.Background()

	// Seed 2 products
	id1 := primitive.NewObjectID()
	repo.products[id1.Hex()] = &entity.Product{ID: id1, Name: "P1", Barcode: "111", Price: 100}
	
	id2 := primitive.NewObjectID()
	repo.products[id2.Hex()] = &entity.Product{ID: id2, Name: "P2", Barcode: "222", Price: 200}

	// Try to update P1 to have barcode of P2
	req := dto.CreateProductRequest{
		Barcode: "222",
		Name:    "P1 Updated",
		Price:   150,
	}

	_, err := svc.UpdateProduct(ctx, id1.Hex(), req)
	if err == nil {
		t.Fatal("Expected error for duplicate barcode, got nil")
	}
}

func TestProductService_DeleteProduct_Success(t *testing.T) {
	repo := newMockProductRepo()
	svc := services.NewProductService(repo)
	ctx := context.Background()

	id := primitive.NewObjectID()
	repo.products[id.Hex()] = &entity.Product{ID: id, Name: "To Be Deleted", Barcode: "999", Price: 100}

	err := svc.DeleteProduct(ctx, id.Hex())
	if err != nil {
		t.Fatalf("Expected no error on delete, got: %v", err)
	}

	if len(repo.products) != 0 {
		t.Errorf("Expected product to be deleted from repo")
	}
}
