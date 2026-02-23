package entities

import (
	"time"

	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driver/dto"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewUser(input dto.RegisterInput) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:           uuid.NewString(),
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		Name:         input.Name,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (u *User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) ToDTO() dto.UserOutput {
	return dto.UserOutput{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
	}
}
