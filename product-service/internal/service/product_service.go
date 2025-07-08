package service

import (
	"context"

	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/repository"
)

type ProdcutService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProdcutService {
	return &ProdcutService{repo: repo}
}

func (s *ProdcutService) Create(ctx context.Context, p domain.Product) error {
	return s.repo.CreateProduct(ctx, p)
}

func (s *ProdcutService) GetAll(ctx context.Context) ([]domain.Product, error) {
	return s.repo.GetAllProducts(ctx)
}

func (s *ProdcutService) GetByID(ctx context.Context, id int64) (domain.Product, error) {
	return s.repo.GetProdcutByID(ctx, id)
}

func (s *ProdcutService) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteProduct(ctx, id)
}
