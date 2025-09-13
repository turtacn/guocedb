package status

import (
	"context"
	"net/http"
	"sync"
	"time"

	"encoding/json"
)

// HealthStatus represents the health of a component.
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "Healthy"
	StatusUnhealthy HealthStatus = "Unhealthy"
	StatusDegraded  HealthStatus = "Degraded"
)

// Check is a single health check result.
type Check struct {
	Component string       `json:"component"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

// StatusProvider is an interface that components can implement to provide their health status.
type StatusProvider interface {
	ComponentName() string
	HealthCheck(ctx context.Context) Check
}

// Reporter aggregates health status from multiple providers.
type Reporter struct {
	mu        sync.RWMutex
	providers []StatusProvider
}

// NewReporter creates a new status reporter.
func NewReporter() *Reporter {
	return &Reporter{
		providers: make([]StatusProvider, 0),
	}
}

// RegisterProvider adds a new status provider to the reporter.
func (r *Reporter) RegisterProvider(p StatusProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = append(r.providers, p)
}

// Report generates a full system health report.
func (r *Reporter) Report(ctx context.Context) []Check {
	r.mu.RLock()
	defer r.mu.RUnlock()

	checks := make([]Check, len(r.providers))
	var wg sync.WaitGroup
	for i, p := range r.providers {
		wg.Add(1)
		go func(i int, p StatusProvider) {
			defer wg.Done()
			checks[i] = p.HealthCheck(ctx)
		}(i, p)
	}
	wg.Wait()
	return checks
}

// NewHandler creates an HTTP handler for the health endpoint.
func (r *Reporter) NewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		report := r.Report(req.Context())
		isHealthy := true
		for _, check := range report {
			if check.Status != StatusHealthy {
				isHealthy = false
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if !isHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(report)
	})
}
