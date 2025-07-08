package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/handler"
	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/service"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	users map[string]domain.User
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	u, ok := m.users[email]
	if !ok {
		return domain.User{}, pgx.ErrNoRows
	}
	return u, nil
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user domain.User) error {
	if _, exists := m.users[user.Email]; exists {
		return errors.New("already exists")
	}
	m.users[user.Email] = user
	return nil
}

func TestRegiserHandler_Success(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	authService := service.NewAuthService(repo)
	h := handler.NewAuthHandler(authService)

	payload := `{
		"name": "Alex",
		"email": "alex@email.com",
		"password": "secret"
	}`

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterHandler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "User registred") {
		t.Fatalf("expected success message, got %s", rec.Body.String())
	}
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	authService := service.NewAuthService(repo)
	h := handler.NewAuthHandler(authService)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("{invalid-json}"))
	rec := httptest.NewRecorder()

	h.RegisterHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestRegisterHandler_Duplicate(t *testing.T) {
	repo := &mockUserRepo{
		users: map[string]domain.User{
			"alex@email.com": {Email: "alex@email.com"},
		},
	}
	authService := service.NewAuthService(repo)
	h := handler.NewAuthHandler(authService)

	payload := `{
		"name": "Alex",
		"email": "alex@email.com",
		"password": "secret"
	}`

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for duplicate, got %d", rec.Code)
	}
}

func TestLoginHandler_Success(t *testing.T) {
	password := "secret"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	repo := &mockUserRepo{users: map[string]domain.User{
		"user@email.com": {
			ID:           1,
			Email:        "user@email.com",
			PasswordHash: string(hashed),
		},
	}}
	authServise := service.NewAuthService(repo)
	h := handler.NewAuthHandler(authServise)

	os.Setenv("JWT_SECRET", "supersecretkey")
	defer os.Unsetenv("JWT_SECRET")

	body := map[string]string{"email": "user@email.com", "password": "secret"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.LoginHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "token") {
		t.Fatalf("expected token is response got %s", rec.Body.String())
	}
}

func TestLoginHandler_InvalidPassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)

	repo := &mockUserRepo{
		users: map[string]domain.User{
			"user@email.com": {
				Email:        "user@email.com",
				PasswordHash: string(hashedPassword),
			},
		},
	}
	authService := service.NewAuthService(repo)
	h := handler.NewAuthHandler(authService)

	os.Setenv("JWT_SECRET", "supersecretkey")
	defer os.Unsetenv("JWT_SECRET")

	payload := `{"email":"user@email.com","password":"wrongpass"}`

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.LoginHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid password, got %d", rec.Code)
	}
}

func TestLoginHandler_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]domain.User)}
	authService := service.NewAuthService(repo)
	h := handler.NewAuthHandler(authService)

	os.Setenv("JWT_SECRET", "supersecretkey")
	defer os.Unsetenv("JWT_SECRET")

	body := map[string]string{"email": "notfound@email.com", "password": "secret"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.LoginHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
