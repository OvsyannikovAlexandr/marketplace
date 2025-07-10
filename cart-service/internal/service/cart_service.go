package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/cache"
	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/repository"
)

type CartService struct {
	repo              repository.CartRepositoryInterface
	productServiceURL string
	cache             *cache.RedisCache
}

func NewCartService(repo repository.CartRepositoryInterface, productServiceURL string, cache *cache.RedisCache) *CartService {
	return &CartService{repo: repo, productServiceURL: productServiceURL, cache: cache}
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
	Checkout(ctx context.Context, userID int64) error
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
	err := s.repo.AddItem(ctx, item)
	if err != nil {
		return err
	}
	cacheKey := fmt.Sprintf("cart:user:%d", item.UserID)
	_ = s.cache.Delete(ctx, cacheKey)
	return nil
}

func (s *CartService) GetItems(ctx context.Context, userID int64) ([]domain.CartItem, error) {
	return s.repo.GetItemsByUserID(ctx, userID)
}

func (s *CartService) DeleteItem(ctx context.Context, userID, productID int64) error {
	err := s.repo.DeleteItem(ctx, userID, productID)
	if err != nil {
		return err
	}
	cacheKey := fmt.Sprintf("cart:user%d", userID)
	_ = s.cache.Delete(ctx, cacheKey)
	return nil
}

func (s *CartService) ClearCart(ctx context.Context, userID int64) error {
	err := s.repo.ClearCart(ctx, userID)
	if err != nil {
		return err
	}
	cacheKey := fmt.Sprintf("cart:user:%d", userID)
	_ = s.cache.Delete(ctx, cacheKey)
	return nil
}

func (s *CartService) GetCartWithDetails(ctx context.Context, userID int64) ([]domain.CartItemDetail, error) {
	cacheKey := fmt.Sprintf("cart:user:%d", userID)

	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		log.Println("Cache cart HIT:", cacheKey)
		var items []domain.CartItemDetail
		if err := json.Unmarshal([]byte(cached), &items); err == nil {
			return items, nil
		}
	}

	log.Println("Cache cart MISS:", cacheKey)

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

	data, _ := json.Marshal(detailedItems)
	_ = s.cache.Set(ctx, cacheKey, string(data), time.Minute*30)

	return detailedItems, nil
}

func (s *CartService) Checkout(ctx context.Context, userID int64) error {
	items, err := s.repo.GetItemsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return errors.New("cart is empty")
	}

	var productIDs []int64
	totalPrice := 0.0
	totalQuantity := 0

	for _, item := range items {
		product, err := s.getProductDetails(item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to fetch product %d: %w", item.ProductID, err)
		}
		totalPrice += product.Price * float64(item.Quantity)
		totalQuantity += item.Quantity
		productIDs = append(productIDs, item.ProductID)
	}

	order := map[string]interface{}{
		"user_id":     userID,
		"product_ids": productIDs,
		"quantity":    totalQuantity,
		"total_price": totalPrice,
		"status":      "new",
	}

	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
	resp, err := http.Post(orderServiceURL+"/orders", "application/json", encodeToJSON(order))
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("order-service returned status %d", resp.StatusCode)
	}

	_ = s.cache.Delete(ctx, fmt.Sprintf("cart:user:%d", userID))
	// Очистить корзину
	return s.repo.ClearCart(ctx, userID)
}

func encodeToJSON(v interface{}) *bytes.Buffer {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(v)
	return buf
}
