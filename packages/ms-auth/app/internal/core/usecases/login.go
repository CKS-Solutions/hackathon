package usecases

import (
	"context"
	"errors"

	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/ports"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type LoginUsecase interface {
	Execute(ctx context.Context, input dto.LoginInput) (*dto.AuthResponse, error)
}

type LoginUsecaseImpl struct {
	userRepository ports.UserRepository
	tokenService   ports.TokenService
}

func NewLoginUsecase(userRepository ports.UserRepository, tokenService ports.TokenService) LoginUsecase {
	return &LoginUsecaseImpl{
		userRepository: userRepository,
		tokenService:   tokenService,
	}
}

func (u *LoginUsecaseImpl) Execute(ctx context.Context, input dto.LoginInput) (*dto.AuthResponse, error) {
	if input.Email == "" || input.Password == "" {
		return nil, ErrInvalidInput
	}

	user, err := u.userRepository.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.ValidatePassword(input.Password) {
		return nil, ErrInvalidCredentials
	}

	token, expiresAt, err := u.tokenService.Generate(user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.ToDTO(),
	}, nil
}
