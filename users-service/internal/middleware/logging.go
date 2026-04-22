package middleware

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
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

			entry := map[string]interface{}{
				"timestamp":   time.Now().UTC().Format(time.RFC3339),
				"level":       "INFO",
				"service":     serviceName,
				"request_id":  GetRequestID(r.Context()),
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      rw.statusCode,
				"duration_ms": time.Since(started).Milliseconds(),
				"remote_addr": r.RemoteAddr,
			}

			_ = json.NewEncoder(os.Stdout).Encode(entry)
		})
	}
}