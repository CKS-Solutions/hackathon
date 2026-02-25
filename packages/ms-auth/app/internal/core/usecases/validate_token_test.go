package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/ports"
)

func TestValidateTokenUsecase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		input  dto.ValidateTokenInput
		svc    *fakeTokenService
		valid  bool
		userID string
		email  string
	}{
		{
			name:  "empty token",
			input: dto.ValidateTokenInput{Token: ""},
			svc:   &fakeTokenService{},
			valid: false,
		},
		{
			name:  "validate error",
			input: dto.ValidateTokenInput{Token: "bad"},
			svc:   &fakeTokenService{validateErr: errors.New("invalid")},
			valid: false,
		},
		{
			name:   "success",
			input:  dto.ValidateTokenInput{Token: "valid.jwt"},
			svc:    &fakeTokenService{validate: &ports.TokenClaims{UserID: "uid-1", Email: "u@e.com"}},
			valid:  true,
			userID: "uid-1",
			email:  "u@e.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewValidateTokenUsecase(tt.svc)
			got, err := uc.Execute(ctx, tt.input)
			if err != nil {
				t.Errorf("Execute() err = %v", err)
				return
			}
			if got.Valid != tt.valid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.valid)
			}
			if got.UserID != tt.userID {
				t.Errorf("UserID = %q, want %q", got.UserID, tt.userID)
			}
			if got.Email != tt.email {
				t.Errorf("Email = %q, want %q", got.Email, tt.email)
			}
		})
	}
}
