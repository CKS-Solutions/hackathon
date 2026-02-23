package ports

import (
	"context"

	"github.com/cks-solutions/hackathon/ms-auth/internal/core/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, id string) (*entities.User, error)
}
