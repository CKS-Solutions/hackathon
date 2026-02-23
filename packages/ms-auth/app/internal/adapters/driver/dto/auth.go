package dto

import "time"

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	User      UserOutput `json:"user"`
}

type UserOutput struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ValidateTokenInput struct {
	Token string `json:"token"`
}

type ValidateTokenOutput struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
}
