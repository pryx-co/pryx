package server

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// MetricsMiddleware tracks performance metrics for HTTP requests.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Use chi's response writer wrapper to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		// Log performance metric
		// In production, this would go to Prometheus/OTLP
		if duration > 500*time.Millisecond {
			log.Printf("⚠️ SLOW REQUEST: %s %s took %v (status: %d)", r.Method, r.URL.Path, duration, ww.Status())
		} else {
			log.Printf("METRIC: %s %s took %v (status: %d)", r.Method, r.URL.Path, duration, ww.Status())
		}
	})
}
