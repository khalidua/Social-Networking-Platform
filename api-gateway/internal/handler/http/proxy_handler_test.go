package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-networking-platform/api-gateway/internal/config"
	"social-networking-platform/api-gateway/internal/domain"
	"social-networking-platform/api-gateway/internal/middleware"
	"social-networking-platform/api-gateway/internal/security"
)

type fakeSessionStore struct {
	session *domain.Session
	err     error
}

func (s *fakeSessionStore) GetByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	return s.session, s.err
}

func TestProxyUsersForwardsAuthenticatedRequest(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-User-ID") != "google:123" {
			t.Fatalf("expected forwarded X-User-ID header, got %q", r.Header.Get("X-User-ID"))
		}
		if r.Header.Get(middleware.TraceParentHeader) != "00-4bf92f3577b34da6a3ce929d0e0e4736-1111111111111111-01" {
			t.Fatalf("expected forwarded traceparent header, got %q", r.Header.Get(middleware.TraceParentHeader))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	handler := NewProxyHandler(config.Config{
		UsersServiceURL: upstream.URL,
	}, security.NewTokenVerifier("secret", "auth-service"), &fakeSessionStore{
		session: &domain.Session{
			ID:        "session-1",
			UserID:    "google:123",
			Email:     "user@example.com",
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		},
	}, middleware.NewUserRateLimiter(100, time.Minute)) // 100 requests per minute for each user

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req-123")
	ctx = context.WithValue(ctx, middleware.TraceParentKey, "00-4bf92f3577b34da6a3ce929d0e0e4736-1111111111111111-01")
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+mustSignTestToken(t, "secret", "auth-service", "google:123", "session-1", "user@example.com"))
	recorder := httptest.NewRecorder()

	handler.ProxyUsers(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code %d", recorder.Code)
	}
}

func TestProxyUsersRejectsRevokedSession(t *testing.T) {
	handler := NewProxyHandler(config.Config{
		UsersServiceURL: "http://users-service",
	}, security.NewTokenVerifier("secret", "auth-service"), &fakeSessionStore{}, middleware.NewUserRateLimiter(100, time.Minute))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "req-123"))
	req.Header.Set("Authorization", "Bearer "+mustSignTestToken(t, "secret", "auth-service", "google:123", "session-1", "user@example.com"))
	recorder := httptest.NewRecorder()

	handler.ProxyUsers(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status code %d", recorder.Code)
	}
}

func TestProxyUsersRejectsRateLimitedUser(t *testing.T) {
	upstreamCalls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	handler := NewProxyHandler(config.Config{
		UsersServiceURL: upstream.URL,
		UpstreamTimeout: time.Second,
	}, security.NewTokenVerifier("secret", "auth-service"), &fakeSessionStore{
		session: &domain.Session{
			ID:        "session-1",
			UserID:    "google:123",
			Email:     "user@example.com",
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		},
	}, middleware.NewUserRateLimiter(1, time.Minute))

	token := mustSignTestToken(t, "secret", "auth-service", "google:123", "session-1", "user@example.com")

	firstReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	firstReq = firstReq.WithContext(context.WithValue(firstReq.Context(), middleware.RequestIDKey, "req-1"))
	firstReq.Header.Set("Authorization", "Bearer "+token)
	firstRecorder := httptest.NewRecorder()

	handler.ProxyUsers(firstRecorder, firstReq)

	if firstRecorder.Code != http.StatusOK {
		t.Fatalf("first request expected 200, got %d", firstRecorder.Code)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	secondReq = secondReq.WithContext(context.WithValue(secondReq.Context(), middleware.RequestIDKey, "req-2"))
	secondReq.Header.Set("Authorization", "Bearer "+token)
	secondRecorder := httptest.NewRecorder()

	handler.ProxyUsers(secondRecorder, secondReq)

	if secondRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("second request expected 429, got %d body=%s", secondRecorder.Code, secondRecorder.Body.String())
	}

	if upstreamCalls != 1 {
		t.Fatalf("expected only one upstream call, got %d", upstreamCalls)
	}

	if secondRecorder.Header().Get("Retry-After") == "" {
		t.Fatalf("expected Retry-After header")
	}

	if secondRecorder.Header().Get("X-RateLimit-Limit") != "1" {
		t.Fatalf("expected X-RateLimit-Limit=1, got %q", secondRecorder.Header().Get("X-RateLimit-Limit"))
	}
}

func TestProxyAuthRetriesTransientUpstreamFailure(t *testing.T) {
	upstreamCalls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls++
		if upstreamCalls < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"temporary"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	handler := NewProxyHandler(config.Config{
		AuthServiceURL:        upstream.URL,
		UpstreamTimeout:       time.Second,
		UpstreamRetryAttempts: 3,
		UpstreamRetryBackoff:  time.Nanosecond,
	}, security.NewTokenVerifier("secret", "auth-service"), &fakeSessionStore{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "req-retry"))
	recorder := httptest.NewRecorder()

	handler.ProxyAuth(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected final status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if upstreamCalls != 3 {
		t.Fatalf("upstreamCalls = %d, want 3", upstreamCalls)
	}
}

func TestProxyAuthCircuitBreakerOpensAfterFailures(t *testing.T) {
	upstreamCalls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls++
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"down"}`))
	}))
	defer upstream.Close()

	handler := NewProxyHandler(config.Config{
		AuthServiceURL:         upstream.URL,
		UpstreamTimeout:        time.Second,
		UpstreamRetryAttempts:  1,
		UpstreamRetryBackoff:   time.Nanosecond,
		CircuitBreakerFailures: 1,
		CircuitBreakerOpenFor:  time.Minute,
	}, security.NewTokenVerifier("secret", "auth-service"), &fakeSessionStore{}, nil)

	firstReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	firstReq = firstReq.WithContext(context.WithValue(firstReq.Context(), middleware.RequestIDKey, "req-breaker-1"))
	firstRecorder := httptest.NewRecorder()
	handler.ProxyAuth(firstRecorder, firstReq)

	if firstRecorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("first request expected upstream 503, got %d body=%s", firstRecorder.Code, firstRecorder.Body.String())
	}
	if upstreamCalls != 1 {
		t.Fatalf("upstreamCalls after first request = %d, want 1", upstreamCalls)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	secondReq = secondReq.WithContext(context.WithValue(secondReq.Context(), middleware.RequestIDKey, "req-breaker-2"))
	secondRecorder := httptest.NewRecorder()
	handler.ProxyAuth(secondRecorder, secondReq)

	if secondRecorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("second request expected gateway 503, got %d body=%s", secondRecorder.Code, secondRecorder.Body.String())
	}
	if upstreamCalls != 1 {
		t.Fatalf("circuit breaker should prevent second upstream call, got %d calls", upstreamCalls)
	}
	if secondRecorder.Header().Get("X-Upstream-Service") != "auth-service" {
		t.Fatalf("expected X-Upstream-Service auth-service, got %q", secondRecorder.Header().Get("X-Upstream-Service"))
	}
	if secondRecorder.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header while circuit is open")
	}
}

func mustSignTestToken(t *testing.T, secret string, issuer string, subject string, sessionID string, email string) string {
	t.Helper()

	header, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}

	payload, err := json.Marshal(map[string]interface{}{
		"iss":   issuer,
		"sub":   subject,
		"sid":   sessionID,
		"email": email,
		"iat":   time.Now().UTC().Unix(),
		"exp":   time.Now().UTC().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	headerSegment := base64.RawURLEncoding.EncodeToString(header)
	payloadSegment := base64.RawURLEncoding.EncodeToString(payload)
	signingInput := headerSegment + "." + payloadSegment

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature
}
