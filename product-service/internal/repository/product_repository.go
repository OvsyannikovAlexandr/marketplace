package repository

import (
	"context"
	"errors"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, p domain.Product) error {
	query := `
		INSERT INTO product (name, description, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, p.Name, p.Description, p.Price, p.CreatedAt, time.Now(), time.Now())

	return err
}

func (r *ProductRepository) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
	query := `SELECT id, name, description, price, created_at, updated_at FROM products`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) GetProdcutByID(ctx context.Context, id int64) (domain.Product, error) {
	query := `SELECT id, name, description, price, created_at, updated_at FROM products WHERE id = $1`
	var p domain.Product
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return p, errors.New("product not fuond")
	}
	return p, nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}
