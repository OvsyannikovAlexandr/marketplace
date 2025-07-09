package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/repository"
)

type CartService struct {
	repo              repository.CartRepositoryInterface
	productServiceURL string
}

func NewCartService(repo repository.CartRepositoryInterface, productServiceURL string) *CartService {
	return &CartService{repo: repo, productServiceURL: productServiceURL}
}

func (s *CartService) getProductDetails(productID int64) (domain.Product, error) {
	url := fmt.Sprintf("%s/products/%d", s.productServiceURL, productID)

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return domain.Product{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.Product{}, fmt.Errorf("product-service returned status %d", resp.StatusCode)
	}

	var product domain.Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return domain.Product{}, err
	}

	return product, nil
}

type CartServiceInterface interface {
	AddItem(ctx context.Context, item domain.CartItem) error
	GetItems(ctx context.Context, userID int64) ([]domain.CartItem, error)
	DeleteItem(ctx context.Context, userID, productID int64) error
	ClearCart(ctx context.Context, userID int64) error
	GetCartWithDetails(ctx context.Context, userID int64) ([]domain.CartItemDetail, error)
}

func (s *CartService) AddItem(ctx context.Context, item domain.CartItem) error {
	if item.UserID == 0 {
		return errors.New("user id must be set")
	}
	if item.ProductID == 0 {
		return errors.New("product id must be set")
	}
	if item.Quantity <= 0 {
		return errors.New("quantity must be possible")
	}

	return s.repo.AddItem(ctx, item)
}

func (s *CartService) GetItems(ctx context.Context, userID int64) ([]domain.CartItem, error) {
	return s.repo.GetItemsByUserID(ctx, userID)
}

func (s *CartService) DeleteItem(ctx context.Context, userID, productID int64) error {
	return s.repo.DeleteItem(ctx, userID, productID)
}

func (s *CartService) ClearCart(ctx context.Context, userID int64) error {
	return s.repo.ClearCart(ctx, userID)
}

func (s *CartService) GetCartWithDetails(ctx context.Context, userID int64) ([]domain.CartItemDetail, error) {
	items, err := s.repo.GetItemsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var detailedItems []domain.CartItemDetail
	for _, item := range items {
		product, err := s.getProductDetails(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product details for product %d: %w", item.ProductID, err)
		}
		detailedItems = append(detailedItems, domain.CartItemDetail{
			Product:  product,
			Quantity: item.Quantity,
		})
	}

	return detailedItems, nil
}
