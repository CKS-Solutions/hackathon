package main

import (
	"encoding/json"
	"net/http"
)

// handleRoot returns 200 with a JSON body containing the service name (from SERVICE_NAME env).
// Accepts GET on any path so path-based Ingress (e.g. /auth, /video, /notify) works without rewrite.
func handleRoot(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"service": serviceName})
	}
}

// handleHealth returns 200 with {"status":"ok"} for liveness probes.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/health" || r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleReady returns 200 for readiness probes (always ready in this stub).
func handleReady(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ready" || r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}
