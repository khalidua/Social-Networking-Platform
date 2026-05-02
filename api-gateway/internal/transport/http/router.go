package httptransport

import (
	"net/http"

	"social-networking-platform/api-gateway/internal/config"
	handlers "social-networking-platform/api-gateway/internal/handler/http"
	"social-networking-platform/api-gateway/internal/middleware"
)

func NewRouter(cfg config.Config, proxyHandler *handlers.ProxyHandler) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(cfg.ServiceName)

	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/api/v1/auth/", proxyHandler.ProxyAuth)
	mux.HandleFunc("/api/v1/users/", proxyHandler.ProxyUsers)
	mux.HandleFunc("/api/v1/posts", proxyHandler.ProxyPosts)
	mux.HandleFunc("/api/v1/posts/", proxyHandler.ProxyPosts)
	mux.HandleFunc("/api/v1/feed", proxyHandler.ProxyFeed)
	mux.HandleFunc("/api/v1/notifications", proxyHandler.ProxyNotifications)

	return middleware.RequestID(
		middleware.ProxyHeaders(cfg.TrustProxyHeaders)(
			middleware.RequireHTTPS(cfg.RequireHTTPS, cfg.TrustProxyHeaders)(
				middleware.Logging(cfg.ServiceName)(
					middleware.Recovery(cfg.ServiceName)(mux),
				),
			),
		),
	)
}
