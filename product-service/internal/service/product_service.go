package service

import (
	"context"

	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/repository"
)

type ProductService struct {
	repo repository.ProductRepositoryInterface
}

type ProductServiceInterface interface {
	Create(ctx context.Context, p domain.Product) error
	GetAll(ctx context.Context) ([]domain.Product, error)
	GetByID(ctx context.Context, id int64) (domain.Product, error)
	Delete(ctx context.Context, id int64) error
}

func NewProductService(repo repository.ProductRepositoryInterface) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, p domain.Product) error {
	return s.repo.CreateProduct(ctx, p)
}

func (s *ProductService) GetAll(ctx context.Context) ([]domain.Product, error) {
	return s.repo.GetAllProducts(ctx)
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (domain.Product, error) {
	return s.repo.GetProductByID(ctx, id)
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteProduct(ctx, id)
}
