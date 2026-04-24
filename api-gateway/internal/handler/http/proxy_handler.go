package handlers

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"social-networking-platform/api-gateway/internal/apiresponse"
	"social-networking-platform/api-gateway/internal/apperrors"
	"social-networking-platform/api-gateway/internal/config"
	"social-networking-platform/api-gateway/internal/domain"
	"social-networking-platform/api-gateway/internal/middleware"
	"social-networking-platform/api-gateway/internal/security"
)

type ProxyHandler struct {
	authServiceURL         string
	usersServiceURL        string
	postsServiceURL        string
	feedServiceURL         string
	notificationServiceURL string
	tokenVerifier          *security.TokenVerifier
	sessions               sessionReader
}

type sessionReader interface {
	GetByID(ctx context.Context, sessionID string) (*domain.Session, error)
}

func NewProxyHandler(cfg config.Config, tokenVerifier *security.TokenVerifier, sessions sessionReader) *ProxyHandler {
	return &ProxyHandler{
		authServiceURL:         cfg.AuthServiceURL,
		usersServiceURL:        cfg.UsersServiceURL,
		postsServiceURL:        cfg.PostsServiceURL,
		feedServiceURL:         cfg.FeedServiceURL,
		notificationServiceURL: cfg.NotificationServiceURL,
		tokenVerifier:          tokenVerifier,
		sessions:               sessions,
	}
}

func (h *ProxyHandler) ProxyAuth(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, h.authServiceURL, nil)
}

func (h *ProxyHandler) ProxyUsers(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.requireAuthenticated(w, r)
	if !ok {
		return
	}
	h.proxyRequest(w, r, h.usersServiceURL, claims)
}

func (h *ProxyHandler) ProxyPosts(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.requireAuthenticated(w, r)
	if !ok {
		return
	}
	h.proxyRequest(w, r, h.postsServiceURL, claims)
}

func (h *ProxyHandler) ProxyFeed(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.requireAuthenticated(w, r)
	if !ok {
		return
	}
	h.proxyRequest(w, r, h.feedServiceURL, claims)
}

func (h *ProxyHandler) ProxyNotifications(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.requireAuthenticated(w, r)
	if !ok {
		return
	}
	h.proxyRequest(w, r, h.notificationServiceURL, claims)
}

func (h *ProxyHandler) requireAuthenticated(w http.ResponseWriter, r *http.Request) (*security.TokenClaims, bool) {
	requestID := middleware.GetRequestID(r.Context())
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	if token == "" {
		apiresponse.Error(
			w,
			http.StatusUnauthorized,
			requestID,
			apperrors.CodeUnauthenticated,
			"missing bearer token",
			nil,
		)
		return nil, false
	}

	claims, err := h.tokenVerifier.Parse(token)
	if err != nil {
		apiresponse.Error(
			w,
			http.StatusUnauthorized,
			requestID,
			apperrors.CodeUnauthenticated,
			"invalid bearer token",
			err.Error(),
		)
		return nil, false
	}

	session, err := h.sessions.GetByID(r.Context(), claims.SessionID)
	if err != nil {
		apiresponse.Error(
			w,
			http.StatusInternalServerError,
			requestID,
			apperrors.CodeInternalError,
			"failed to verify session",
			err.Error(),
		)
		return nil, false
	}
	if session == nil || session.UserID != claims.Subject {
		apiresponse.Error(
			w,
			http.StatusUnauthorized,
			requestID,
			apperrors.CodeUnauthenticated,
			"session is invalid or revoked",
			nil,
		)
		return nil, false
	}
	if session.ExpiresAt.UTC().Before(time.Now().UTC()) {
		apiresponse.Error(
			w,
			http.StatusUnauthorized,
			requestID,
			apperrors.CodeUnauthenticated,
			"session has expired",
			nil,
		)
		return nil, false
	}
	return claims, true
}

func (h *ProxyHandler) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string, claims *security.TokenClaims) {
	requestID := middleware.GetRequestID(r.Context())
	target, err := url.Parse(targetURL)
	if err != nil {
		apiresponse.Error(
			w,
			http.StatusInternalServerError,
			requestID,
			apperrors.CodeInternalError,
			"gateway upstream configuration is invalid",
			nil,
		)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = r.URL.Path
		req.URL.RawPath = r.URL.RawPath
		req.URL.RawQuery = r.URL.RawQuery
		req.Host = target.Host
		req.Header.Set(middleware.RequestIDHeader, requestID)
		if claims != nil {
			req.Header.Set("X-User-ID", claims.Subject)
			req.Header.Set("X-User-Email", claims.Email)
			req.Header.Set("X-Session-ID", claims.SessionID)
		}
	}
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, proxyErr error) {
		apiresponse.Error(
			rw,
			http.StatusBadGateway,
			requestID,
			apperrors.CodeUpstreamUnavailable,
			"upstream service is unavailable",
			proxyErr.Error(),
		)
	}

	proxy.ServeHTTP(w, r)
}
