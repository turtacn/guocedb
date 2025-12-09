package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Status represents health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// CheckFunc is a function that performs a health check
type CheckFunc func(ctx context.Context) error

// CheckResult represents the result of a health check
type CheckResult struct {
	Name     string        `json:"name"`
	Status   Status        `json:"status"`
	Message  string        `json:"message,omitempty"`
	Duration time.Duration `json:"duration_ms"`
}

// HealthResponse represents the overall health status
type HealthResponse struct {
	Status    Status         `json:"status"`
	Timestamp time.Time      `json:"timestamp"`
	Checks    []CheckResult  `json:"checks"`
	Version   string         `json:"version,omitempty"`
	Uptime    time.Duration  `json:"uptime_ms"`
}

// Checker manages health checks
type Checker struct {
	checks    map[string]CheckFunc
	mu        sync.RWMutex
	startTime time.Time
	timeout   time.Duration
	version   string
}

// NewChecker creates a new Checker
func NewChecker() *Checker {
	return &Checker{
		checks:    make(map[string]CheckFunc),
		startTime: time.Now(),
		timeout:   5 * time.Second,
	}
}

// SetVersion sets the version string
func (c *Checker) SetVersion(version string) {
	c.version = version
}

// SetTimeout sets the timeout for health checks
func (c *Checker) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// AddCheck adds a health check
func (c *Checker) AddCheck(name string, check CheckFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// RemoveCheck removes a health check
func (c *Checker) RemoveCheck(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.checks, name)
}

// Check runs all health checks and returns the result
func (c *Checker) Check(ctx context.Context) *HealthResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	response := &HealthResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Checks:    make([]CheckResult, 0, len(c.checks)),
		Version:   c.version,
		Uptime:    time.Since(c.startTime),
	}

	var wg sync.WaitGroup
	resultChan := make(chan CheckResult, len(c.checks))

	for name, check := range c.checks {
		wg.Add(1)
		go func(name string, check CheckFunc) {
			defer wg.Done()
			start := time.Now()
			err := check(ctx)

			result := CheckResult{
				Name:     name,
				Duration: time.Since(start),
			}

			if err != nil {
				result.Status = StatusUnhealthy
				result.Message = err.Error()
			} else {
				result.Status = StatusHealthy
			}

			resultChan <- result
		}(name, check)
	}

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		response.Checks = append(response.Checks, result)
		if result.Status == StatusUnhealthy {
			response.Status = StatusUnhealthy
		}
	}

	return response
}

// IsHealthy returns whether the system is healthy
func (c *Checker) IsHealthy(ctx context.Context) bool {
	return c.Check(ctx).Status == StatusHealthy
}

// Uptime returns how long the service has been running
func (c *Checker) Uptime() time.Duration {
	return time.Since(c.startTime)
}

// StorageHealthCheck creates a health check for storage
func StorageHealthCheck(storage interface {
	Get([]byte) ([]byte, error)
	Set([]byte, []byte) error
}) CheckFunc {
	return func(ctx context.Context) error {
		testKey := []byte("__health_check__")
		testVal := []byte("ok")

		// Try to write
		if err := storage.Set(testKey, testVal); err != nil {
			return fmt.Errorf("storage write failed: %w", err)
		}

		// Try to read
		if _, err := storage.Get(testKey); err != nil {
			return fmt.Errorf("storage read failed: %w", err)
		}

		return nil
	}
}

// AlwaysHealthyCheck creates a check that always returns healthy (for testing)
func AlwaysHealthyCheck() CheckFunc {
	return func(ctx context.Context) error {
		return nil
	}
}
