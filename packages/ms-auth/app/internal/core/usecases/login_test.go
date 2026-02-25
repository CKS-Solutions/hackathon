package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/ports"
)

type fakeLoginUserRepo struct {
	findByEmailUser *entities.User
	findByEmailErr  error
}

func (f *fakeLoginUserRepo) Create(ctx context.Context, user *entities.User) error {
	return nil
}

func (f *fakeLoginUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	if f.findByEmailErr != nil {
		return nil, f.findByEmailErr
	}
	return f.findByEmailUser, nil
}

func (f *fakeLoginUserRepo) FindByID(ctx context.Context, id string) (*entities.User, error) {
	return nil, nil
}

type fakeTokenService struct {
	token     string
	expiresAt time.Time
	genErr    error
	validate  *ports.TokenClaims
	validateErr error
}

func (f *fakeTokenService) Generate(user *entities.User) (string, time.Time, error) {
	if f.genErr != nil {
		return "", time.Time{}, f.genErr
	}
	return f.token, f.expiresAt, nil
}

func (f *fakeTokenService) Validate(token string) (*ports.TokenClaims, error) {
	if f.validateErr != nil {
		return nil, f.validateErr
	}
	return f.validate, nil
}

func TestLoginUsecase_Execute(t *testing.T) {
	ctx := context.Background()

	userWithPass, err := entities.NewUser(dto.RegisterInput{
		Email:    "login@test.com",
		Password: "correctpass",
		Name:     "Login User",
	})
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}

	tests := []struct {
		name    string
		input   dto.LoginInput
		repo    *fakeLoginUserRepo
		svc     *fakeTokenService
		wantErr error
		check   func(t *testing.T, out *dto.AuthResponse)
	}{
		{
			name:    "empty email",
			input:   dto.LoginInput{Email: "", Password: "p"},
			repo:    &fakeLoginUserRepo{},
			wantErr: ErrInvalidInput,
		},
		{
			name:    "empty password",
			input:   dto.LoginInput{Email: "a@b.com", Password: ""},
			repo:    &fakeLoginUserRepo{},
			wantErr: ErrInvalidInput,
		},
		{
			name:    "user not found",
			input:   dto.LoginInput{Email: "nobody@b.com", Password: "p"},
			repo:    &fakeLoginUserRepo{findByEmailErr: errors.New("not found")},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:    "user nil",
			input:   dto.LoginInput{Email: "nobody@b.com", Password: "p"},
			repo:    &fakeLoginUserRepo{findByEmailUser: nil},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:  "wrong password",
			input: dto.LoginInput{Email: "login@test.com", Password: "wrong"},
			repo:  &fakeLoginUserRepo{findByEmailUser: userWithPass},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:  "token generate error",
			input: dto.LoginInput{Email: "login@test.com", Password: "correctpass"},
			repo:  &fakeLoginUserRepo{findByEmailUser: userWithPass},
			svc:   &fakeTokenService{genErr: errors.New("token error")},
			wantErr: errors.New("token error"),
		},
		{
			name:  "success",
			input: dto.LoginInput{Email: "login@test.com", Password: "correctpass"},
			repo:  &fakeLoginUserRepo{findByEmailUser: userWithPass},
			svc:   &fakeTokenService{token: "jwt.here", expiresAt: time.Unix(1, 0)},
			check: func(t *testing.T, out *dto.AuthResponse) {
				if out == nil {
					t.Fatal("expected non-nil response")
				}
				if out.Token != "jwt.here" {
					t.Errorf("Token = %q, want jwt.here", out.Token)
				}
				if !out.ExpiresAt.Equal(time.Unix(1, 0)) {
					t.Errorf("ExpiresAt = %v", out.ExpiresAt)
				}
				if out.User.Email != "login@test.com" {
					t.Errorf("User.Email = %q", out.User.Email)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc
			if svc == nil {
				svc = &fakeTokenService{}
			}
			uc := NewLoginUsecase(tt.repo, svc)
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
