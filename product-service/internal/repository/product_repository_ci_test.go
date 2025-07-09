//go:build ci

package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
)

var dbpool *pgxpool.Pool
var pgContainer testcontainers.Container

func TestMain(m *testing.M) {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback для GitHub Actions
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")

		dsn = "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=disable"
	}

	var err error
	dbpool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	_, err = dbpool.Exec(ctx, `DELETE FROM products`)
	if err != nil {
		panic("failed to clear table: " + err.Error())
	}

	code := m.Run()

	dbpool.Close()
	pgContainer.Terminate(ctx)
	os.Exit(code)
}

func clearProductsTable(t *testing.T) {
	_, err := dbpool.Exec(context.Background(), "DELETE FROM products")
	if err != nil {
		t.Fatalf("failed to clear users table: %v", err)
	}
}

func TestCreateAndGetALLProduct(t *testing.T) {
	clearProductsTable(t)
	ctx := context.Background()
	repo := repository.NewProductRepository(dbpool)

	product := domain.Product{
		Name:        "Test Product",
		Description: "Cool product",
		Price:       99.99,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := repo.CreateProduct(ctx, product); err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}

	products, err := repo.GetAllProducts(ctx)
	if err != nil {
		t.Fatalf("GetAllProducts failed: %v", err)
	}

	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}

	got := products[0]
	if got.Name != product.Name {
		t.Errorf("expected name %s, got %s", product.Name, got.Name)
	}
	if got.Description != product.Description {
		t.Errorf("expected description %s, got %s", product.Description, got.Description)
	}
}

func TestGetProductByID_NotFound(t *testing.T) {
	clearProductsTable(t)
	ctx := context.Background()
	repo := repository.NewProductRepository(dbpool)

	_, err := repo.GetProductByID(ctx, 9999999)
	if err == nil {
		t.Fatalf("expected error not dound product, got nil")
	}
}

func TestDeleteProduct_Success(t *testing.T) {
	clearProductsTable(t)
	ctx := context.Background()
	repo := repository.NewProductRepository(dbpool)

	product := domain.Product{
		Name:        "Test to delete Product",
		Description: "To be deleted product",
		Price:       99.99,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateProduct(ctx, product)
	if err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}

	products, err := repo.GetAllProducts(ctx)
	if err != nil {
		t.Fatalf("GetAllProducts failed: %v", err)
	}

	var id int64 = -1
	for _, p := range products {
		if p.Name == "Test to delete Product" {
			id = p.ID
			break
		}
	}
	if id == -1 {
		t.Fatal("Inserted product not found in list")
	}

	err = repo.DeleteProduct(ctx, id)
	if err != nil {
		t.Fatalf("DeleteProduct failed: %v", err)
	}

	_, err = repo.GetProductByID(ctx, id)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}
