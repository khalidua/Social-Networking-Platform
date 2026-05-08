package httptransport

import (
	"net/http"

	handlers "social-networking-platform/feed-service/internal/handler/http"
	"social-networking-platform/feed-service/internal/middleware"
)

func NewRouter(serviceName string) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(serviceName)
	featureHandler := handlers.NewFeedHandler()

	mux.HandleFunc("/health", healthHandler.Health)

	mux.HandleFunc("/api/v1/feed", featureHandler.GetFeed)

	return middleware.RequestID(
		middleware.Logging(serviceName)(
			middleware.Recovery(mux),
		),
	)
}
