package httptransport

import (
	"net/http"
	"strings"

	"social-networking-platform/posts-service/internal/apiresponse"
	"social-networking-platform/posts-service/internal/apperrors"
	handlers "social-networking-platform/posts-service/internal/handler/http"
	"social-networking-platform/posts-service/internal/middleware"
)

func NewRouter(serviceName string, postHandler *handlers.PostHandler) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(serviceName)

	mux.HandleFunc("/health", healthHandler.Health)
	mux.Handle("/metrics", middleware.MetricsHandler(serviceName))

	mux.HandleFunc("/api/v1/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			postHandler.CreatePost(w, r)
		case http.MethodGet:
			postHandler.ListPostsByAuthor(w, r)
		default:
			methodNotAllowed(w, r)
		}
	})
	mux.HandleFunc("/api/v1/posts/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/interactions") {
			if r.Method != http.MethodPost {
				methodNotAllowed(w, r)
				return
			}
			postHandler.InteractWithPost(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			postHandler.GetPost(w, r)
		case http.MethodPut:
			postHandler.UpdatePost(w, r)
		case http.MethodDelete:
			postHandler.DeletePost(w, r)
		default:
			methodNotAllowed(w, r)
		}
	})

	return middleware.RequestID(
		middleware.Tracing(
			middleware.AuthenticatedUser(
				middleware.Logging(serviceName)(
					middleware.Metrics(serviceName)(
						middleware.Recovery(mux),
					),
				),
			),
		),
	)
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	apiresponse.Error(w, http.StatusBadRequest, apperrors.CodeBadRequest, "method not supported for this route")
}
