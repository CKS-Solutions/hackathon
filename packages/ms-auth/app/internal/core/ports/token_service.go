package ports

import (
	"time"

	"github.com/cks-solutions/hackathon/ms-auth/internal/core/entities"
)

type TokenClaims struct {
	UserID string
	Email  string
}

type TokenService interface {
	Generate(user *entities.User) (token string, expiresAt time.Time, err error)
	Validate(token string) (*TokenClaims, error)
}
