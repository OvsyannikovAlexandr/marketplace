package service

import (
	"context"
	"errors"

	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/repository"
)

type CartService struct {
	repo repository.CartRepositoryInterface
}

func NewCartService(repo repository.CartRepositoryInterface) *CartService {
	return &CartService{repo: repo}
}

type CartServiceInterface interface {
	AddItem(ctx context.Context, item domain.CartItem) error
	GetItems(ctx context.Context, userID int64) ([]domain.CartItem, error)
	DeleteItem(ctx context.Context, userID, productID int64) error
	ClearCart(ctx context.Context, userID int64) error
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
	return s.repo.GetItemsByUser(ctx, userID)
}

func (s *CartService) DeleteItem(ctx context.Context, userID, productID int64) error {
	return s.repo.DeleteItem(ctx, userID, productID)
}

func (s *CartService) ClearCart(ctx context.Context, userID int64) error {
	return s.repo.ClearCart(ctx, userID)
}
