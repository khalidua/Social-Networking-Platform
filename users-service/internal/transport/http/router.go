package httptransport

import (
	"net/http"
	"strings"

	handlers "social-networking-platform/users-service/internal/handler/http"
	"social-networking-platform/users-service/internal/middleware"
	"social-networking-platform/users-service/internal/apperrors"
	"social-networking-platform/users-service/internal/apiresponse"
)

func NewRouter(serviceName string) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(serviceName)
	userHandler := handlers.NewUserHandler()

	mux.HandleFunc("/health", healthHandler.Health)

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
			middleware.Recovery(serviceName)(mux),
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