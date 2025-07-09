package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/OvsyannikovAlexandr/marketplace/order-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

type OrderRepositoryInterface interface {
	CreateOrder(ctx context.Context, order domain.Order) error
	GetOrderByID(ctx context.Context, id int64) (domain.Order, error)
	GetAllOrders(ctx context.Context) ([]domain.Order, error)
	DeleteOrder(ctx context.Context, id int64) error
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order domain.Order) error {
	query := `
		INSERT INTO order_service.orders (user_id, product_ids, quantity, total_price, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`
	productIDs := fmt.Sprintf("{%s}", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(order.ProductIDs)), ","), "[]"))

	err := r.db.QueryRow(ctx, query, order.UserID, productIDs, order.Quantity, order.TotalPrice, order.Status).Scan(&order.ID)

	return err
}

func (r *OrderRepository) GetAllOrders(ctx context.Context) ([]domain.Order, error) {
	query := `SELECT id, user_id, product_ids, quantity, total_price, status, created_at, updated_at FROM order_service.orders`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		var productIDs []int64

		err := rows.Scan(
			&o.ID,
			&o.UserID,
			&productIDs,
			&o.Quantity,
			&o.TotalPrice,
			&o.Status,
			&o.CreatedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		o.ProductIDs = productIDs
		orders = append(orders, o)
	}

	return orders, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id int64) (domain.Order, error) {
	query := `
		SELECT id, user_id, product_ids, quantity, total_price, status, created_at, updated_at 
		FROM order_service.orders
		WHERE id = $1
	`
	var o domain.Order
	var productIDs []int64

	err := r.db.QueryRow(ctx, query, id).Scan(
		&o.ID,
		&o.UserID,
		&productIDs,
		&o.Quantity,
		&o.TotalPrice,
		&o.Status,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		return domain.Order{}, err
	}
	o.ProductIDs = productIDs
	return o, nil
}

func (r *OrderRepository) DeleteOrder(ctx context.Context, id int64) error {
	query := `DELETE FROM order_service.orders WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
