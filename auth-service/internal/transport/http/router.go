package httptransport

import (
	"net/http"

	"social-networking-platform/auth-service/internal/apiresponse"
	"social-networking-platform/auth-service/internal/apperrors"
	handlers "social-networking-platform/auth-service/internal/handler/http"
	"social-networking-platform/auth-service/internal/middleware"
)

func NewRouter(serviceName string, authHandler *handlers.AuthHandler) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(serviceName)

	mux.HandleFunc("/health", healthHandler.Health)
	mux.Handle("/metrics", middleware.MetricsHandler(serviceName))
	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, r)
			return
		}
		authHandler.Login(w, r)
	})
	mux.HandleFunc("/api/v1/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, r)
			return
		}
		authHandler.Callback(w, r)
	})
	mux.HandleFunc("/api/v1/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			methodNotAllowed(w, r)
			return
		}
		authHandler.Logout(w, r)
	})
	mux.HandleFunc("/api/v1/auth/session", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, r)
			return
		}
		authHandler.Session(w, r)
	})

	return middleware.RequestID(
		middleware.Logging(serviceName)(
			middleware.Metrics(serviceName)(
				middleware.Recovery(mux),
			),
		),
	)
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	apiresponse.Error(
		w,
		http.StatusMethodNotAllowed,
		middleware.GetRequestID(r.Context()),
		apperrors.CodeBadRequest,
		"method not supported for this route",
		nil,
	)
}
