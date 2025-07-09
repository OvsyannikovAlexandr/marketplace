//go:build ci

package repository_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbpool *pgxpool.Pool

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

	_, err = dbpool.Exec(ctx, `DROP SCHEMA IF EXISTS user_service CASCADE`)
	if err != nil {
		panic(err)
	}

	schema := `
		CREATE SCHEMA IF NOT EXISTS user_service;

		CREATE TABLE user_service.users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			price NUMERIC(10,2) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			updated_at TIMESTAMP NOT NULL DEFAULT now()
		);`

	_, err = dbpool.Exec(ctx, schema)
	if err != nil {
		panic(err)
	}

	_, err = dbpool.Exec(ctx, `DELETE FROM user_service.users`)
	if err != nil {
		panic("failed to clear table: " + err.Error())
	}

	code := m.Run()
	os.Exit(code)
}

func clearUsersTable(t *testing.T) {
	_, err := dbpool.Exec(context.Background(), `DELETE FROM user_service.users`)
	if err != nil {
		t.Fatalf("failed to clear users table: %v", err)
	}
}

func TestCreateAndGetUser(t *testing.T) {
	clearUsersTable(t)
	ctx := context.Background()
	repo := repository.NewUserRepository(dbpool)

	user := domain.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hash123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	got, err := repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}

	if got.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, got.Email)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	clearUsersTable(t)
	ctx := context.Background()
	repo := repository.NewUserRepository(dbpool)

	user := domain.User{
		Name:         "User One",
		Email:        "dup@example.com",
		PasswordHash: "hash1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed unexpectedly: %v", err)
	}

	if err := repo.CreateUser(ctx, user); err == nil {
		t.Fatal("expected error on duplicate email, got nil")
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	clearUsersTable(t)
	ctx := context.Background()
	repo := repository.NewUserRepository(dbpool)

	_, err := repo.GetUserByEmail(ctx, "missing@example.com")
	if err == nil {
		t.Fatal("expected error when user not found, got nil")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("unexpected error: %v", err)
	}
}
