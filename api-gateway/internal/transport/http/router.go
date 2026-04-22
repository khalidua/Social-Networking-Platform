package httptransport

import (
	"net/http"

	handlers "social-networking-platform/api-gateway/internal/handler/http"
	"social-networking-platform/api-gateway/internal/middleware"
)

func NewRouter(serviceName string) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(serviceName)
	proxyHandler := handlers.NewProxyHandler()

	mux.HandleFunc("/health", healthHandler.Health)

	mux.HandleFunc("/api/v1/auth/", proxyHandler.ProxyAuth)
	mux.HandleFunc("/api/v1/users/", proxyHandler.ProxyUsers)
	mux.HandleFunc("/api/v1/posts/", proxyHandler.ProxyPosts)
	mux.HandleFunc("/api/v1/feed", proxyHandler.ProxyFeed)
	mux.HandleFunc("/api/v1/notifications", proxyHandler.ProxyNotifications)

	return middleware.RequestID(
		middleware.Logging(serviceName)(
			middleware.Recovery(serviceName)(mux),
		),
	)
}