package repository_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	pgContainer testcontainers.Container
	dbpool      *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "users",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		Tmpfs:      map[string]string{"/var/lib/postgresql/data": "rw"},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}
	var err error
	pgContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		log.Fatalf("failed to get host: %v", err)
	}
	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		log.Fatalf("failed to get port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/users?sslmode=disable", host, mappedPort.Port())
	dbpool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	// Создаём таблицу
	schema := `
	CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err = dbpool.Exec(ctx, schema)
	if err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}

	// Выполняем тесты
	code := m.Run()

	// Чистим ресурсы
	dbpool.Close()
	if err := pgContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}

	os.Exit(code)
}

func TestCreateAndGetUser(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewUserRepository(dbpool)

	user := domain.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hash123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	got, err := repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}

	if got.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, got.Email)
	}
	if got.Name != user.Name {
		t.Errorf("expected name %s, got %s", user.Name, got.Name)
	}
	if got.PasswordHash != user.PasswordHash {
		t.Errorf("expected hash %s, got %s", user.PasswordHash, got.PasswordHash)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewUserRepository(dbpool)

	user := domain.User{
		Name:         "User One",
		Email:        "dup@example.com",
		PasswordHash: "hash1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Первый раз создаём — должно быть успешно
	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed unexpectedly: %v", err)
	}

	// Попытка создать с тем же email — ожидаем ошибку
	err := repo.CreateUser(ctx, user)
	if err == nil {
		t.Fatal("expected error on duplicate email, got nil")
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewUserRepository(dbpool)

	_, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")
	if err == nil {
		t.Fatal("expected error when user not found, got nil")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("unexpected error message: %v", err)
	}
}
