package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"social-networking-platform/auth-service/internal/apperrors"
	"social-networking-platform/auth-service/internal/domain"
	"social-networking-platform/auth-service/internal/provider"
	redisrepo "social-networking-platform/auth-service/internal/repository/redis"
	"social-networking-platform/auth-service/internal/security"
)

type AuthService interface {
	BeginLogin(ctx context.Context) (string, error)
	HandleCallback(ctx context.Context, code string, state string) (*CallbackResult, error)
	Logout(ctx context.Context, bearerToken string) error
	ValidateSession(ctx context.Context, bearerToken string) (*ValidationResult, error)
}

type StateManager interface {
	Generate() (string, error)
	Validate(state string) error
}

type TokenManager interface {
	Issue(user domain.AuthUser, session domain.Session) (string, error)
	Parse(token string) (*security.TokenClaims, error)
}

type CallbackResult struct {
	Token     string          `json:"token"`
	ExpiresAt time.Time       `json:"expires_at"`
	User      domain.AuthUser `json:"user"`
	SessionID string          `json:"session_id"`
}

type ValidationResult struct {
	User      domain.AuthUser `json:"user"`
	SessionID string          `json:"session_id"`
	ExpiresAt time.Time       `json:"expires_at"`
}

type ServiceError struct {
	Status  int
	Code    string
	Message string
	Details interface{}
}

func (e *ServiceError) Error() string {
	return e.Message
}

type DefaultAuthService struct {
	provider     provider.OAuthProvider
	stateManager StateManager
	tokenManager TokenManager
	sessions     redisrepo.SessionRepository
	sessionTTL   time.Duration
}

func NewAuthService(
	oauthProvider provider.OAuthProvider,
	stateManager StateManager,
	tokenManager TokenManager,
	sessions redisrepo.SessionRepository,
	sessionTTL time.Duration,
) *DefaultAuthService {
	return &DefaultAuthService{
		provider:     oauthProvider,
		stateManager: stateManager,
		tokenManager: tokenManager,
		sessions:     sessions,
		sessionTTL:   sessionTTL,
	}
}

func (s *DefaultAuthService) BeginLogin(ctx context.Context) (string, error) {
	_ = ctx
	state, err := s.stateManager.Generate()
	if err != nil {
		return "", &ServiceError{
			Status:  http.StatusInternalServerError,
			Code:    apperrors.CodeInternalError,
			Message: "failed to generate login state",
		}
	}
	redirectURL, err := s.provider.AuthCodeURL(state)
	if err != nil {
		return "", &ServiceError{
			Status:  http.StatusInternalServerError,
			Code:    apperrors.CodeInternalError,
			Message: "Google OAuth provider is not configured",
			Details: err.Error(),
		}
	}
	return redirectURL, nil
}

func (s *DefaultAuthService) HandleCallback(ctx context.Context, code string, state string) (*CallbackResult, error) {
	started := time.Now()
	status := statusFailure
	defer func() {
		observeBusinessOperation(operationAuthenticateUser, started, status)
	}()

	if strings.TrimSpace(code) == "" {
		return nil, &ServiceError{
			Status:  http.StatusBadRequest,
			Code:    apperrors.CodeBadRequest,
			Message: "missing authorization code",
		}
	}
	if strings.TrimSpace(state) == "" {
		return nil, &ServiceError{
			Status:  http.StatusBadRequest,
			Code:    apperrors.CodeBadRequest,
			Message: "missing OAuth state",
		}
	}
	if err := s.stateManager.Validate(state); err != nil {
		return nil, &ServiceError{
			Status:  http.StatusBadRequest,
			Code:    apperrors.CodeBadRequest,
			Message: "invalid OAuth state",
			Details: err.Error(),
		}
	}

	tokenResponse, err := s.provider.ExchangeCode(ctx, code)
	if err != nil {
		return nil, &ServiceError{
			Status:  http.StatusBadGateway,
			Code:    apperrors.CodeUpstreamUnavailable,
			Message: "failed to exchange code with Google",
			Details: err.Error(),
		}
	}

	user, err := s.provider.FetchUser(ctx, tokenResponse.AccessToken)
	if err != nil {
		return nil, &ServiceError{
			Status:  http.StatusBadGateway,
			Code:    apperrors.CodeUpstreamUnavailable,
			Message: "failed to load Google user profile",
			Details: err.Error(),
		}
	}

	sessionID, err := newSessionID()
	if err != nil {
		return nil, &ServiceError{
			Status:  http.StatusInternalServerError,
			Code:    apperrors.CodeInternalError,
			Message: "failed to create session",
		}
	}

	expiresAt := time.Now().UTC().Add(s.sessionTTL)
	session := domain.Session{
		ID:        sessionID,
		UserID:    user.ID,
		Email:     user.Email,
		ExpiresAt: expiresAt,
	}

	signedToken, err := s.tokenManager.Issue(*user, session)
	if err != nil {
		return nil, &ServiceError{
			Status:  http.StatusInternalServerError,
			Code:    apperrors.CodeInternalError,
			Message: "failed to sign JWT",
		}
	}

	if err := s.sessions.Save(ctx, session); err != nil {
		return nil, &ServiceError{
			Status:  http.StatusInternalServerError,
			Code:    apperrors.CodeInternalError,
			Message: "failed to persist session",
			Details: err.Error(),
		}
	}

	status = statusSuccess
	return &CallbackResult{
		Token:     signedToken,
		ExpiresAt: expiresAt,
		User:      *user,
		SessionID: sessionID,
	}, nil
}

func (s *DefaultAuthService) Logout(ctx context.Context, bearerToken string) error {
	claims, err := s.parseBearerToken(bearerToken)
	if err != nil {
		return err
	}
	if err := s.sessions.DeleteByID(ctx, claims.SessionID); err != nil {
		return &ServiceError{
			Status:  http.StatusInternalServerError,
			Code:    apperrors.CodeInternalError,
			Message: "failed to invalidate session",
			Details: err.Error(),
		}
	}
	return nil
}

func (s *DefaultAuthService) ValidateSession(ctx context.Context, bearerToken string) (*ValidationResult, error) {
	claims, err := s.parseBearerToken(bearerToken)
	if err != nil {
		return nil, err
	}

	session, err := s.sessions.GetByID(ctx, claims.SessionID)
	if err != nil {
		return nil, &ServiceError{
			Status:  http.StatusInternalServerError,
			Code:    apperrors.CodeInternalError,
			Message: "failed to load session",
			Details: err.Error(),
		}
	}
	if session == nil || session.UserID != claims.Subject {
		return nil, &ServiceError{
			Status:  http.StatusUnauthorized,
			Code:    apperrors.CodeUnauthenticated,
			Message: "session is invalid or revoked",
		}
	}
	if session.ExpiresAt.UTC().Before(time.Now().UTC()) {
		return nil, &ServiceError{
			Status:  http.StatusUnauthorized,
			Code:    apperrors.CodeUnauthenticated,
			Message: "session has expired",
		}
	}

	return &ValidationResult{
		User: domain.AuthUser{
			ID:            claims.Subject,
			Email:         claims.Email,
			Name:          claims.Name,
			ProfilePicURL: claims.Picture,
		},
		SessionID: claims.SessionID,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func (s *DefaultAuthService) parseBearerToken(bearerToken string) (*security.TokenClaims, error) {
	token := strings.TrimSpace(bearerToken)
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	if token == "" {
		return nil, &ServiceError{
			Status:  http.StatusUnauthorized,
			Code:    apperrors.CodeUnauthenticated,
			Message: "missing bearer token",
		}
	}

	claims, err := s.tokenManager.Parse(token)
	if err != nil {
		return nil, &ServiceError{
			Status:  http.StatusUnauthorized,
			Code:    apperrors.CodeUnauthenticated,
			Message: "invalid bearer token",
			Details: err.Error(),
		}
	}
	return claims, nil
}

func AsServiceError(err error) *ServiceError {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr
	}
	return &ServiceError{
		Status:  http.StatusInternalServerError,
		Code:    apperrors.CodeInternalError,
		Message: fmt.Sprintf("unexpected error: %v", err),
	}
}

func newSessionID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
