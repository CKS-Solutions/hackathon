package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultPort        = "8080"
	defaultServiceName = "ms-stub"
	shutdownTimeout    = 10 * time.Second
)

func main() {
	port := getEnv("PORT", defaultPort)
	serviceName := getEnv("SERVICE_NAME", defaultServiceName)

	addr := ListenAddr(port)
	srv := NewServer(addr, serviceName)

	go func() {
		slog.Info("server listening", "addr", addr, "service", serviceName)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
	os.Exit(0)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
