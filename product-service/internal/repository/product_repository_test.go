//go:build !ci

package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var dbpool *pgxpool.Pool
var pgContainer testcontainers.Container

func TestMain(m *testing.M) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "products",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	var err error
	pgContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/products?sslmode=disable", host, port.Port())
	dbpool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}

	schema := `
	CREATE TABLE products (
    	id SERIAL PRIMARY KEY,
    	name TEXT NOT NULL,
    	description TEXT,
    	price NUMERIC(10,2) NOT NULL,
    	created_at TIMESTAMP NOT NULL DEFAULT now(),
    	updated_at TIMESTAMP NOT NULL DEFAULT now()
	)`
	_, err = dbpool.Exec(ctx, schema)
	if err != nil {
		panic(err)
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

	// Проверяем, что продукт больше не доступен
	_, err = repo.GetProductByID(ctx, id)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}
