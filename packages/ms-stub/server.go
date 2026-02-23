package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsRegistry     *prometheus.Registry
	metricsRegistryOnce sync.Once
)

func getMetricsHandler() http.Handler {
	metricsRegistryOnce.Do(func() {
		metricsRegistry = prometheus.NewRegistry()
		metricsRegistry.MustRegister(
			prometheus.NewGoCollector(),
			prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		)
	})
	return promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})
}

// NewServer creates an http.Server with routes for /, /health, /ready, and /metrics.
// It listens on the given address (e.g. ":8080").
func NewServer(addr string, serviceName string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", handleRoot(serviceName))
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/ready", handleReady)
	mux.Handle("/metrics", getMetricsHandler())

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

// ListenAddr returns the address string for the given port (e.g. port "8080" -> ":8080").
func ListenAddr(port string) string {
	return fmt.Sprintf(":%s", port)
}
