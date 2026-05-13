package httptransport

import (
	"net/http"

	handlers "social-networking-platform/notification-service/internal/handler/http"
	"social-networking-platform/notification-service/internal/middleware"
	"social-networking-platform/notification-service/internal/service"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(serviceName string, notificationService service.NotificationService) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(serviceName)
	featureHandler := handlers.NewNotificationHandler(notificationService)

	mux.HandleFunc("/health", healthHandler.Health)
	// mux.Handle("/metrics", middleware.MetricsHandler(serviceName))
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/api/v1/notifications", featureHandler.GetNotifications)

	return middleware.RequestID(
		middleware.Tracing(
			middleware.Logging(serviceName)(
				middleware.Metrics(serviceName)(
					middleware.Recovery(mux),
				),
			),
		),
	)
}
