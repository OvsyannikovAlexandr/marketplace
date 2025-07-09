package repository

import (
	"context"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CartRepository struct {
	db *pgxpool.Pool
}

func NewCartRepository(db *pgxpool.Pool) *CartRepository {
	return &CartRepository{db: db}
}

type CartRepositoryInterface interface {
	AddItem(ctx context.Context, item domain.CartItem) error
	GetItemsByUser(ctx context.Context, userID int64) ([]domain.CartItem, error)
	DeleteItem(ctx context.Context, userID, productID int64) error
	ClearCart(ctx context.Context, userID int64) error
}

func (r *CartRepository) AddItem(ctx context.Context, item domain.CartItem) error {
	query := `
		INSERT INTO cart_service.cart_items (user_id, product_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, product_id) DO UPDATE
		SET quantity = cart_items.quantity + EXCLUDED.quantity,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(ctx, query, item.UserID, item.ProductID, item.Quantity, time.Now(), time.Now())
	return err
}

func (r *CartRepository) GetItemsByUser(ctx context.Context, userID int64) ([]domain.CartItem, error) {
	query := `
		SELECT id, user_id, product_id, quantity, created_at, updated_at
		FROM cart_service.cart_items
		WHERE user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.CartItem
	for rows.Next() {
		var item domain.CartItem
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ProductID,
			&item.Quantity,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *CartRepository) DeleteItem(ctx context.Context, userID, productID int64) error {
	query := `DELETE FROM cart_service.cart_items WHERE user_id=$1 AND product_id=$2`
	_, err := r.db.Exec(ctx, query, userID, productID)
	return err
}

func (r *CartRepository) ClearCart(ctx context.Context, userID int64) error {
	query := `DELETE FROM cart_service.cart_items WHERE user_id=$1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
