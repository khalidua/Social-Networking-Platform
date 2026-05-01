package security

import (
	"testing"
	"time"

	"social-networking-platform/auth-service/internal/domain"
)

func TestTokenManagerIssueAndParse(t *testing.T) {
	manager := NewTokenManager("secret", "auth-service", time.Hour)

	token, err := manager.Issue(domain.AuthUser{
		ID:            "google:123",
		Email:         "user@example.com",
		Name:          "Example User",
		ProfilePicURL: "https://example.com/avatar.png",
	}, domain.Session{
		ID:        "session-1",
		UserID:    "google:123",
		Email:     "user@example.com",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if claims.Subject != "google:123" {
		t.Fatalf("unexpected subject %q", claims.Subject)
	}
	if claims.SessionID != "session-1" {
		t.Fatalf("unexpected session id %q", claims.SessionID)
	}
}

func TestTokenManagerRejectsExpiredToken(t *testing.T) {
	manager := NewTokenManager("secret", "auth-service", -time.Minute)

	token, err := manager.Issue(domain.AuthUser{
		ID:    "google:123",
		Email: "user@example.com",
	}, domain.Session{
		ID:        "session-1",
		UserID:    "google:123",
		Email:     "user@example.com",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := manager.Parse(token); err == nil {
		t.Fatal("Parse() expected expired token error")
	}
}
