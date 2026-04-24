package bootstrap

import (
	"fmt"
	"net/http"

	"social-networking-platform/auth-service/internal/config"
	handlers "social-networking-platform/auth-service/internal/handler/http"
	"social-networking-platform/auth-service/internal/provider"
	redisrepo "social-networking-platform/auth-service/internal/repository/redis"
	"social-networking-platform/auth-service/internal/security"
	"social-networking-platform/auth-service/internal/service"
	httptransport "social-networking-platform/auth-service/internal/transport/http"
)

type App struct {
	Router http.Handler
}

func NewApp(cfg config.Config) (*App, error) {
	httpClient := &http.Client{Timeout: cfg.UpstreamTimeout}
	googleProvider := provider.NewGoogleProvider(provider.GoogleConfig{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		AuthURL:      cfg.GoogleAuthURL,
		TokenURL:     cfg.GoogleTokenURL,
		UserInfoURL:  cfg.GoogleUserInfoURL,
	}, httpClient)

	stateManager := security.NewStateManager(cfg.OAuthStateSecret, cfg.OAuthStateTTL)
	tokenManager := security.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiresIn)
	redisClient := redisrepo.NewClient(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, cfg.RedisDB, cfg.UpstreamTimeout)
	sessionRepository := redisrepo.NewSessionRepository(redisClient)
	authService := service.NewAuthService(googleProvider, stateManager, tokenManager, sessionRepository, cfg.SessionTTL)
	authHandler := handlers.NewAuthHandler(authService)

	router := httptransport.NewRouter(cfg.ServiceName, authHandler)
	if router == nil {
		return nil, fmt.Errorf("failed to initialize router")
	}
	return &App{Router: router}, nil
}
