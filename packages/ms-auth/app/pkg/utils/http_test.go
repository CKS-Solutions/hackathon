package utils

import (
	"net/http"
	"testing"
)

func TestHTTPHelpers(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(string) *HTTPError
		msg        string
		wantStatus int
	}{
		{"BadRequest", HTTPBadRequest, "bad", http.StatusBadRequest},
		{"Unauthorized", HTTPUnauthorized, "unauthorized", http.StatusUnauthorized},
		{"Conflict", HTTPConflict, "conflict", http.StatusConflict},
		{"MethodNotAllowed", HTTPMethodNotAllowed, "not allowed", http.StatusMethodNotAllowed},
		{"InternalServerError", HTTPInternalServerError, "internal", http.StatusInternalServerError},
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
