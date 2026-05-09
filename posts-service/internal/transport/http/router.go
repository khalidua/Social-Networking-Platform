package httptransport

import (
	"encoding/json"
	"net/http"

	handlers "social-networking-platform/posts-service/internal/handler/http"
	"social-networking-platform/posts-service/internal/middleware"
)

func NewRouter(serviceName string, postHandler *handlers.PostHandler) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(serviceName)

	mux.HandleFunc("/health", healthHandler.Health)

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
		middleware.AuthenticatedUser(
			middleware.Logging(serviceName)(
				middleware.Recovery(mux),
			),
		),
	)
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    "BAD_REQUEST",
			"message": "method not supported for this route",
		},
		"request_id": middleware.GetRequestID(r.Context()),
	})
}
