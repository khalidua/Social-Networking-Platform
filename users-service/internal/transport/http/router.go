package httptransport

import (
    "net/http"

    handlers "social-networking-platform/users-service/internal/handler/http"
    "social-networking-platform/users-service/internal/middleware"
)

func NewRouter(serviceName string) http.Handler {
    mux := http.NewServeMux()

    healthHandler := handlers.NewHealthHandler(serviceName)
    featureHandler := handlers.NewUserHandler()

    mux.HandleFunc("/health", healthHandler.Health)

    mux.HandleFunc("/api/v1/users/me", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            featureHandler.GetMe(w, r)
        case http.MethodPatch:
            featureHandler.UpdateMe(w, r)
        default:
            http.NotFound(w, r)
        }
    })
    mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
        if len(r.URL.Path) >= len("/api/v1/users/") && r.URL.Path != "/api/v1/users/" {
            if r.Method == http.MethodGet {
                featureHandler.GetByID(w, r)
                return
            }
            if r.Method == http.MethodPost && len(r.URL.Path) >= len("/follow") && r.URL.Path[len(r.URL.Path)-7:] == "/follow" {
                featureHandler.FollowUser(w, r)
                return
            }
            if r.Method == http.MethodDelete && len(r.URL.Path) >= len("/follow") && r.URL.Path[len(r.URL.Path)-7:] == "/follow" {
                featureHandler.UnfollowUser(w, r)
                return
            }
        }
        http.NotFound(w, r)
    })


    return middleware.RequestID(
        middleware.Logging(serviceName)(
            middleware.Recovery(mux),
        ),
    )
}
