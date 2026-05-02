package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey string

const (
	RequestIDKey    contextKey = "request_id"
	CorrelationIDKey contextKey = "correlation_id"

	RequestIDHeader string     = "X-Request-ID"
	CorrelationIDHeader string = "X-Correlation-ID"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = generateID("req")
		}

		correlationID := r.Header.Get(CorrelationIDHeader)
		if correlationID == "" {
			correlationID = requestID
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, CorrelationIDKey, correlationID)
		
		r = r.WithContext(ctx)

		w.Header().Set(RequestIDHeader, requestID)
		w.Header().Set(CorrelationIDHeader, correlationID)
		
		next.ServeHTTP(w, r)
	})
}

func GetRequestID(ctx context.Context) string {
	value, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return ""
	}
	return value
}

func GetCorrelationID(ctx context.Context) string {
	value, ok := ctx.Value(CorrelationIDKey).(string)
	if !ok {
		return GetRequestID(ctx)
	}
	return value
}

func generateID(prefix string) string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		return prefix + "-fallback"
	}
	return prefix + "-" + hex.EncodeToString(bytes)
}