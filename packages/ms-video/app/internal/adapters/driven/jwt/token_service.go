package jwt

import (
	"errors"

	"github.com/cks-solutions/hackathon/ms-video/internal/core/ports"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type TokenServiceImpl struct {
	secretKey string
}

func NewTokenService(secretKey string) ports.TokenService {
	return &TokenServiceImpl{
		secretKey: secretKey,
	}
}

func (s *TokenServiceImpl) Validate(tokenString string) (*ports.TokenClaims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return &ports.TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}
