package jwt

import (
	"errors"
	"time"

	"github.com/cks-solutions/hackathon/ms-auth/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/ports"
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
	secretKey      string
	expirationTime time.Duration
}

func NewTokenService(secretKey string, expirationHours int) ports.TokenService {
	return &TokenServiceImpl{
		secretKey:      secretKey,
		expirationTime: time.Duration(expirationHours) * time.Hour,
	}
}

func (s *TokenServiceImpl) Generate(user *entities.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.expirationTime)

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
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
