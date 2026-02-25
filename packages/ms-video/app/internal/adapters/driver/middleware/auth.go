package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cks-solutions/hackathon/ms-video/internal/core/ports"
	"github.com/cks-solutions/hackathon/ms-video/pkg/utils"
)

type contextKey string

const UserIDContextKey contextKey = "userID"
const EmailContextKey contextKey = "email"

type ErrorResponse struct {
	Message string `json:"message"`
}

func AuthMiddleware(tokenService ports.TokenService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "missing authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := tokenService.Validate(tokenString)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid or expired token"})
			return
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailContextKey, claims.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	if !ok {
		return "", utils.NewUnauthorizedError("user not authenticated")
	}
	return userID, nil
}

func GetEmailFromContext(ctx context.Context) (string, error) {
	email, ok := ctx.Value(EmailContextKey).(string)
	if !ok {
		return "", utils.NewUnauthorizedError("user not authenticated")
	}
	return email, nil
}
