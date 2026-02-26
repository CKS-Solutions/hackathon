package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cks-solutions/hackathon/ms-video/internal/core/ports"
)

// MockTokenService is a mock implementation of TokenService interface
type MockTokenService struct {
	ValidateFunc func(tokenString string) (*ports.TokenClaims, error)
}

func (m *MockTokenService) Validate(tokenString string) (*ports.TokenClaims, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(tokenString)
	}
	return nil, nil
}

func TestAuthMiddleware_Success(t *testing.T) {
	userID := "user-123"
	email := "user@example.com"

	tokenService := &MockTokenService{
		ValidateFunc: func(tokenString string) (*ports.TokenClaims, error) {
			if tokenString == "valid-token" {
				return &ports.TokenClaims{
					UserID: userID,
					Email:  email,
				}, nil
			}
			return nil, errors.New("invalid token")
		},
	}

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		// Verify context values
		ctxUserID, err := GetUserIDFromContext(r.Context())
		if err != nil {
			t.Errorf("expected no error getting userID from context, got %v", err)
		}
		if ctxUserID != userID {
			t.Errorf("expected userID '%s' in context, got '%s'", userID, ctxUserID)
		}

		ctxEmail, err := GetEmailFromContext(r.Context())
		if err != nil {
			t.Errorf("expected no error getting email from context, got %v", err)
		}
		if ctxEmail != email {
			t.Errorf("expected email '%s' in context, got '%s'", email, ctxEmail)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(tokenService, nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("expected next handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	tokenService := &MockTokenService{}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called when authorization header is missing")
	})

	handler := AuthMiddleware(tokenService, nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Message != "missing authorization header" {
		t.Errorf("expected message 'missing authorization header', got '%s'", response.Message)
	}
}

func TestAuthMiddleware_InvalidAuthorizationHeaderFormat(t *testing.T) {
	tests := []struct {
		name             string
		authHeaderValue  string
		expectedMessage  string
	}{
		{
			name:            "missing Bearer prefix",
			authHeaderValue: "invalid-token",
			expectedMessage: "invalid authorization header format",
		},
		{
			name:            "wrong prefix",
			authHeaderValue: "Basic invalid-token",
			expectedMessage: "invalid authorization header format",
		},
		{
			name:            "empty token",
			authHeaderValue: "Bearer ",
			expectedMessage: "invalid or expired token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenService := &MockTokenService{
				ValidateFunc: func(tokenString string) (*ports.TokenClaims, error) {
					return nil, errors.New("token validation failed")
				},
			}

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("next handler should not be called with invalid auth header format")
			})

			handler := AuthMiddleware(tokenService, nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.authHeaderValue)

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, w.Code)
			}
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	tokenService := &MockTokenService{
		ValidateFunc: func(tokenString string) (*ports.TokenClaims, error) {
			return nil, errors.New("token validation failed")
		},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called when token is invalid")
	})

	handler := AuthMiddleware(tokenService, nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Message != "invalid or expired token" {
		t.Errorf("expected message 'invalid or expired token', got '%s'", response.Message)
	}
}

func TestGetUserIDFromContext_Success(t *testing.T) {
	expectedUserID := "user-123"
	ctx := context.WithValue(context.Background(), UserIDContextKey, expectedUserID)

	userID, err := GetUserIDFromContext(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if userID != expectedUserID {
		t.Errorf("expected userID '%s', got '%s'", expectedUserID, userID)
	}
}

func TestGetUserIDFromContext_Missing(t *testing.T) {
	ctx := context.Background()

	userID, err := GetUserIDFromContext(ctx)

	if err == nil {
		t.Fatal("expected error when userID is missing from context, got nil")
	}

	if userID != "" {
		t.Errorf("expected empty userID, got '%s'", userID)
	}
}

func TestGetUserIDFromContext_WrongType(t *testing.T) {
	// Set value with wrong type
	ctx := context.WithValue(context.Background(), UserIDContextKey, 12345)

	userID, err := GetUserIDFromContext(ctx)

	if err == nil {
		t.Fatal("expected error when userID has wrong type, got nil")
	}

	if userID != "" {
		t.Errorf("expected empty userID, got '%s'", userID)
	}
}

func TestGetEmailFromContext_Success(t *testing.T) {
	expectedEmail := "user@example.com"
	ctx := context.WithValue(context.Background(), EmailContextKey, expectedEmail)

	email, err := GetEmailFromContext(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if email != expectedEmail {
		t.Errorf("expected email '%s', got '%s'", expectedEmail, email)
	}
}

func TestGetEmailFromContext_Missing(t *testing.T) {
	ctx := context.Background()

	email, err := GetEmailFromContext(ctx)

	if err == nil {
		t.Fatal("expected error when email is missing from context, got nil")
	}

	if email != "" {
		t.Errorf("expected empty email, got '%s'", email)
	}
}

func TestGetEmailFromContext_WrongType(t *testing.T) {
	// Set value with wrong type
	ctx := context.WithValue(context.Background(), EmailContextKey, 12345)

	email, err := GetEmailFromContext(ctx)

	if err == nil {
		t.Fatal("expected error when email has wrong type, got nil")
	}

	if email != "" {
		t.Errorf("expected empty email, got '%s'", email)
	}
}

func TestAuthMiddleware_ContentTypeHeader(t *testing.T) {
	tokenService := &MockTokenService{}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handler := AuthMiddleware(tokenService, nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}
}
