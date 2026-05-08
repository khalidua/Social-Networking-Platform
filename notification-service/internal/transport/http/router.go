package httptransport

import (
	"net/http"

	handlers "social-networking-platform/notification-service/internal/handler/http"
	"social-networking-platform/notification-service/internal/middleware"
)

func NewRouter(serviceName string) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(serviceName)
	featureHandler := handlers.NewNotificationHandler()

	mux.HandleFunc("/health", healthHandler.Health)

	mux.HandleFunc("/api/v1/notifications", featureHandler.GetNotifications)

	return middleware.RequestID(
		middleware.Logging(serviceName)(
			middleware.Recovery(mux),
		),
	)
}
