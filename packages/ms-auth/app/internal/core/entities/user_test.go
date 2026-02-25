package entities

import (
	"testing"

	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driver/dto"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name    string
		input   dto.RegisterInput
		wantErr bool
		check   func(t *testing.T, u *User)
	}{
		{
			name: "valid input",
			input: dto.RegisterInput{
				Email:    "user@example.com",
				Password: "secret123",
				Name:     "Test User",
			},
			wantErr: false,
			check: func(t *testing.T, u *User) {
				if u == nil {
					t.Fatal("expected non-nil user")
				}
				if u.ID == "" {
					t.Error("expected non-empty ID")
				}
				if u.Email != "user@example.com" {
					t.Errorf("expected email user@example.com, got %q", u.Email)
				}
				if u.Name != "Test User" {
					t.Errorf("expected name Test User, got %q", u.Name)
				}
				if u.PasswordHash == "" {
					t.Error("expected non-empty password hash")
				}
				if u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
					t.Error("expected non-zero CreatedAt and UpdatedAt")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUser(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestUser_ValidatePassword(t *testing.T) {
	user, err := NewUser(dto.RegisterInput{
		Email:    "u@x.com",
		Password: "correct",
		Name:     "U",
	})
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"correct password", "correct", true},
		{"wrong password", "wrong", false},
		{"empty password", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := user.ValidatePassword(tt.password); got != tt.want {
				t.Errorf("ValidatePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_ToDTO(t *testing.T) {
	user, err := NewUser(dto.RegisterInput{
		Email:    "dto@test.com",
		Password: "p",
		Name:     "DTO User",
	})
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}

	out := user.ToDTO()
	if out.ID != user.ID {
		t.Errorf("ToDTO().ID = %q, want %q", out.ID, user.ID)
	}
	if out.Email != user.Email {
		t.Errorf("ToDTO().Email = %q, want %q", out.Email, user.Email)
	}
	if out.Name != user.Name {
		t.Errorf("ToDTO().Name = %q, want %q", out.Name, user.Name)
	}
	if !out.CreatedAt.Equal(user.CreatedAt) {
		t.Errorf("ToDTO().CreatedAt = %v, want %v", out.CreatedAt, user.CreatedAt)
	}
}
