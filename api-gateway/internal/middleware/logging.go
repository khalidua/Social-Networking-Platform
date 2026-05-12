package middleware

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(payload []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(payload)
	rw.bytes += n
	return n, err
}

func Logging(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			statusGroup := statusClass(rw.statusCode)

			entry := map[string]interface{}{
				"timestamp":      time.Now().UTC().Format(time.RFC3339),
				"level":          "INFO",
				"event":          "http_access",
				"service":        serviceName,
				"request_id":     GetRequestID(r.Context()),
				"correlation_id": GetCorrelationID(r.Context()),
				"trace_id":       GetTraceID(r.Context()),
				"span_id":        GetSpanID(r.Context()),
				"method":         r.Method,
				"path":           r.URL.Path,
				"route_group":    routeGroup(r.URL.Path),
				"status":         rw.statusCode,
				"status_group":   statusGroup,
				"duration_ms":    time.Since(started).Milliseconds(),
				"response_bytes": rw.bytes,
				"remote_addr":    r.RemoteAddr,
				"user_agent":     r.UserAgent(),
			}

			if userID := r.Header.Get("X-User-ID"); userID != "" {
				entry["user_id"] = userID
			}

			if upstream := r.Header.Get("X-Upstream-Service"); upstream != "" {
				entry["upstream_service"] = upstream
			}

			_ = json.NewEncoder(os.Stdout).Encode(entry)
		})
	}
}

func statusClass(status int) string {
	if status <= 0 {
		return "unknown"
	}
	return strconv.Itoa(status/100) + "xx"
}

func levelForStatus(status int) string {
	switch {
	case status >= 500:
		return "ERROR"
	case status >= 400:
		return "WARN"
	default:
		return "INFO"
	}
}

func routeGroup(path string) string {
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
