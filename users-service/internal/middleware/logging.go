package middleware

import (
    "log"
    "net/http"
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
            rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

            next.ServeHTTP(rw, r)

            log.Printf(
                "service=%s method=%s path=%s status=%d duration_ms=%d request_id=%s",
                serviceName,
                r.Method,
                r.URL.Path,
                rw.statusCode,
                time.Since(started).Milliseconds(),
                GetRequestID(r.Context()),
            )
        })
    }
}
