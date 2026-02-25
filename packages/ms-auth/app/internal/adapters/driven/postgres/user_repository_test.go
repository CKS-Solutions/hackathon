package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cks-solutions/hackathon/ms-auth/internal/core/entities"
)

func TestNewUserRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	repo := NewUserRepository(db)
	if repo == nil {
		t.Fatal("NewUserRepository returned nil")
	}
}

func TestUserRepositoryImpl_Create(t *testing.T) {
	ctx := context.Background()
	user := &entities.User{
		ID:           "id-1",
		Email:        "u@x.com",
		PasswordHash: "hash",
		Name:         "User",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.ID, user.Email, user.PasswordHash, user.Name, user.CreatedAt, user.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(0, 1))
		repo := NewUserRepository(db).(*UserRepositoryImpl)
		err = repo.Create(ctx, user)
		if err != nil {
			t.Errorf("Create: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})

	t.Run("exec error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.ID, user.Email, user.PasswordHash, user.Name, user.CreatedAt, user.UpdatedAt).
			WillReturnError(errors.New("constraint violation"))
		repo := NewUserRepository(db).(*UserRepositoryImpl)
		err = repo.Create(ctx, user)
		if err == nil {
			t.Error("expected error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})
}

func TestUserRepositoryImpl_FindByEmail(t *testing.T) {
	ctx := context.Background()
	email := "u@x.com"
	now := time.Now()

	t.Run("found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT id, email, password_hash, name, created_at, updated_at").
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "created_at", "updated_at"}).
				AddRow("id-1", "u@x.com", "hash", "User", now, now))
		repo := NewUserRepository(db).(*UserRepositoryImpl)
		u, err := repo.FindByEmail(ctx, email)
		if err != nil {
			t.Errorf("FindByEmail: %v", err)
		}
		if u == nil || u.Email != email {
			t.Errorf("user = %+v", u)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT id, email, password_hash, name, created_at, updated_at").
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)
		repo := NewUserRepository(db).(*UserRepositoryImpl)
		u, err := repo.FindByEmail(ctx, email)
		if err != ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
		if u != nil {
			t.Error("expected nil user")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT id, email, password_hash, name, created_at, updated_at").
			WithArgs(email).
			WillReturnError(errors.New("db error"))
		repo := NewUserRepository(db).(*UserRepositoryImpl)
		u, err := repo.FindByEmail(ctx, email)
		if err == nil {
			t.Error("expected error")
		}
		if u != nil {
			t.Error("expected nil user")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})
}

func TestUserRepositoryImpl_FindByID(t *testing.T) {
	ctx := context.Background()
	id := "id-1"
	now := time.Now()

	t.Run("found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT id, email, password_hash, name, created_at, updated_at").
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "created_at", "updated_at"}).
				AddRow("id-1", "u@x.com", "hash", "User", now, now))
		repo := NewUserRepository(db).(*UserRepositoryImpl)
		u, err := repo.FindByID(ctx, id)
		if err != nil {
			t.Errorf("FindByID: %v", err)
		}
		if u == nil || u.ID != id {
			t.Errorf("user = %+v", u)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT id, email, password_hash, name, created_at, updated_at").
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)
		repo := NewUserRepository(db).(*UserRepositoryImpl)
		u, err := repo.FindByID(ctx, id)
		if err != ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
		if u != nil {
			t.Error("expected nil user")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT id, email, password_hash, name, created_at, updated_at").
			WithArgs(id).
			WillReturnError(errors.New("db error"))
		repo := NewUserRepository(db).(*UserRepositoryImpl)
		u, err := repo.FindByID(ctx, id)
		if err == nil {
			t.Error("expected error")
		}
		if u != nil {
			t.Error("expected nil user")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})
}
