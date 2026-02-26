package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driven/jwt"
	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driven/postgres"
	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driver/controller"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/usecases"
	"github.com/cks-solutions/hackathon/ms-auth/pkg/utils"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Router struct {
	*http.ServeMux
	Ctx context.Context
}

func handleError(w http.ResponseWriter, err error) {
	var httpErr *utils.HTTPError
	if errors.As(err, &httpErr) {
		w.WriteHeader(httpErr.StatusCode)
		json.NewEncoder(w).Encode(ErrorResponse{
			Message: err.Error(),
		})
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Message: "internal server error",
		})
		log.Println("Internal error:", err)
	}
}

func NewRouter(ctx context.Context, db *sql.DB, jwtSecret string, jwtExpiration int) *Router {
	mux := &Router{ServeMux: http.NewServeMux(), Ctx: ctx}

	userRepository := postgres.NewUserRepository(db)
	tokenService := jwt.NewTokenService(jwtSecret, jwtExpiration)

	registerUsecase := usecases.NewRegisterUsecase(userRepository)
	loginUsecase := usecases.NewLoginUsecase(userRepository, tokenService)
	validateTokenUsecase := usecases.NewValidateTokenUsecase(tokenService)

	authController := controller.NewAuthController(registerUsecase, loginUsecase, validateTokenUsecase)

	healthResp := []byte(`{"status":"healthy","service":"ms-auth"}`)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(healthResp)
	})
	mux.HandleFunc("/auth/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(healthResp)
	})

	mux.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := authController.Register(ctx, w, r); err != nil {
			handleError(w, err)
		}
	})

	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := authController.Login(ctx, w, r); err != nil {
			handleError(w, err)
		}
	})

	mux.HandleFunc("/auth/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := authController.ValidateToken(ctx, w, r); err != nil {
			handleError(w, err)
		}
	})

	return mux
}
