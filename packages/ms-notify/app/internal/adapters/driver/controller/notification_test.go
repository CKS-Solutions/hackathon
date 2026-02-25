package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driver/dto"
	"github.com/cks-solutions/hackathon/ms-notify/pkg/utils"
)

type mockProducer struct {
	err error
}

func (m *mockProducer) Run(ctx context.Context, input dto.NotificationInput) error {
	return m.err
}

func TestNotificationController_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid body", func(t *testing.T) {
		c := NewNotificationController(&mockProducer{})
		req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewReader([]byte("not json")))
		w := httptest.NewRecorder()
		err := c.Create(ctx, w, req)
		if err == nil {
			t.Fatal("expected error")
		}
		if he, ok := err.(*utils.HttpError); !ok || he.StatusCode != http.StatusBadRequest {
			t.Errorf("expected BadRequest, got %v", err)
		}
	})

	t.Run("usecase error", func(t *testing.T) {
		c := NewNotificationController(&mockProducer{err: errors.New("queue error")})
		body, _ := json.Marshal(dto.NotificationInput{Subject: "S", To: []string{"a@b.com"}, Html: "h"})
		req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewReader(body))
		w := httptest.NewRecorder()
		err := c.Create(ctx, w, req)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "queue error" {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		c := NewNotificationController(&mockProducer{})
		body, _ := json.Marshal(dto.NotificationInput{Subject: "Sub", To: []string{"a@b.com"}, Html: "<p>hi</p>"})
		req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewReader(body))
		w := httptest.NewRecorder()
		err := c.Create(ctx, w, req)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if w.Code != http.StatusAccepted {
			t.Errorf("status = %d, want 202", w.Code)
		}
		var decoded map[string]string
		if err := json.NewDecoder(w.Body).Decode(&decoded); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if decoded["message"] != "Notification request accepted" {
			t.Errorf("body message = %q", decoded["message"])
		}
	})
}
