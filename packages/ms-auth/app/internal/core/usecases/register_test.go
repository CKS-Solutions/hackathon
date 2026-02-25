package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/entities"
)

type fakeUserRepo struct {
	findByEmailUser *entities.User
	createErr       error
}

func (f *fakeUserRepo) Create(ctx context.Context, user *entities.User) error {
	return f.createErr
}

func (f *fakeUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return f.findByEmailUser, nil
}

func (f *fakeUserRepo) FindByID(ctx context.Context, id string) (*entities.User, error) {
	return nil, nil
}

func TestRegisterUsecase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		input   dto.RegisterInput
		repo    *fakeUserRepo
		wantErr error
		check   func(t *testing.T, out *dto.UserOutput)
	}{
		{
			name: "empty email",
			input: dto.RegisterInput{Email: "", Password: "p", Name: "N"},
			repo:  &fakeUserRepo{},
			wantErr: ErrInvalidInput,
		},
		{
			name: "empty password",
			input: dto.RegisterInput{Email: "a@b.com", Password: "", Name: "N"},
			repo:  &fakeUserRepo{},
			wantErr: ErrInvalidInput,
		},
		{
			name: "empty name",
			input: dto.RegisterInput{Email: "a@b.com", Password: "p", Name: ""},
			repo:  &fakeUserRepo{},
			wantErr: ErrInvalidInput,
		},
		{
			name: "email already exists",
			input: dto.RegisterInput{Email: "exists@b.com", Password: "p", Name: "N"},
			repo: &fakeUserRepo{
				findByEmailUser: &entities.User{ID: "1", Email: "exists@b.com", Name: "Existing"},
			},
			wantErr: ErrEmailAlreadyExists,
		},
		{
			name:  "create returns error",
			input: dto.RegisterInput{Email: "new@b.com", Password: "secret", Name: "New"},
			repo:  &fakeUserRepo{createErr: errors.New("db error")},
			wantErr: errors.New("db error"),
		},
		{
			name:  "success",
			input: dto.RegisterInput{Email: "ok@b.com", Password: "pass", Name: "Ok User"},
			repo:  &fakeUserRepo{},
			check: func(t *testing.T, out *dto.UserOutput) {
				if out == nil {
					t.Fatal("expected non-nil output")
				}
				if out.Email != "ok@b.com" {
					t.Errorf("output.Email = %q, want ok@b.com", out.Email)
				}
				if out.Name != "Ok User" {
					t.Errorf("output.Name = %q, want Ok User", out.Name)
				}
				if out.ID == "" {
					t.Error("expected non-empty ID")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewRegisterUsecase(tt.repo)
			got, err := uc.Execute(ctx, tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Execute() err = nil, want %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Errorf("Execute() err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Execute() err = %v", err)
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
