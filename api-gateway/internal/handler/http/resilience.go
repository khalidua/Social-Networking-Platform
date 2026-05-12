package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"social-networking-platform/api-gateway/internal/middleware"
)

var errCircuitOpen = errors.New("circuit breaker is open")

type retryConfig struct {
	attempts int
	backoff  time.Duration
}

type circuitConfig struct {
	failures int
	openFor  time.Duration
}

type circuitBreaker struct {
	mu                sync.Mutex
	service           string
	failureThreshold  int
	openFor           time.Duration
	consecutiveFailed int
	state             string
	openedUntil       time.Time
}

func newCircuitBreaker(service string, cfg circuitConfig) *circuitBreaker {
	if cfg.failures <= 0 {
		cfg.failures = 5
	}
	if cfg.openFor <= 0 {
		cfg.openFor = 30 * time.Second
	}
	return &circuitBreaker{
		service:          service,
		failureThreshold: cfg.failures,
		openFor:          cfg.openFor,
		state:            "closed",
	}
}

func (b *circuitBreaker) allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state != "open" {
		return true
	}
	if now.Before(b.openedUntil) {
		return false
	}
	b.state = "half_open"
	return true
}

func (b *circuitBreaker) recordSuccess(r *http.Request) {
	b.mu.Lock()
	changed := b.state != "closed" || b.consecutiveFailed != 0
	b.state = "closed"
	b.consecutiveFailed = 0
	b.openedUntil = time.Time{}
	b.mu.Unlock()
	if changed {
		logCircuitBreakerTransition(r, b.service, "closed")
	}
}

func (b *circuitBreaker) recordFailure(r *http.Request) {
	b.mu.Lock()
	b.consecutiveFailed++
	opened := false
	if b.state == "half_open" || b.consecutiveFailed >= b.failureThreshold {
		b.state = "open"
		b.openedUntil = time.Now().UTC().Add(b.openFor)
		opened = true
	}
	b.mu.Unlock()
	if opened {
		logCircuitBreakerTransition(r, b.service, "open")
	}
}

type resilientTransport struct {
	base            http.RoundTripper
	upstreamService string
	retry           retryConfig
	breaker         *circuitBreaker
}

func (t *resilientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.breaker != nil && !t.breaker.allow(time.Now().UTC()) {
		logRetryAttempt(req, t.upstreamService, 0, errCircuitOpen)
		return nil, errCircuitOpen
	}

	attempts := t.retry.attempts
	if attempts <= 0 {
		attempts = 1
	}
	backoff := t.retry.backoff
	if backoff <= 0 {
		backoff = 100 * time.Millisecond
	}
	transport := t.base
	if transport == nil {
		transport = http.DefaultTransport
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				if t.breaker != nil {
					t.breaker.recordFailure(req)
				}
				return nil, err
			}
			req.Body = body
		}
		resp, err := transport.RoundTrip(req)
		if err == nil && !shouldRetryResponse(resp) {
			if t.breaker != nil {
				t.breaker.recordSuccess(req)
			}
			return resp, nil
		}
		if err != nil {
			lastErr = err
		}

		if attempt == attempts || !canReplay(req) {
			if t.breaker != nil {
				t.breaker.recordFailure(req)
			}
			if err != nil {
				return nil, err
			}
			return resp, nil
		}

		if resp != nil && resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		logRetryAttempt(req, t.upstreamService, attempt, err)
		if sleepErr := sleepWithContext(req.Context(), backoffForAttempt(backoff, attempt)); sleepErr != nil {
			if t.breaker != nil {
				t.breaker.recordFailure(req)
			}
			return nil, sleepErr
		}
	}

	if t.breaker != nil {
		t.breaker.recordFailure(req)
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("upstream retry attempts exhausted")
}

func shouldRetryResponse(resp *http.Response) bool {
	if resp == nil {
		return false
	}
	return resp.StatusCode == http.StatusBadGateway ||
		resp.StatusCode == http.StatusServiceUnavailable ||
		resp.StatusCode == http.StatusGatewayTimeout
}

func canReplay(req *http.Request) bool {
	if req.Body == nil || req.Body == http.NoBody {
		return true
	}
	if req.GetBody != nil {
		return true
	}
	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func backoffForAttempt(base time.Duration, attempt int) time.Duration {
	if attempt <= 1 {
		return base
	}
	return base * time.Duration(1<<(attempt-1))
}

func logRetryAttempt(r *http.Request, upstreamService string, attempt int, err error) {
	entry := map[string]interface{}{
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"level":            "WARN",
		"event":            "upstream_retry",
		"service":          "api-gateway",
		"request_id":       middleware.GetRequestID(r.Context()),
		"correlation_id":   middleware.GetCorrelationID(r.Context()),
		"trace_id":         middleware.GetTraceID(r.Context()),
		"span_id":          middleware.GetSpanID(r.Context()),
		"method":           r.Method,
		"path":             r.URL.Path,
		"upstream_service": upstreamService,
		"attempt":          attempt,
	}
	if err != nil {
		entry["error"] = err.Error()
	}
	_ = json.NewEncoder(os.Stdout).Encode(entry)
}

func logCircuitBreakerTransition(r *http.Request, upstreamService string, state string) {
	entry := map[string]interface{}{
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"level":            "WARN",
		"event":            "circuit_breaker_transition",
		"service":          "api-gateway",
		"request_id":       middleware.GetRequestID(r.Context()),
		"correlation_id":   middleware.GetCorrelationID(r.Context()),
		"trace_id":         middleware.GetTraceID(r.Context()),
		"span_id":          middleware.GetSpanID(r.Context()),
		"method":           r.Method,
		"path":             r.URL.Path,
		"upstream_service": upstreamService,
		"state":            state,
	}
	_ = json.NewEncoder(os.Stdout).Encode(entry)
}
