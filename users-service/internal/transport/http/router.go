package httptransport

import (
	"net/http"
	"strings"

	"social-networking-platform/users-service/internal/apiresponse"
	"social-networking-platform/users-service/internal/apperrors"
	handlers "social-networking-platform/users-service/internal/handler/http"
	"social-networking-platform/users-service/internal/middleware"
)

func NewRouter(serviceName string, userHandler *handlers.UserHandler) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(serviceName)

	mux.HandleFunc("/health", healthHandler.Health)
	mux.Handle("/metrics", middleware.MetricsHandler(serviceName))

	mux.HandleFunc("/api/v1/users/me", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetMe(w, r)
		case http.MethodPatch:
			userHandler.UpdateMe(w, r)
		default:
			methodNotAllowed(w, r)
		}
	})

	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, "/followers") {
			if r.Method != http.MethodGet {
				methodNotAllowed(w, r)
				return
			}
			userHandler.ListFollowers(w, r)
			return
		}

		if strings.HasSuffix(path, "/follow") {
			switch r.Method {
			case http.MethodPost:
				userHandler.FollowUser(w, r)
			case http.MethodDelete:
				userHandler.UnfollowUser(w, r)
			default:
				methodNotAllowed(w, r)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			userHandler.GetByID(w, r)
		default:
			methodNotAllowed(w, r)
		}
	})

	return middleware.RequestID(
		middleware.Logging(serviceName)(
			middleware.Metrics(serviceName)(
				middleware.Recovery(serviceName)(mux),
			),
		),
	)
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	apiresponse.Error(
		w,
		http.StatusBadRequest,
		middleware.GetRequestID(r.Context()),
		apperrors.CodeBadRequest,
		"method not supported for this route",
		nil,
	)
}
