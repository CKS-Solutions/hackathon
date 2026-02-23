package usecases

import (
	"context"
	"errors"

	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/ports"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidInput       = errors.New("invalid input")
)

type RegisterUsecase interface {
	Execute(ctx context.Context, input dto.RegisterInput) (*dto.UserOutput, error)
}

type RegisterUsecaseImpl struct {
	userRepository ports.UserRepository
}

func NewRegisterUsecase(userRepository ports.UserRepository) RegisterUsecase {
	return &RegisterUsecaseImpl{
		userRepository: userRepository,
	}
}

func (u *RegisterUsecaseImpl) Execute(ctx context.Context, input dto.RegisterInput) (*dto.UserOutput, error) {
	if input.Email == "" || input.Password == "" || input.Name == "" {
		return nil, ErrInvalidInput
	}

	existingUser, _ := u.userRepository.FindByEmail(ctx, input.Email)
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	user, err := entities.NewUser(input)
	if err != nil {
		return nil, err
	}

	err = u.userRepository.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	output := user.ToDTO()
	return &output, nil
}
