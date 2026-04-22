package middleware

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"social-networking-platform/api-gateway/internal/apperrors"
	"social-networking-platform/api-gateway/internal/apiresponse"
)

func Recovery(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					entry := map[string]interface{}{
						"timestamp":  time.Now().UTC().Format(time.RFC3339),
						"level":      "ERROR",
						"service":    serviceName,
						"request_id": GetRequestID(r.Context()),
						"method":     r.Method,
						"path":       r.URL.Path,
						"panic":      rec,
					}
					_ = json.NewEncoder(os.Stdout).Encode(entry)

					apiresponse.Error(
						w,
						http.StatusInternalServerError,
						GetRequestID(r.Context()),
						apperrors.CodeInternalError,
						"internal server error",
						nil,
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}