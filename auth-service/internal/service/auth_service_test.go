package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-networking-platform/auth-service/internal/domain"
	"social-networking-platform/auth-service/internal/provider"
	"social-networking-platform/auth-service/internal/security"
)

type stubOAuthProvider struct {
	authURL       string
	tokenResponse *provider.TokenResponse
	user          *domain.AuthUser
	exchangeErr   error
	userErr       error
}

func (s *stubOAuthProvider) AuthCodeURL(state string) (string, error) {
	return s.authURL + "?state=" + state, nil
}

func (s *stubOAuthProvider) ExchangeCode(ctx context.Context, code string) (*provider.TokenResponse, error) {
	if s.exchangeErr != nil {
		return nil, s.exchangeErr
	}
	return s.tokenResponse, nil
}

func (s *stubOAuthProvider) FetchUser(ctx context.Context, accessToken string) (*domain.AuthUser, error) {
	if s.userErr != nil {
		return nil, s.userErr
	}
	return s.user, nil
}

type memorySessionRepository struct {
	sessions map[string]domain.Session
}

func (r *memorySessionRepository) Save(ctx context.Context, session domain.Session) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *memorySessionRepository) GetByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	session, ok := r.sessions[sessionID]
	if !ok {
		return nil, nil
	}
	return &session, nil
}

func (r *memorySessionRepository) DeleteByID(ctx context.Context, sessionID string) error {
	delete(r.sessions, sessionID)
	return nil
}

func TestAuthServiceHandleCallbackCreatesSessionAndToken(t *testing.T) {
	stateManager := security.NewStateManager("secret", 10*time.Minute)
	tokenManager := security.NewTokenManager("secret", "auth-service", time.Hour)
	providerStub := &stubOAuthProvider{
		authURL: "https://accounts.google.com/o/oauth2/v2/auth",
		tokenResponse: &provider.TokenResponse{
			AccessToken: "google-access-token",
		},
		user: &domain.AuthUser{
			ID:            "google:123",
			Provider:      "google",
			Email:         "user@example.com",
			Name:          "Example User",
			ProfilePicURL: "https://example.com/avatar.png",
		},
	}
	sessionRepo := &memorySessionRepository{sessions: map[string]domain.Session{}}
	service := NewAuthService(providerStub, stateManager, tokenManager, sessionRepo, time.Hour)

	state, err := stateManager.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result, err := service.HandleCallback(context.Background(), "google-code", state)
	if err != nil {
		t.Fatalf("HandleCallback() error = %v", err)
	}

	if result.Token == "" {
		t.Fatal("expected JWT token in callback result")
	}
	if result.SessionID == "" {
		t.Fatal("expected session ID in callback result")
	}
	if _, ok := sessionRepo.sessions[result.SessionID]; !ok {
		t.Fatal("expected session to be persisted")
	}
}

func TestAuthServiceHandleCallbackReturnsUpstreamError(t *testing.T) {
	stateManager := security.NewStateManager("secret", 10*time.Minute)
	tokenManager := security.NewTokenManager("secret", "auth-service", time.Hour)
	service := NewAuthService(
		&stubOAuthProvider{exchangeErr: errors.New("google down")},
		stateManager,
		tokenManager,
		&memorySessionRepository{sessions: map[string]domain.Session{}},
		time.Hour,
	)

	state, err := stateManager.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = service.HandleCallback(context.Background(), "google-code", state)
	if err == nil {
		t.Fatal("expected callback error")
	}

	serviceErr := AsServiceError(err)
	if serviceErr.Code != "UPSTREAM_UNAVAILABLE" {
		t.Fatalf("unexpected error code %q", serviceErr.Code)
	}
}

func TestAuthServiceLogoutInvalidatesSession(t *testing.T) {
	stateManager := security.NewStateManager("secret", 10*time.Minute)
	tokenManager := security.NewTokenManager("secret", "auth-service", time.Hour)
	sessionRepo := &memorySessionRepository{sessions: map[string]domain.Session{}}
	service := NewAuthService(
		&stubOAuthProvider{},
		stateManager,
		tokenManager,
		sessionRepo,
		time.Hour,
	)

	user := domain.AuthUser{
		ID:    "google:123",
		Email: "user@example.com",
	}
	session := domain.Session{
		ID:        "session-1",
		UserID:    user.ID,
		Email:     user.Email,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	if err := sessionRepo.Save(context.Background(), session); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	token, err := tokenManager.Issue(user, session)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if err := service.Logout(context.Background(), "Bearer "+token); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if _, ok := sessionRepo.sessions[session.ID]; ok {
		t.Fatal("expected session to be deleted")
	}
}
