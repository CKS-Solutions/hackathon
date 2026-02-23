package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleRoot(t *testing.T) {
	tests := []struct {
		name         string
		serviceName  string
		path         string
		method       string
		wantStatus   int
		wantService  string
	}{
		{"ok with service name", "ms-auth", "/", "GET", http.StatusOK, "ms-auth"},
		{"ok default", "ms-stub", "/", "GET", http.StatusOK, "ms-stub"},
		{"ok path for ingress", "ms-notify", "/notify", "GET", http.StatusOK, "ms-notify"},
		{"not found wrong method", "ms-auth", "/", "POST", http.StatusNotFound, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := handleRoot(tt.serviceName)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantService != "" {
				var body struct {
					Service string `json:"service"`
				}
				if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.Service != tt.wantService {
					t.Errorf("service = %q, want %q", body.Service, tt.wantService)
				}
			}
		})
	}
}

func TestHandleHealth(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		method     string
		wantStatus int
		wantOK     bool
	}{
		{"ok", "/health", "GET", http.StatusOK, true},
		{"not found wrong path", "/healthz", "GET", http.StatusNotFound, false},
		{"not found wrong method", "/health", "POST", http.StatusNotFound, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			handleHealth(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantOK {
				var body struct {
					Status string `json:"status"`
				}
				if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.Status != "ok" {
					t.Errorf("status = %q, want ok", body.Status)
				}
			}
		})
	}
}

func TestHandleReady(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		method     string
		wantStatus int
	}{
		{"ok", "/ready", "GET", http.StatusOK},
		{"not found wrong path", "/readyz", "GET", http.StatusNotFound},
		{"not found wrong method", "/ready", "POST", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			handleReady(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestMetricsEndpoint(t *testing.T) {
	srv := NewServer(":0", "ms-auth")
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /metrics status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if body == "" {
		t.Error("GET /metrics body is empty")
	}
	if !strings.Contains(body, "# TYPE") {
		t.Errorf("GET /metrics body should contain Prometheus output (# TYPE); got %d bytes", len(body))
	}
}
