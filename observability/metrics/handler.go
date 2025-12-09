package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler returns Prometheus metrics HTTP handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// HandlerWithAuth returns metrics handler with basic authentication
func HandlerWithAuth(username, password string) http.Handler {
	return basicAuth(promhttp.Handler(), username, password)
}

func basicAuth(next http.Handler, user, pass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != user || p != pass {
			w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
