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
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "req-123"))
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
	}, security.NewTokenVerifier("secret", "auth-service"), &fakeSessionStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "req-123"))
	req.Header.Set("Authorization", "Bearer "+mustSignTestToken(t, "secret", "auth-service", "google:123", "session-1", "user@example.com"))
	recorder := httptest.NewRecorder()

	handler.ProxyUsers(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status code %d", recorder.Code)
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
