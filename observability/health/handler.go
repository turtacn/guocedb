package health

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler returns HTTP handler for health check endpoints
func (c *Checker) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", c.healthHandler)
	mux.HandleFunc("/ready", c.readyHandler)
	mux.HandleFunc("/live", c.liveHandler)
	return mux
}

func (c *Checker) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := c.Check(r.Context())

	w.Header().Set("Content-Type", "application/json")
	if response.Status == StatusHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(response)
}

func (c *Checker) readyHandler(w http.ResponseWriter, r *http.Request) {
	// Readiness: service is ready to receive traffic
	response := c.Check(r.Context())

	w.Header().Set("Content-Type", "application/json")
	if response.Status == StatusHealthy {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		reason := "not ready"
		if len(response.Checks) > 0 {
			for _, check := range response.Checks {
				if check.Status == StatusUnhealthy {
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

func (c *Checker) liveHandler(w http.ResponseWriter, r *http.Request) {
	// Liveness: process is alive (simple check)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "alive",
		"uptime": time.Since(c.startTime).String(),
	})
}
