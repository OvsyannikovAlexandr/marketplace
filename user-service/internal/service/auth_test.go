package service_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/service"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	users map[string]domain.User
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	user, ok := m.users[email]
	if !ok {
		return domain.User{}, pgx.ErrNoRows
	}
	return user, nil
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user domain.User) error {
	if _, exists := m.users[user.Email]; exists {
		return errors.New("already exists")
	}
	m.users[user.Email] = user
	return nil
}

func TestRegister_Success(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	authService := service.NewAuthService(repo)

	err := authService.Register(context.Background(), "Alex", "alex@email.com", "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRegister_Duplicate(t *testing.T) {
	repo := &mockUserRepo{
		users: map[string]domain.User{
			"alex@email.com": {Email: "alex@email.com"},
		},
	}
	authService := service.NewAuthService(repo)

	err := authService.Register(context.Background(), "Alex", "alex@email.com", "secret")
	if err == nil || err.Error() != "user already exists" {
		t.Fatalf("expected 'user already exists', got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	// хеш пароля "secret"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)

	repo := &mockUserRepo{users: map[string]domain.User{
		"alex@email.com": {
			ID:           1,
			Email:        "alex@email.com",
			PasswordHash: string(hashedPassword),
		},
	}}

	authService := service.NewAuthService(repo)

	os.Setenv("JWT_SECRET", "supersecretkey")
	defer os.Unsetenv("JWT_SECRET")

	token, err := authService.Login(context.Background(), "alex@email.com", "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatalf("expected token, got empty string")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)

	repo := &mockUserRepo{users: map[string]domain.User{
		"user@email.com": {
			Email:        "user@email.com",
			PasswordHash: string(hashedPassword),
		},
	}}
	authService := service.NewAuthService(repo)

	os.Setenv("JWT_SECRET", "supersecretkey")
	defer os.Unsetenv("JWT_SECRET")

	_, err := authService.Login(context.Background(), "user@email.com", "wrongpass")
	if err == nil || err.Error() != "invalid password" {
		t.Fatalf("expected invalid password error, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{users: map[string]domain.User{}}
	authService := service.NewAuthService(repo)

	os.Setenv("JWT_SECRET", "supersecretkey")
	defer os.Unsetenv("JWT_SECRET")

	_, err := authService.Login(context.Background(), "missing@email.com", "secret")
	if err == nil || err.Error() != "invalid email, or password" {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestLogin_MissingJWTSecret(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)

	repo := &mockUserRepo{users: map[string]domain.User{
		"user@email.com": {
			Email:        "user@email.com",
			PasswordHash: string(hashedPassword),
		},
	}}

	authService := service.NewAuthService(repo)

	os.Unsetenv("JWT_SECRET") // ❗️ удаляем секрет

	_, err := authService.Login(context.Background(), "user@email.com", "secret")
	if err == nil || err.Error() != "jwt secret not ser" {
		t.Fatalf("expected jwt secret error, got %v", err)
	}
}
