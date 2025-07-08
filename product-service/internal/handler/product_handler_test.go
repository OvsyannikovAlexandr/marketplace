package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/handler"
	"github.com/gorilla/mux"
)

type mockService struct {
	products map[int64]domain.Product
	nextID   int64
}

func (m *mockService) Create(ctx context.Context, p domain.Product) error {
	m.nextID++
	p.ID = m.nextID
	m.products[p.ID] = p
	return nil
}

func (m *mockService) GetAll(ctx context.Context) ([]domain.Product, error) {
	var result []domain.Product
	for _, p := range m.products {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockService) GetByID(ctx context.Context, id int64) (domain.Product, error) {
	if p, ok := m.products[id]; ok {
		return p, nil
	}
	return domain.Product{}, errors.New("product not found")
}

func (m *mockService) Delete(ctx context.Context, id int64) error {
	if _, ok := m.products[id]; ok {
		delete(m.products, id)
		return nil
	}
	return errors.New("product not found")
}

func TestCreateProductHandler(t *testing.T) {
	s := &mockService{products: make(map[int64]domain.Product)}
	h := handler.NewProductHandler(s)

	body := `{"name":"Test","description":"Demo","price":99.99}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestGetAllProductsHandler(t *testing.T) {
	s := &mockService{products: make(map[int64]domain.Product)}
	_ = s.Create(context.Background(), domain.Product{Name: "A"})
	_ = s.Create(context.Background(), domain.Product{Name: "B"})

	h := handler.NewProductHandler(s)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()

	h.GetAll(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var products []domain.Product
	if err := json.NewDecoder(rec.Body).Decode(&products); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
}

func TestGetByIDProductHandler(t *testing.T) {
	s := &mockService{products: make(map[int64]domain.Product)}
	_ = s.Create(context.Background(), domain.Product{Name: "Item 1"})

	h := handler.NewProductHandler(s)

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDeleteProductHandler(t *testing.T) {
	s := &mockService{products: make(map[int64]domain.Product)}
	_ = s.Create(context.Background(), domain.Product{Name: "ToDelete"})

	h := handler.NewProductHandler(s)

	req := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestGetByIDProductHandler_NotFound(t *testing.T) {
	s := &mockService{products: make(map[int64]domain.Product)}
	h := handler.NewProductHandler(s)

	req := httptest.NewRequest(http.MethodGet, "/products/99", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "99"})
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "not found") {
		t.Fatalf("expected not found error, got %s", rec.Body.String())
	}
}

func TestCreateProductHandler_InvalidBody(t *testing.T) {
	s := &mockService{products: make(map[int64]domain.Product)}
	h := handler.NewProductHandler(s)

	body := `{"name": "Test"` // некорректный JSON
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid body") {
		t.Fatalf("expected body error, got %s", rec.Body.String())
	}
}

func TestDeleteProductHandler_InvalidID(t *testing.T) {
	s := &mockService{products: make(map[int64]domain.Product)}
	h := handler.NewProductHandler(s)

	req := httptest.NewRequest(http.MethodDelete, "/products/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"}) // некорректный ID
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ID, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid ID") {
		t.Fatalf("expected invalid ID error, got %s", rec.Body.String())
	}
}
