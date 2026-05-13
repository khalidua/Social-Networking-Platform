package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Define Prometheus vectors
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests.",
		},
		[]string{"service", "method", "route", "status", "status_group"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration histogram.",
			Buckets: []float64{0.1, 0.3, 0.5, 1, 2, 5}, // spec buckets
		},
		[]string{"service", "method", "route", "status", "status_group"},
	)

	httpRequestsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_active",
			Help: "Active HTTP requests.",
		},
		[]string{"service"},
	)

	serviceOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_operations_total",
			Help: "Total service operations grouped by route and status.",
		},
		[]string{"service", "method", "route", "status", "status_group"},
	)
)

// Metrics middleware – same signature, now uses Prometheus
func Metrics(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Increment active gauge
			httpRequestsActive.WithLabelValues(serviceName).Inc()
			start := time.Now()

			// Wrap response writer to capture status code
			rw := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			defer func() {
				// Decrement active gauge
				httpRequestsActive.WithLabelValues(serviceName).Dec()

				route := metricRoute(r.URL.Path)
				statusCode := rw.statusCode
				statusGroupVal := statusGroup(statusCode)
				statusStr := strconv.Itoa(statusCode)

				// Increment counters and observe histogram
				httpRequestsTotal.WithLabelValues(serviceName, r.Method, route, statusStr, statusGroupVal).Inc()
				serviceOperationsTotal.WithLabelValues(serviceName, r.Method, route, statusStr, statusGroupVal).Inc()
				httpRequestDuration.WithLabelValues(serviceName, r.Method, route, statusStr, statusGroupVal).Observe(time.Since(start).Seconds())
			}()

			next.ServeHTTP(rw, r)
		})
	}
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// statusGroup returns "2xx", "4xx", "5xx", or "other"
func statusGroup(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "other"
	}
}

// metricRoute extracts a low-cardinality route pattern
func metricRoute(path string) string {
	switch {
	case path == "/health":
		return "health"
	case strings.HasPrefix(path, "/metrics"):
		return "metrics"
	case strings.HasPrefix(path, "/api/v1/auth"):
		return "auth"
	case strings.HasPrefix(path, "/api/v1/users"):
		return "users"
	case strings.HasPrefix(path, "/api/v1/posts"):
		return "posts"
	case strings.HasPrefix(path, "/api/v1/feed"):
		return "feed"
	case strings.HasPrefix(path, "/api/v1/notifications"):
		return "notifications"
	default:
		return "unknown"
	}
}