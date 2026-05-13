package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"social-networking-platform/auth-service/internal/domain"
	"social-networking-platform/auth-service/internal/provider"
	"social-networking-platform/auth-service/internal/security"
)

type stubOAuthProvider struct {
	authURL       string
	tokenResponse *provider.TokenResponse
	user          *domain.AuthUser
	authErr       error
	exchangeErr   error
	userErr       error
}

func (s *stubOAuthProvider) AuthCodeURL(state string) (string, error) {
	if s.authErr != nil {
		return "", s.authErr
	}
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
	sessions  map[string]domain.Session
	saveErr   error
	getErr    error
	deleteErr error
}

func (r *memorySessionRepository) Save(ctx context.Context, session domain.Session) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.sessions[session.ID] = session
	return nil
}

func (r *memorySessionRepository) GetByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	session, ok := r.sessions[sessionID]
	if !ok {
		return nil, nil
	}
	return &session, nil
}

func (r *memorySessionRepository) DeleteByID(ctx context.Context, sessionID string) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.sessions, sessionID)
	return nil
}

type failingStateManager struct{}

func (failingStateManager) Generate() (string, error) {
	return "", errors.New("state generation failed")
}

func (failingStateManager) Validate(state string) error {
	return nil
}

type fakeTokenManager struct {
	claims   *security.TokenClaims
	issueErr error
	parseErr error
}

func (m *fakeTokenManager) Issue(user domain.AuthUser, session domain.Session) (string, error) {
	if m.issueErr != nil {
		return "", m.issueErr
	}
	return "signed-token", nil
}

func (m *fakeTokenManager) Parse(token string) (*security.TokenClaims, error) {
	if m.parseErr != nil {
		return nil, m.parseErr
	}
	return m.claims, nil
}

func TestAuthServiceBeginLogin(t *testing.T) {
	stateManager := security.NewStateManager("secret", 10*time.Minute)
	service := NewAuthService(
		&stubOAuthProvider{authURL: "https://accounts.google.com/o/oauth2/v2/auth"},
		stateManager,
		&fakeTokenManager{},
		&memorySessionRepository{sessions: map[string]domain.Session{}},
		time.Hour,
	)

	redirectURL, err := service.BeginLogin(context.Background())
	if err != nil {
		t.Fatalf("BeginLogin returned error: %v", err)
	}
	if redirectURL == "" {
		t.Fatal("expected redirect URL")
	}
}

func TestAuthServiceBeginLoginMapsStateAndProviderErrors(t *testing.T) {
	service := NewAuthService(
		&stubOAuthProvider{},
		failingStateManager{},
		&fakeTokenManager{},
		&memorySessionRepository{sessions: map[string]domain.Session{}},
		time.Hour,
	)
	if _, err := service.BeginLogin(context.Background()); AsServiceError(err).Code != "INTERNAL_ERROR" {
		t.Fatalf("expected state error to map to INTERNAL_ERROR, got %v", err)
	}

	stateManager := security.NewStateManager("secret", 10*time.Minute)
	service = NewAuthService(
		&stubOAuthProvider{authErr: errors.New("not configured")},
		stateManager,
		&fakeTokenManager{},
		&memorySessionRepository{sessions: map[string]domain.Session{}},
		time.Hour,
	)
	if _, err := service.BeginLogin(context.Background()); AsServiceError(err).Code != "INTERNAL_ERROR" {
		t.Fatalf("expected provider error to map to INTERNAL_ERROR, got %v", err)
	}
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

func TestAuthServiceHandleCallbackValidationErrors(t *testing.T) {
	stateManager := security.NewStateManager("secret", 10*time.Minute)
	service := NewAuthService(
		&stubOAuthProvider{},
		stateManager,
		&fakeTokenManager{},
		&memorySessionRepository{sessions: map[string]domain.Session{}},
		time.Hour,
	)

	if _, err := service.HandleCallback(context.Background(), " ", "state"); AsServiceError(err).Code != "BAD_REQUEST" {
		t.Fatalf("expected missing code BAD_REQUEST, got %v", err)
	}
	if _, err := service.HandleCallback(context.Background(), "code", " "); AsServiceError(err).Code != "BAD_REQUEST" {
		t.Fatalf("expected missing state BAD_REQUEST, got %v", err)
	}
	if _, err := service.HandleCallback(context.Background(), "code", "invalid-state"); AsServiceError(err).Code != "BAD_REQUEST" {
		t.Fatalf("expected invalid state BAD_REQUEST, got %v", err)
	}
}

func TestAuthServiceHandleCallbackMapsFetchTokenAndSaveErrors(t *testing.T) {
	stateManager := security.NewStateManager("secret", 10*time.Minute)
	state, err := stateManager.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	service := NewAuthService(
		&stubOAuthProvider{
			tokenResponse: &provider.TokenResponse{AccessToken: "google-access-token"},
			userErr:       errors.New("userinfo down"),
		},
		stateManager,
		&fakeTokenManager{},
		&memorySessionRepository{sessions: map[string]domain.Session{}},
		time.Hour,
	)
	if _, err := service.HandleCallback(context.Background(), "code", state); AsServiceError(err).Code != "UPSTREAM_UNAVAILABLE" {
		t.Fatalf("expected FetchUser error to map to UPSTREAM_UNAVAILABLE, got %v", err)
	}

	state, _ = stateManager.Generate()
	service = NewAuthService(
		&stubOAuthProvider{
			tokenResponse: &provider.TokenResponse{AccessToken: "google-access-token"},
			user:          &domain.AuthUser{ID: "google:123", Email: "user@example.com"},
		},
		stateManager,
		&fakeTokenManager{issueErr: errors.New("sign failed")},
		&memorySessionRepository{sessions: map[string]domain.Session{}},
		time.Hour,
	)
	if _, err := service.HandleCallback(context.Background(), "code", state); AsServiceError(err).Code != "INTERNAL_ERROR" {
		t.Fatalf("expected token issue error to map to INTERNAL_ERROR, got %v", err)
	}

	state, _ = stateManager.Generate()
	service = NewAuthService(
		&stubOAuthProvider{
			tokenResponse: &provider.TokenResponse{AccessToken: "google-access-token"},
			user:          &domain.AuthUser{ID: "google:123", Email: "user@example.com"},
		},
		stateManager,
		&fakeTokenManager{},
		&memorySessionRepository{sessions: map[string]domain.Session{}, saveErr: errors.New("redis down")},
		time.Hour,
	)
	if _, err := service.HandleCallback(context.Background(), "code", state); AsServiceError(err).Code != "INTERNAL_ERROR" {
		t.Fatalf("expected session save error to map to INTERNAL_ERROR, got %v", err)
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

func TestAuthServiceLogoutMapsDeleteError(t *testing.T) {
	tokenManager := &fakeTokenManager{
		claims: &security.TokenClaims{
			Subject:   "google:123",
			SessionID: "session-1",
		},
	}
	service := NewAuthService(
		&stubOAuthProvider{},
		security.NewStateManager("secret", 10*time.Minute),
		tokenManager,
		&memorySessionRepository{sessions: map[string]domain.Session{}, deleteErr: errors.New("redis down")},
		time.Hour,
	)

	err := service.Logout(context.Background(), "Bearer token")
	if AsServiceError(err).Code != "INTERNAL_ERROR" {
		t.Fatalf("expected delete error to map to INTERNAL_ERROR, got %v", err)
	}
}

func TestAuthServiceValidateSessionReturnsClaimsBackedUser(t *testing.T) {
	expiresAt := time.Now().UTC().Add(time.Hour)
	session := domain.Session{
		ID:        "session-1",
		UserID:    "google:123",
		Email:     "user@example.com",
		ExpiresAt: expiresAt,
	}
	service := NewAuthService(
		&stubOAuthProvider{},
		security.NewStateManager("secret", 10*time.Minute),
		&fakeTokenManager{
			claims: &security.TokenClaims{
				Subject:   "google:123",
				Email:     "user@example.com",
				Name:      "Example User",
				Picture:   "https://example.com/avatar.png",
				SessionID: "session-1",
			},
		},
		&memorySessionRepository{sessions: map[string]domain.Session{"session-1": session}},
		time.Hour,
	)

	result, err := service.ValidateSession(context.Background(), "Bearer token")
	if err != nil {
		t.Fatalf("ValidateSession returned error: %v", err)
	}
	if result.User.ID != "google:123" || result.SessionID != "session-1" || !result.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("unexpected validation result: %+v", result)
	}
}

func TestAuthServiceValidateSessionRejectsInvalidOrExpiredSessions(t *testing.T) {
	tokenManager := &fakeTokenManager{
		claims: &security.TokenClaims{
			Subject:   "google:123",
			SessionID: "session-1",
		},
	}
	service := NewAuthService(
		&stubOAuthProvider{},
		security.NewStateManager("secret", 10*time.Minute),
		tokenManager,
		&memorySessionRepository{sessions: map[string]domain.Session{}},
		time.Hour,
	)
	if _, err := service.ValidateSession(context.Background(), "Bearer token"); AsServiceError(err).Code != "UNAUTHENTICATED" {
		t.Fatalf("expected revoked session UNAUTHENTICATED, got %v", err)
	}

	service = NewAuthService(
		&stubOAuthProvider{},
		security.NewStateManager("secret", 10*time.Minute),
		tokenManager,
		&memorySessionRepository{sessions: map[string]domain.Session{
			"session-1": {
				ID:        "session-1",
				UserID:    "other-user",
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			},
		}},
		time.Hour,
	)
	if _, err := service.ValidateSession(context.Background(), "Bearer token"); AsServiceError(err).Code != "UNAUTHENTICATED" {
		t.Fatalf("expected subject mismatch UNAUTHENTICATED, got %v", err)
	}

	service = NewAuthService(
		&stubOAuthProvider{},
		security.NewStateManager("secret", 10*time.Minute),
		tokenManager,
		&memorySessionRepository{sessions: map[string]domain.Session{
			"session-1": {
				ID:        "session-1",
				UserID:    "google:123",
				ExpiresAt: time.Now().UTC().Add(-time.Minute),
			},
		}},
		time.Hour,
	)
	if _, err := service.ValidateSession(context.Background(), "Bearer token"); AsServiceError(err).Code != "UNAUTHENTICATED" {
		t.Fatalf("expected expired session UNAUTHENTICATED, got %v", err)
	}
}

func TestAuthServiceParseBearerTokenErrors(t *testing.T) {
	service := NewAuthService(
		&stubOAuthProvider{},
		security.NewStateManager("secret", 10*time.Minute),
		&fakeTokenManager{parseErr: errors.New("bad signature")},
		&memorySessionRepository{sessions: map[string]domain.Session{}},
		time.Hour,
	)

	if _, err := service.ValidateSession(context.Background(), " "); AsServiceError(err).Code != "UNAUTHENTICATED" {
		t.Fatalf("expected missing bearer token UNAUTHENTICATED, got %v", err)
	}
	if _, err := service.ValidateSession(context.Background(), "Bearer bad-token"); AsServiceError(err).Code != "UNAUTHENTICATED" {
		t.Fatalf("expected invalid bearer token UNAUTHENTICATED, got %v", err)
	}
}

func TestAuthServiceValidateSessionMapsRepositoryError(t *testing.T) {
	service := NewAuthService(
		&stubOAuthProvider{},
		security.NewStateManager("secret", 10*time.Minute),
		&fakeTokenManager{
			claims: &security.TokenClaims{Subject: "google:123", SessionID: "session-1"},
		},
		&memorySessionRepository{sessions: map[string]domain.Session{}, getErr: errors.New("redis down")},
		time.Hour,
	)

	if _, err := service.ValidateSession(context.Background(), "Bearer token"); AsServiceError(err).Code != "INTERNAL_ERROR" {
		t.Fatalf("expected session repository error to map to INTERNAL_ERROR, got %v", err)
	}
}

func TestAuthServiceHandleCallbackRecordsBusinessMetrics(t *testing.T) {
	stateManager := security.NewStateManager("secret", 10*time.Minute)
	successService := NewAuthService(
		&stubOAuthProvider{
			tokenResponse: &provider.TokenResponse{AccessToken: "google-access-token"},
			user: &domain.AuthUser{
				ID:    "google:123",
				Email: "user@example.com",
			},
		},
		stateManager,
		&fakeTokenManager{},
		&memorySessionRepository{sessions: map[string]domain.Session{}},
		time.Hour,
	)

	successState, err := stateManager.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	successBefore := testutil.ToFloat64(
		businessOperationTotal.WithLabelValues(statusSuccess),
	)
	durationBefore := histogramSampleCount(
		t,
		"auth_service_business_operation_duration_seconds",
		map[string]string{"operation": operationAuthenticateUser},
	)

	if _, err := successService.HandleCallback(context.Background(), "google-code", successState); err != nil {
		t.Fatalf("HandleCallback() error = %v", err)
	}

	successAfter := testutil.ToFloat64(
		businessOperationTotal.WithLabelValues(statusSuccess),
	)
	if successAfter-successBefore != 1 {
		t.Fatalf("expected success counter delta 1, got %v", successAfter-successBefore)
	}

	durationAfter := histogramSampleCount(
		t,
		"auth_service_business_operation_duration_seconds",
		map[string]string{"operation": operationAuthenticateUser},
	)
	if durationAfter-durationBefore != 1 {
		t.Fatalf("expected histogram count delta 1 after success, got %d", durationAfter-durationBefore)
	}

	failureService := NewAuthService(
		&stubOAuthProvider{exchangeErr: errors.New("google down")},
		stateManager,
		&fakeTokenManager{},
		&memorySessionRepository{sessions: map[string]domain.Session{}},
		time.Hour,
	)

	failureState, err := stateManager.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	failureBefore := testutil.ToFloat64(
		businessOperationTotal.WithLabelValues(statusFailure),
	)
	durationBefore = histogramSampleCount(
		t,
		"auth_service_business_operation_duration_seconds",
		map[string]string{"operation": operationAuthenticateUser},
	)

	if _, err := failureService.HandleCallback(context.Background(), "google-code", failureState); err == nil {
		t.Fatal("expected callback error")
	}

	failureAfter := testutil.ToFloat64(
		businessOperationTotal.WithLabelValues(statusFailure),
	)
	if failureAfter-failureBefore != 1 {
		t.Fatalf("expected failure counter delta 1, got %v", failureAfter-failureBefore)
	}

	durationAfter = histogramSampleCount(
		t,
		"auth_service_business_operation_duration_seconds",
		map[string]string{"operation": operationAuthenticateUser},
	)
	if durationAfter-durationBefore != 1 {
		t.Fatalf("expected histogram count delta 1 after failure, got %d", durationAfter-durationBefore)
	}
}

func histogramSampleCount(t *testing.T, metricName string, labels map[string]string) uint64 {
	t.Helper()

	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}

	for _, family := range metricFamilies {
		if family.GetName() != metricName {
			continue
		}
		for _, metric := range family.GetMetric() {
			if labelsMatch(metric, labels) {
				histogram := metric.GetHistogram()
				if histogram == nil {
					t.Fatalf("metric %s is not a histogram", metricName)
				}
				return histogram.GetSampleCount()
			}
		}
	}

	return 0
}

func labelsMatch(metric *dto.Metric, labels map[string]string) bool {
	if len(metric.GetLabel()) != len(labels) {
		return false
	}

	for _, label := range metric.GetLabel() {
		if labels[label.GetName()] != label.GetValue() {
			return false
		}
	}

	return true
}
