package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/turtacn/guocedb/observability/diagnostic"
	"github.com/turtacn/guocedb/observability/health"
	"github.com/turtacn/guocedb/observability/metrics"
)

// ServerConfig configures the observability server
type ServerConfig struct {
	Enabled     bool
	Address     string // e.g., ":9090"
	MetricsPath string // e.g., "/metrics"
	EnablePprof bool
}

// DefaultConfig returns default observability configuration
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Enabled:     true,
		Address:     ":9090",
		MetricsPath: "/metrics",
		EnablePprof: true,
	}
}

// Server manages observability HTTP endpoints
type Server struct {
	config      ServerConfig
	httpServer  *http.Server
	checker     *health.Checker
	diagnostics *diagnostic.Diagnostics
}

// NewServer creates a new observability server
func NewServer(config ServerConfig, checker *health.Checker, diag *diagnostic.Diagnostics) *Server {
	return &Server{
		config:      config,
		checker:     checker,
		diagnostics: diag,
	}
}

// Start starts the observability server
func (s *Server) Start() error {
	if !s.config.Enabled {
		return nil
	}

	mux := http.NewServeMux()

	// Metrics endpoint
	metricsPath := s.config.MetricsPath
	if metricsPath == "" {
		metricsPath = "/metrics"
	}
	mux.Handle(metricsPath, metrics.Handler())

	// Health check endpoints
	if s.checker != nil {
		mux.HandleFunc("/health", s.healthHandler)
		mux.HandleFunc("/ready", s.readyHandler)
		mux.HandleFunc("/live", s.liveHandler)
	}

	// Diagnostic endpoints
	if s.diagnostics != nil && s.config.EnablePprof {
		diagHandler := s.diagnostics.Handler()
		mux.Handle("/debug/", diagHandler)
	}

	// Root path returns endpoint list
	mux.HandleFunc("/", s.indexHandler)

	s.httpServer = &http.Server{
		Addr:         s.config.Address,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Observability server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the observability server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if s.checker == nil {
		http.Error(w, "Health checker not configured", http.StatusNotImplemented)
		return
	}
	
	response := s.checker.Check(r.Context())

	w.Header().Set("Content-Type", "application/json")
	if response.Status == health.StatusHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	if s.checker == nil {
		http.Error(w, "Health checker not configured", http.StatusNotImplemented)
		return
	}

	response := s.checker.Check(r.Context())

	w.Header().Set("Content-Type", "application/json")
	if response.Status == health.StatusHealthy {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		reason := "not ready"
		if len(response.Checks) > 0 {
			for _, check := range response.Checks {
				if check.Status == health.StatusUnhealthy {
					reason = check.Message
					break
				}
			}
		}
		json.NewEncoder(w).Encode(map[string]string{
			"status": "not ready",
			"reason": reason,
		})
	}
}

func (s *Server) liveHandler(w http.ResponseWriter, r *http.Request) {
	if s.checker == nil {
		http.Error(w, "Health checker not configured", http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "alive",
		"uptime": s.checker.Uptime().String(),
	})
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	endpoints := map[string]string{
		"metrics": s.config.MetricsPath,
	}

	if s.checker != nil {
		endpoints["health"] = "/health"
		endpoints["ready"] = "/ready"
		endpoints["live"] = "/live"
	}

	if s.diagnostics != nil && s.config.EnablePprof {
		endpoints["diagnostic"] = "/debug/diagnostic"
		endpoints["pprof"] = "/debug/pprof/"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":   "guocedb",
		"endpoints": endpoints,
	})
}
