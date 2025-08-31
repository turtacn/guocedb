// Package status provides health checking and status reporting for guocedb.
package status

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/turtacn/guocedb/common/types/enum"
)

// HealthStatus represents the health of a component.
type HealthStatus struct {
	Component string          `json:"component"`
	Status    enum.SystemStatus `json:"status"`
	Message   string          `json:"message,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// Checker is an interface for a component that can be health-checked.
type Checker interface {
	Check(ctx context.Context) HealthStatus
}

// Manager aggregates health statuses from multiple components.
type Manager struct {
	mu       sync.RWMutex
	checkers map[string]Checker
}

// NewManager creates a new status manager.
func NewManager() *Manager {
	return &Manager{
		checkers: make(map[string]Checker),
	}
}

// Register adds a new component to be checked.
func (m *Manager) Register(name string, checker Checker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers[name] = checker
}

// GetSystemStatus returns the aggregated health status of the system.
func (m *Manager) GetSystemStatus() []HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var statuses []HealthStatus
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for name, checker := range m.checkers {
		status := checker.Check(ctx)
		status.Component = name
		statuses = append(statuses, status)
	}
	return statuses
}

// ServeHealthCheckEndpoint starts an HTTP server for health checks.
// This is a simplified example. A production system might use a library
// like github.com/heptiolabs/healthcheck.
func (m *Manager) ServeHealthCheckEndpoint(addr string) {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		statuses := m.GetSystemStatus()
		isHealthy := true
		for _, s := range statuses {
			if s.Status != enum.Running {
				isHealthy = false
				break
			}
		}

		if isHealthy {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("ERROR"))
			// In a real system, you would write the JSON of statuses here.
		}
	})
	go http.ListenAndServe(addr, nil)
}

// The following are placeholders for more advanced features.

func (m *Manager) SubscribeToStatusChanges() (chan HealthStatus, error) {
	return nil, nil
}

func (m *Manager) GetStatusHistory(component string) ([]HealthStatus, error) {
	return nil, nil
}
