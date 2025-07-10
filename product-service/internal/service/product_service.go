package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/cache"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/repository"
)

type ProductService struct {
	repo  repository.ProductRepositoryInterface
	cache *cache.RedisCache
}

type ProductServiceInterface interface {
	Create(ctx context.Context, p domain.Product) error
	GetAll(ctx context.Context) ([]domain.Product, error)
	GetByID(ctx context.Context, id int64) (domain.Product, error)
	Delete(ctx context.Context, id int64) error
}

func NewProductService(repo repository.ProductRepositoryInterface, cache *cache.RedisCache) *ProductService {
	return &ProductService{repo: repo, cache: cache}
}

func (s *ProductService) Create(ctx context.Context, p domain.Product) error {
	return s.repo.CreateProduct(ctx, p)
}

func (s *ProductService) GetAll(ctx context.Context) ([]domain.Product, error) {
	return s.repo.GetAllProducts(ctx)
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (domain.Product, error) {
	cacheKey := fmt.Sprintf("product:%d", id)

	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		log.Println("Cache HIT:", cacheKey)
		var product domain.Product
		if err := json.Unmarshal([]byte(cached), &product); err == nil {
			return product, nil
		}
	}

	log.Println("Cache MISS:", cacheKey)

	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}

	data, _ := json.Marshal(product)
	_ = s.cache.Set(ctx, cacheKey, string(data), time.Minute*5)

	return product, nil
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	err := s.repo.DeleteProduct(ctx, id)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("product:%d", id)
	_ = s.cache.Delete(ctx, cacheKey)

	return nil
}
