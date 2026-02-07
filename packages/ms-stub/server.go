package main

import (
	"fmt"
	"net/http"
)

// NewServer creates an http.Server with routes for /, /health, and /ready.
// It listens on the given address (e.g. ":8080").
func NewServer(addr string, serviceName string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", handleRoot(serviceName))
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/ready", handleReady)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

// ListenAddr returns the address string for the given port (e.g. port "8080" -> ":8080").
func ListenAddr(port string) string {
	return fmt.Sprintf(":%s", port)
}
