package middleware

import (
    "context"
    "fmt"
    "math/rand"
    "net/http"
    "time"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = generateRequestID()
        }

        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
        w.Header().Set("X-Request-ID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func GetRequestID(ctx context.Context) string {
    if v, ok := ctx.Value(RequestIDKey).(string); ok {
        return v
    }
    return ""
}

func generateRequestID() string {
    rand.New(rand.NewSource(time.Now().UnixNano()))
    return fmt.Sprintf("req-%d-%06d", time.Now().UnixNano(), rand.Intn(1000000))
}
