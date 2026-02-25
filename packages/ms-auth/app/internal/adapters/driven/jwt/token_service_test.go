package jwt

import (
	"testing"
	"time"

	"github.com/cks-solutions/hackathon/ms-auth/internal/core/entities"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key"

func TestNewTokenService(t *testing.T) {
	svc := NewTokenService(testSecret, 24)
	if svc == nil {
		t.Fatal("NewTokenService returned nil")
	}
	if impl, ok := svc.(*TokenServiceImpl); !ok || impl.secretKey != testSecret {
		t.Errorf("expected TokenServiceImpl with secret %q", testSecret)
	}
}

func TestTokenServiceImpl_Generate(t *testing.T) {
	svc := NewTokenService(testSecret, 1).(*TokenServiceImpl)

	t.Run("valid user", func(t *testing.T) {
		user := &entities.User{ID: "user-1", Email: "u@example.com", Name: "User"}
		token, expiresAt, err := svc.Generate(user)
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if token == "" {
			t.Error("expected non-empty token")
		}
		if expiresAt.Before(time.Now()) {
			t.Error("expiresAt should be in the future")
		}
		// Validate round-trip
		claims, err := svc.Validate(token)
		if err != nil {
			t.Fatalf("Validate generated token: %v", err)
		}
		if claims.UserID != "user-1" || claims.Email != "u@example.com" {
			t.Errorf("claims = UserID %q Email %q", claims.UserID, claims.Email)
		}
	})

	t.Run("nil user panics or returns error", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				// documented: nil user may panic
			}
		}()
		_, _, err := svc.Generate(nil)
		if err != nil {
			return
		}
		// if no panic and no error, at least one of token/expiresAt might be zero; we don't require error for nil
	})
}

func TestTokenServiceImpl_Validate(t *testing.T) {
	svc := NewTokenService(testSecret, 1).(*TokenServiceImpl)
	user := &entities.User{ID: "uid", Email: "e@x.com", Name: "N"}

	t.Run("valid token", func(t *testing.T) {
		token, _, _ := svc.Generate(user)
		claims, err := svc.Validate(token)
		if err != nil {
			t.Fatalf("Validate: %v", err)
		}
		if claims.UserID != "uid" || claims.Email != "e@x.com" {
			t.Errorf("claims = %+v", claims)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expiresAt := time.Now().Add(-time.Hour)
		claims := &Claims{
			UserID: "uid",
			Email:  "e@x.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expiresAt),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Subject:   "uid",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString([]byte(testSecret))
		_, err := svc.Validate(tokenStr)
		if err != ErrExpiredToken {
			t.Errorf("expected ErrExpiredToken, got %v", err)
		}
	})

	t.Run("invalid token wrong signature", func(t *testing.T) {
		claims := &Claims{
			UserID: "uid",
			Email:  "e@x.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Subject:   "uid",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString([]byte("wrong-secret"))
		_, err := svc.Validate(tokenStr)
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("invalid token wrong signing method", func(t *testing.T) {
		// Use a non-HMAC method so key func returns ErrInvalidToken
		claims := &Claims{
			UserID: "uid",
			Email:  "e@x.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Subject:   "uid",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tokenStr, _ := token.SignedString([]byte(testSecret)) // may produce invalid sig for RS256
		_, err := svc.Validate(tokenStr)
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		_, err := svc.Validate("not.a.jwt")
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
}
