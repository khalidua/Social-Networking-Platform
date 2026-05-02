package bootstrap

import (
	"fmt"
	"net/http"

	"social-networking-platform/api-gateway/internal/config"
	handlers "social-networking-platform/api-gateway/internal/handler/http"
	"social-networking-platform/api-gateway/internal/middleware"
	redisrepo "social-networking-platform/api-gateway/internal/repository/redis"
	"social-networking-platform/api-gateway/internal/security"
	httptransport "social-networking-platform/api-gateway/internal/transport/http"
)

type App struct {
	Router http.Handler
}

func NewApp(cfg config.Config) (*App, error) {
	tokenVerifier := security.NewTokenVerifier(cfg.JWTSecret, cfg.JWTIssuer)

	redisClient := redisrepo.NewClient(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, cfg.RedisDB, cfg.UpstreamTimeout)

	sessionRepository := redisrepo.NewSessionRepository(redisClient)

	rateLimiter := middleware.NewUserRateLimiter(
		cfg.RateLimitPerMinute,
		cfg.RateLimitWindow,
	)

	proxyHandler := handlers.NewProxyHandler(cfg, tokenVerifier, sessionRepository, rateLimiter)

	router := httptransport.NewRouter(cfg.ServiceName, proxyHandler)
	if router == nil {
		return nil, fmt.Errorf("failed to initialize router")
	}
	return &App{Router: router}, nil
}
