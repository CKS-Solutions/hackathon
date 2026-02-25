package utils

import (
	"net/http"
	"testing"
)

func TestHTTPHelpers(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(string) *HttpError
		msg        string
		wantStatus int
	}{
		{"BadRequest", HTTPBadRequest, "bad", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.msg)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Error() != tt.msg {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.msg)
			}
			if err.StatusCode != tt.wantStatus {
				t.Errorf("StatusCode = %d, want %d", err.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestNewHttpError(t *testing.T) {
	msg := "custom error"
	code := 418
	err := NewHttpError(code, msg)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Error() != msg {
		t.Errorf("Error() = %q, want %q", err.Error(), msg)
	}
	if err.StatusCode != code {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, code)
	}
}
