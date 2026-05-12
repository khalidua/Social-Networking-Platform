package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTracingPreservesIncomingTraceIDAndCreatesNewSpan(t *testing.T) {
	incoming := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	var gotTraceID string
	var gotSpanID string
	handler := Tracing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceID = GetTraceID(r.Context())
		gotSpanID = GetSpanID(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(TraceParentHeader, incoming)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotTraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("trace id = %q", gotTraceID)
	}
	if gotSpanID == "" || gotSpanID == "00f067aa0ba902b7" {
		t.Fatalf("expected new span id, got %q", gotSpanID)
	}
	if !strings.Contains(rec.Header().Get(TraceParentHeader), gotTraceID) {
		t.Fatalf("response traceparent missing trace id: %q", rec.Header().Get(TraceParentHeader))
	}
}

func TestTracingGeneratesTraceIDWhenMissing(t *testing.T) {
	var gotTraceID string
	handler := Tracing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceID = GetTraceID(r.Context())
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if len(gotTraceID) != 32 {
		t.Fatalf("generated trace id length = %d, want 32", len(gotTraceID))
	}
	if rec.Header().Get(TraceParentHeader) == "" {
		t.Fatal("expected response traceparent header")
	}
}
