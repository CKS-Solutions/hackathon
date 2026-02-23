package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/usecases"
	"github.com/cks-solutions/hackathon/ms-auth/pkg/utils"
)

type AuthController interface {
	Register(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	Login(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	ValidateToken(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

type AuthControllerImpl struct {
	registerUsecase      usecases.RegisterUsecase
	loginUsecase         usecases.LoginUsecase
	validateTokenUsecase usecases.ValidateTokenUsecase
}

func NewAuthController(
	registerUsecase usecases.RegisterUsecase,
	loginUsecase usecases.LoginUsecase,
	validateTokenUsecase usecases.ValidateTokenUsecase,
) AuthController {
	return &AuthControllerImpl{
		registerUsecase:      registerUsecase,
		loginUsecase:         loginUsecase,
		validateTokenUsecase: validateTokenUsecase,
	}
}

func (c *AuthControllerImpl) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.HTTPMethodNotAllowed("method not allowed")
	}

	var input dto.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return utils.HTTPBadRequest("invalid request body")
	}

	user, err := c.registerUsecase.Execute(ctx, input)
	if err != nil {
		if err == usecases.ErrEmailAlreadyExists {
			return utils.HTTPConflict(err.Error())
		}
		if err == usecases.ErrInvalidInput {
			return utils.HTTPBadRequest(err.Error())
		}
		return utils.HTTPInternalServerError("failed to register user")
	}

	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(user)
}

func (c *AuthControllerImpl) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.HTTPMethodNotAllowed("method not allowed")
	}

	var input dto.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return utils.HTTPBadRequest("invalid request body")
	}

	response, err := c.loginUsecase.Execute(ctx, input)
	if err != nil {
		if err == usecases.ErrInvalidCredentials {
			return utils.HTTPUnauthorized(err.Error())
		}
		return utils.HTTPInternalServerError("failed to login")
	}

	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func (c *AuthControllerImpl) ValidateToken(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.HTTPMethodNotAllowed("method not allowed")
	}

	var input dto.ValidateTokenInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return utils.HTTPBadRequest("invalid request body")
	}

	result, err := c.validateTokenUsecase.Execute(ctx, input)
	if err != nil {
		return utils.HTTPInternalServerError("failed to validate token")
	}

	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(result)
}
