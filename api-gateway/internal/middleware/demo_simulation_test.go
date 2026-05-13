package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDemoSimulationDisabledPassesThrough(t *testing.T) {
	nextCalled := false
	handler := DemoSimulation(DemoSimulationOptions{
		Enabled:     false,
		PathPrefix:  "/api/v1/feed",
		FailureRate: 1,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/feed", nil))

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestDemoSimulationFailureReturnsEnvelope(t *testing.T) {
	handler := RequestID(DemoSimulation(DemoSimulationOptions{
		Enabled:     true,
		PathPrefix:  "/api/v1/feed",
		FailureRate: 1,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/feed", nil)
	req.Header.Set(RequestIDHeader, "req-test")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusServiceUnavailable)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"success":false`) || !strings.Contains(body, `"request_id":"req-test"`) {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func TestDemoSimulationLatencyAppliesToMatchingPath(t *testing.T) {
	delay := 5 * time.Millisecond
	handler := DemoSimulation(DemoSimulationOptions{
		Enabled:     true,
		PathPrefix:  "/api/v1/feed",
		Latency:     delay,
		FailureRate: 0,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	started := time.Now()
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/feed", nil))

	if elapsed := time.Since(started); elapsed < delay {
		t.Fatalf("latency elapsed = %s, want at least %s", elapsed, delay)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
