// Package health provides health checking functionality for GuoceDB.
package health

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
	"encoding/json"

	"github.com/dgraph-io/badger/v3"
)

// Status represents the health status of a component.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Checker performs health checks on various system components.
type Checker struct {
	db     *badger.DB
	checks []Check
}

// Check represents a single health check.
type Check struct {
	Name     string
	Fn       func() error
	Critical bool // Critical checks cause unhealthy status when they fail
}

// HealthReport contains the results of all health checks.
type HealthReport struct {
	Status    Status        `json:"status"`
	Checks    []CheckResult `json:"checks"`
	Timestamp time.Time     `json:"timestamp"`
}

// CheckResult contains the result of a single health check.
type CheckResult struct {
	Name     string        `json:"name"`
	Status   Status        `json:"status"`
	Message  string        `json:"message"`
	Duration time.Duration `json:"duration"`
}

// NewChecker creates a new health checker.
func NewChecker(db *badger.DB) *Checker {
	c := &Checker{
		db: db,
	}
	
	// Register default checks
	c.checks = []Check{
		{Name: "storage", Fn: c.checkStorage, Critical: true},
		{Name: "memory", Fn: c.checkMemory, Critical: false},
		{Name: "disk", Fn: c.checkDisk, Critical: false},
	}
	
	return c
}

// Check performs all registered health checks and returns a report.
func (c *Checker) Check() *HealthReport {
	report := &HealthReport{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Checks:    make([]CheckResult, 0, len(c.checks)),
	}
	
	for _, check := range c.checks {
		start := time.Now()
		err := check.Fn()
		result := CheckResult{
			Name:     check.Name,
			Duration: time.Since(start),
		}
		
		if err != nil {
			result.Status = StatusUnhealthy
			result.Message = err.Error()
			if check.Critical {
				report.Status = StatusUnhealthy
			} else if report.Status == StatusHealthy {
				report.Status = StatusDegraded
			}
		} else {
			result.Status = StatusHealthy
			result.Message = "OK"
		}
		
		report.Checks = append(report.Checks, result)
	}
	
	return report
}

// checkStorage verifies that the storage layer is accessible and writable.
func (c *Checker) checkStorage() error {
	if c.db == nil {
		return fmt.Errorf("database is nil")
	}
	
	// Test read/write operation
	err := c.db.Update(func(txn *badger.Txn) error {
		key := []byte("_health_check_")
		val := []byte(time.Now().String())
		if err := txn.Set(key, val); err != nil {
			return err
		}
		return txn.Delete(key)
	})
	
	if err != nil {
		return fmt.Errorf("storage read/write test failed: %w", err)
	}
	
	return nil
}

// checkMemory checks memory usage and warns if it's too high.
func (c *Checker) checkMemory() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Warning threshold: using more than 90% of allocated memory
	usedPercent := float64(m.Alloc) / float64(m.Sys) * 100
	if usedPercent > 90 {
		return fmt.Errorf("memory usage high: %.1f%% (alloc: %d, sys: %d)", 
			usedPercent, m.Alloc, m.Sys)
	}
	
	return nil
}

// checkDisk checks disk space availability.
func (c *Checker) checkDisk() error {
	// For now, this is a placeholder
	// In a real implementation, we would check disk space using syscalls
	// or a third-party library like github.com/shirou/gopsutil
	return nil
}

// HTTPHandler returns an HTTP handler for health checks.
func (c *Checker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := c.Check()
		
		status := http.StatusOK
		if report.Status == StatusUnhealthy {
			status = http.StatusServiceUnavailable
		} else if report.Status == StatusDegraded {
			status = http.StatusOK // Could be 429 Too Many Requests
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		
		if err := json.NewEncoder(w).Encode(report); err != nil {
			http.Error(w, "Failed to encode health report", http.StatusInternalServerError)
		}
	}
}

// AddCheck adds a custom health check.
func (c *Checker) AddCheck(name string, fn func() error, critical bool) {
	c.checks = append(c.checks, Check{
		Name:     name,
		Fn:       fn,
		Critical: critical,
	})
}