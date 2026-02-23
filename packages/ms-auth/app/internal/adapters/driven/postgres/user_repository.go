package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cks-solutions/hackathon/ms-auth/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/ports"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) ports.UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := `
		SELECT id, email, password_hash, name, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &entities.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, id string) (*entities.User, error) {
	query := `
		SELECT id, email, password_hash, name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &entities.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}
