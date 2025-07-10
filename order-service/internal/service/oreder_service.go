package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/order-service/internal/cache"
	"github.com/OvsyannikovAlexandr/marketplace/order-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/order-service/internal/repository"
	"github.com/OvsyannikovAlexandr/marketplace/order-service/pkg/kafka"
)

type OrderServise struct {
	repo     repository.OrderRepositoryInterface
	producer kafka.Producer
	cache    *cache.RedisCache
}

type OrderServiceInterface interface {
	Create(ctx context.Context, order domain.Order) error
	GetByID(ctx context.Context, id int64) (domain.Order, error)
	GetAll(ctx context.Context) ([]domain.Order, error)
	Delete(ctx context.Context, id int64) error
}

func NewOrderService(repo repository.OrderRepositoryInterface, producer kafka.Producer, cache *cache.RedisCache) *OrderServise {
	return &OrderServise{repo: repo, producer: producer, cache: cache}
}

func (s *OrderServise) Create(ctx context.Context, order domain.Order) error {
	if order.UserID == 0 {
		return errors.New("user id must be set")
	}
	if len(order.ProductIDs) == 0 {
		return errors.New("product IDs can't be empty")
	}
	if order.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if order.TotalPrice < 0 {
		return errors.New("total price can't be negative")
	}
	if order.Status == "" {
		order.Status = "new"
	}

	err := s.repo.CreateOrder(ctx, &order)
	if err != nil {
		return err
	}

	_ = s.producer.SendOrderCreated(ctx, kafka.OrderCreatedEvent{
		OrderID:    order.ID,
		UserID:     order.UserID,
		ProductIDs: order.ProductIDs,
		Quantity:   order.Quantity,
		TotalPrice: order.TotalPrice,
		CreatedAt:  time.Now(),
	})

	cacheKey := fmt.Sprintf("order:%d", order.ID)
	_ = s.cache.Delete(ctx, cacheKey)

	return nil
}

func (s *OrderServise) GetAll(ctx context.Context) ([]domain.Order, error) {
	return s.repo.GetAllOrders(ctx)
}

func (s *OrderServise) GetByID(ctx context.Context, id int64) (domain.Order, error) {
	cacheKey := fmt.Sprintf("order:%d", id)

	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var order domain.Order
		log.Println("Cache order HIT: ", cacheKey)
		if err := json.Unmarshal([]byte(cached), &order); err == nil {
			return order, nil
		}
	}

	log.Println("Cache order MISS: ", cacheKey)
	order, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	data, _ := json.Marshal(order)
	_ = s.cache.Set(ctx, cacheKey, string(data), time.Minute*10)

	return order, nil
}

func (s *OrderServise) Delete(ctx context.Context, id int64) error {
	err := s.repo.DeleteOrder(ctx, id)
	if err != nil {
		return err
	}
	cacheKey := fmt.Sprintf("order:%d", id)
	_ = s.cache.Delete(ctx, cacheKey)

	return nil
}
