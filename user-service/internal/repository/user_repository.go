package repository

import (
	"context"
	"time"

	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
}

func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) error {
	query := `
		INSERT INTO user_service.users (name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, user.Name, user.Email, user.PasswordHash, time.Now(), time.Now())
	return err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM user_service.users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}
