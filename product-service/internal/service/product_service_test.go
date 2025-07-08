package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/service"
)

type mockProductRepo struct {
	products map[int64]domain.Product
	nextID   int64
}

func (m *mockProductRepo) CreateProduct(ctx context.Context, p domain.Product) error {
	m.nextID++
	p.ID = m.nextID
	m.products[p.ID] = p
	return nil
}

func (m *mockProductRepo) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
	var result []domain.Product
	for _, p := range m.products {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockProductRepo) GetProductByID(ctx context.Context, id int64) (domain.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return domain.Product{}, errors.New("product not found")
	}

	return p, nil
}

func (m *mockProductRepo) DeleteProduct(ctx context.Context, id int64) error {
	if _, ok := m.products[id]; !ok {
		return errors.New("product not found")
	}
	delete(m.products, id)
	return nil
}

func setupService() (*service.ProductService, *mockProductRepo) {
	mock := &mockProductRepo{
		products: make(map[int64]domain.Product),
	}
	svc := service.NewProductService(mock)
	return svc, mock
}

func TestCreateProduct(t *testing.T) {
	svc, _ := setupService()

	p := domain.Product{
		Name:        "Test",
		Description: "Test description",
		Price:       10.5,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := svc.Create(context.Background(), p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGetAllProducts(t *testing.T) {
	svc, mock := setupService()

	mock.CreateProduct(context.Background(), domain.Product{Name: "A"})
	mock.CreateProduct(context.Background(), domain.Product{Name: "B"})

	products, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
}

func TestGetByID(t *testing.T) {
	svc, mock := setupService()

	_ = mock.CreateProduct(context.Background(), domain.Product{Name: "GetMe"})
	var id int64 = 1

	product, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if product.Name != "GetMe" {
		t.Errorf("expected 'GetMe', got %s", product.Name)
	}
}

func TestDeleteProduct(t *testing.T) {
	svc, mock := setupService()

	_ = mock.CreateProduct(context.Background(), domain.Product{Name: "DeleteMe"})
	var id int64 = 1

	err := svc.Delete(context.Background(), id)
	if err != nil {
		t.Fatalf("expected no error on delete, got %v", err)
	}

	_, err = svc.GetByID(context.Background(), id)
	if err == nil {
		t.Fatalf("expected error after delete, got nil")
	}
}
