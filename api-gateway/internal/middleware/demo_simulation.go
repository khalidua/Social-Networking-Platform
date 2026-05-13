package middleware

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	"social-networking-platform/api-gateway/internal/apiresponse"
	"social-networking-platform/api-gateway/internal/apperrors"
)

type DemoSimulationOptions struct {
	Enabled     bool
	PathPrefix  string
	Latency     time.Duration
	FailureRate float64
}

func DemoSimulation(opts DemoSimulationOptions) func(http.Handler) http.Handler {
	pathPrefix := strings.TrimSpace(opts.PathPrefix)
	return func(next http.Handler) http.Handler {
		if !opts.Enabled || pathPrefix == "" {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, pathPrefix) {
				next.ServeHTTP(w, r)
				return
			}
			if opts.Latency > 0 {
				time.Sleep(opts.Latency)
			}
			if opts.FailureRate >= 1 || (opts.FailureRate > 0 && rand.Float64() < opts.FailureRate) {
				apiresponse.Error(
					w,
					http.StatusServiceUnavailable,
					GetRequestID(r.Context()),
					apperrors.CodeUpstreamUnavailable,
					"demo failure simulation is enabled",
					map[string]any{"path": pathPrefix},
				)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
