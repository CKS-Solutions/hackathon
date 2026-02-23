package usecases

import (
	"context"

	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/ports"
)

type ValidateTokenUsecase interface {
	Execute(ctx context.Context, input dto.ValidateTokenInput) (*dto.ValidateTokenOutput, error)
}

type ValidateTokenUsecaseImpl struct {
	tokenService ports.TokenService
}

func NewValidateTokenUsecase(tokenService ports.TokenService) ValidateTokenUsecase {
	return &ValidateTokenUsecaseImpl{
		tokenService: tokenService,
	}
}

func (u *ValidateTokenUsecaseImpl) Execute(ctx context.Context, input dto.ValidateTokenInput) (*dto.ValidateTokenOutput, error) {
	if input.Token == "" {
		return &dto.ValidateTokenOutput{Valid: false}, nil
	}

	claims, err := u.tokenService.Validate(input.Token)
	if err != nil {
		return &dto.ValidateTokenOutput{Valid: false}, nil
	}

	return &dto.ValidateTokenOutput{
		Valid:  true,
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}
