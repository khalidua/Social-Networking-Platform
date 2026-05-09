package middleware

import (
	"context"
	"net/http"
	"strings"
)

type userContextKey string

const (
	AuthenticatedUserIDKey    userContextKey = "authenticated_user_id"
	AuthenticatedUserIDHeader                = "X-User-ID"
)

func AuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := strings.TrimSpace(r.Header.Get(AuthenticatedUserIDHeader))
		if userID == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), AuthenticatedUserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetAuthenticatedUserID(ctx context.Context) string {
	value, ok := ctx.Value(AuthenticatedUserIDKey).(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}
