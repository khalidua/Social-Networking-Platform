package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	TraceIDKey        contextKey = "trace_id"
	SpanIDKey         contextKey = "span_id"
	TraceParentKey    contextKey = "traceparent"
	TraceParentHeader            = "Traceparent"
)

func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := traceIDFromHeader(r.Header.Get(TraceParentHeader))
		if traceID == "" {
			traceID = randomTraceHex(16)
		}
		spanID := randomTraceHex(8)
		traceparent := "00-" + traceID + "-" + spanID + "-01"

		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
		ctx = context.WithValue(ctx, SpanIDKey, spanID)
		ctx = context.WithValue(ctx, TraceParentKey, traceparent)

		w.Header().Set(TraceParentHeader, traceparent)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetTraceID(ctx context.Context) string {
	value, ok := ctx.Value(TraceIDKey).(string)
	if !ok {
		return ""
	}
	return value
}

func GetSpanID(ctx context.Context) string {
	value, ok := ctx.Value(SpanIDKey).(string)
	if !ok {
		return ""
	}
	return value
}

func GetTraceParent(ctx context.Context) string {
	value, ok := ctx.Value(TraceParentKey).(string)
	if !ok {
		return ""
	}
	return value
}

func traceIDFromHeader(raw string) string {
	parts := strings.Split(strings.TrimSpace(raw), "-")
	if len(parts) != 4 {
		return ""
	}
	traceID := strings.ToLower(parts[1])
	if len(traceID) != 32 || !isHex(traceID) || isAllZero(traceID) {
		return ""
	}
	spanID := strings.ToLower(parts[2])
	if len(spanID) != 16 || !isHex(spanID) || isAllZero(spanID) {
		return ""
	}
	return traceID
}

func randomTraceHex(bytesLen int) string {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return strings.Repeat("1", bytesLen*2)
	}
	return hex.EncodeToString(buf)
}

func isHex(value string) bool {
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

func isAllZero(value string) bool {
	for _, r := range value {
		if r != '0' {
			return false
		}
	}
	return true
}
