package service

import (
	"context"
	"errors"

	"github.com/OvsyannikovAlexandr/marketplace/order-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/order-service/internal/repository"
)

type OrderServise struct {
	repo repository.OrderRepositoryInterface
}

type OrderServiceInterface interface {
	Create(ctx context.Context, order domain.Order) error
	GetByID(ctx context.Context, id int64) (domain.Order, error)
	GetAll(ctx context.Context) ([]domain.Order, error)
	Delete(ctx context.Context, id int64) error
}

func NewOrderService(repo repository.OrderRepositoryInterface) *OrderServise {
	return &OrderServise{repo: repo}
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
	return s.repo.CreateOrder(ctx, order)
}

func (s *OrderServise) GetAll(ctx context.Context) ([]domain.Order, error) {
	return s.repo.GetAllOrders(ctx)
}

func (s *OrderServise) GetByID(ctx context.Context, id int64) (domain.Order, error) {
	return s.repo.GetOrderByID(ctx, id)
}

func (s *OrderServise) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteOrder(ctx, id)
}
