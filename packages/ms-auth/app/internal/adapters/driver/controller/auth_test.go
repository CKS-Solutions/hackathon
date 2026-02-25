package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cks-solutions/hackathon/ms-auth/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/usecases"
	"github.com/cks-solutions/hackathon/ms-auth/pkg/utils"
)

type mockRegister struct {
	out *dto.UserOutput
	err error
}

func (m *mockRegister) Execute(ctx context.Context, input dto.RegisterInput) (*dto.UserOutput, error) {
	return m.out, m.err
}

type mockLogin struct {
	out *dto.AuthResponse
	err error
}

func (m *mockLogin) Execute(ctx context.Context, input dto.LoginInput) (*dto.AuthResponse, error) {
	return m.out, m.err
}

type mockValidateToken struct {
	out *dto.ValidateTokenOutput
	err error
}

func (m *mockValidateToken) Execute(ctx context.Context, input dto.ValidateTokenInput) (*dto.ValidateTokenOutput, error) {
	return m.out, m.err
}

func TestAuthController_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("method not POST", func(t *testing.T) {
		c := NewAuthController(&mockRegister{}, &mockLogin{}, &mockValidateToken{})
		req := httptest.NewRequest(http.MethodGet, "/register", nil)
		w := httptest.NewRecorder()
		err := c.Register(ctx, w, req)
		if err == nil {
			t.Fatal("expected error")
		}
		if he, ok := err.(*utils.HTTPError); !ok || he.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected MethodNotAllowed, got %v", err)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		c := NewAuthController(&mockRegister{}, &mockLogin{}, &mockValidateToken{})
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte("not json")))
		w := httptest.NewRecorder()
		err := c.Register(ctx, w, req)
		if err == nil {
			t.Fatal("expected error")
		}
		if he, ok := err.(*utils.HTTPError); !ok || he.StatusCode != http.StatusBadRequest {
			t.Errorf("expected BadRequest, got %v", err)
		}
	})

	t.Run("email already exists", func(t *testing.T) {
		c := NewAuthController(
			&mockRegister{err: usecases.ErrEmailAlreadyExists},
			&mockLogin{},
			&mockValidateToken{},
		)
		body, _ := json.Marshal(dto.RegisterInput{Email: "a@b.com", Password: "p", Name: "N"})
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()
		err := c.Register(ctx, w, req)
		if err == nil {
			t.Fatal("expected error")
		}
		if he, ok := err.(*utils.HTTPError); !ok || he.StatusCode != http.StatusConflict {
			t.Errorf("expected Conflict, got %v", err)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		c := NewAuthController(
			&mockRegister{err: usecases.ErrInvalidInput},
			&mockLogin{},
			&mockValidateToken{},
		)
		body, _ := json.Marshal(dto.RegisterInput{Email: "", Password: "p", Name: "N"})
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()
		err := c.Register(ctx, w, req)
		if err == nil {
			t.Fatal("expected error")
		}
		if he, ok := err.(*utils.HTTPError); !ok || he.StatusCode != http.StatusBadRequest {
			t.Errorf("expected BadRequest, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		user := &dto.UserOutput{ID: "1", Email: "u@b.com", Name: "U"}
		c := NewAuthController(
			&mockRegister{out: user},
			&mockLogin{},
			&mockValidateToken{},
		)
		body, _ := json.Marshal(dto.RegisterInput{Email: "u@b.com", Password: "p", Name: "U"})
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()
		err := c.Register(ctx, w, req)
		if err != nil {
			t.Fatalf("Register: %v", err)
		}
		if w.Code != http.StatusCreated {
			t.Errorf("status = %d, want 201", w.Code)
		}
		var decoded dto.UserOutput
		if err := json.NewDecoder(w.Body).Decode(&decoded); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if decoded.Email != "u@b.com" {
			t.Errorf("body.Email = %q", decoded.Email)
		}
	})
}

func TestAuthController_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("method not POST", func(t *testing.T) {
		c := NewAuthController(&mockRegister{}, &mockLogin{}, &mockValidateToken{})
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		w := httptest.NewRecorder()
		err := c.Login(ctx, w, req)
		if err == nil {
			t.Fatal("expected error")
		}
		if he, ok := err.(*utils.HTTPError); !ok || he.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected MethodNotAllowed, got %v", err)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		c := NewAuthController(&mockRegister{}, &mockLogin{}, &mockValidateToken{})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("x")))
		w := httptest.NewRecorder()
		err := c.Login(ctx, w, req)
		if err == nil {
			t.Fatal("expected error")
		}
		if he, ok := err.(*utils.HTTPError); !ok || he.StatusCode != http.StatusBadRequest {
			t.Errorf("expected BadRequest, got %v", err)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		c := NewAuthController(
			&mockRegister{},
			&mockLogin{err: usecases.ErrInvalidCredentials},
			&mockValidateToken{},
		)
		body, _ := json.Marshal(dto.LoginInput{Email: "a@b.com", Password: "p"})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		w := httptest.NewRecorder()
		err := c.Login(ctx, w, req)
		if err == nil {
			t.Fatal("expected error")
		}
		if he, ok := err.(*utils.HTTPError); !ok || he.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected Unauthorized, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		resp := &dto.AuthResponse{Token: "jwt", User: dto.UserOutput{Email: "u@b.com"}}
		c := NewAuthController(
			&mockRegister{},
			&mockLogin{out: resp},
			&mockValidateToken{},
		)
		body, _ := json.Marshal(dto.LoginInput{Email: "u@b.com", Password: "p"})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		w := httptest.NewRecorder()
		err := c.Login(ctx, w, req)
		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", w.Code)
		}
		var decoded dto.AuthResponse
		if err := json.NewDecoder(w.Body).Decode(&decoded); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if decoded.Token != "jwt" {
			t.Errorf("body.Token = %q", decoded.Token)
		}
	})
}

func TestAuthController_ValidateToken(t *testing.T) {
	ctx := context.Background()

	t.Run("method not POST", func(t *testing.T) {
		c := NewAuthController(&mockRegister{}, &mockLogin{}, &mockValidateToken{})
		req := httptest.NewRequest(http.MethodGet, "/validate", nil)
		w := httptest.NewRecorder()
		err := c.ValidateToken(ctx, w, req)
		if err == nil {
			t.Fatal("expected error")
		}
		if he, ok := err.(*utils.HTTPError); !ok || he.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected MethodNotAllowed, got %v", err)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		c := NewAuthController(&mockRegister{}, &mockLogin{}, &mockValidateToken{})
		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader([]byte("x")))
		w := httptest.NewRecorder()
		err := c.ValidateToken(ctx, w, req)
		if err == nil {
			t.Fatal("expected error")
		}
		if he, ok := err.(*utils.HTTPError); !ok || he.StatusCode != http.StatusBadRequest {
			t.Errorf("expected BadRequest, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		out := &dto.ValidateTokenOutput{Valid: true, UserID: "uid", Email: "u@e.com"}
		c := NewAuthController(
			&mockRegister{},
			&mockLogin{},
			&mockValidateToken{out: out},
		)
		body, _ := json.Marshal(dto.ValidateTokenInput{Token: "jwt"})
		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()
		err := c.ValidateToken(ctx, w, req)
		if err != nil {
			t.Fatalf("ValidateToken: %v", err)
		}
		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", w.Code)
		}
		var decoded dto.ValidateTokenOutput
		if err := json.NewDecoder(w.Body).Decode(&decoded); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !decoded.Valid || decoded.UserID != "uid" {
			t.Errorf("body Valid=%v UserID=%q", decoded.Valid, decoded.UserID)
		}
	})
}
