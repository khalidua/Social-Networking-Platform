package middleware

import (
	"net"
	"net/http"
	"strings"

	"social-networking-platform/api-gateway/internal/apiresponse"
	"social-networking-platform/api-gateway/internal/apperrors"
)

const (
	ForwardedForHeader   = "X-Forwarded-For"
	ForwardedHostHeader  = "X-Forwarded-Host"
	ForwardedProtoHeader = "X-Forwarded-Proto"
	ForwardedPortHeader  = "X-Forwarded-Port"
	RealIPHeader         = "X-Real-IP"
)

func ProxyHeaders(trustProxyHeaders bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !trustProxyHeaders {
				next.ServeHTTP(w, r)
				return
			}

			if proto := firstHeaderValue(r.Header.Get(ForwardedProtoHeader)); proto != "" {
				r.Header.Set(ForwardedProtoHeader, strings.ToLower(proto))
			}

			if host := firstHeaderValue(r.Header.Get(ForwardedHostHeader)); host != "" {
				r.Host = host
			}

			if realIP := firstHeaderValue(r.Header.Get(RealIPHeader)); realIP != "" {
				r.RemoteAddr = normalizeRemoteAddr(realIP, r.RemoteAddr)
			} else if forwardedFor := firstHeaderValue(r.Header.Get(ForwardedForHeader)); forwardedFor != "" {
				r.RemoteAddr = normalizeRemoteAddr(forwardedFor, r.RemoteAddr)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireHTTPS(requireHTTPS bool, trustProxyHeaders bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !requireHTTPS {
				next.ServeHTTP(w, r)
				return
			}

			if r.TLS != nil {
				next.ServeHTTP(w, r)
				return
			}

			if trustProxyHeaders {
				proto := strings.ToLower(firstHeaderValue(r.Header.Get(ForwardedProtoHeader)))
				if proto == "https" {
					next.ServeHTTP(w, r)
					return
				}
			}

			apiresponse.Error(
				w,
				http.StatusForbidden,
				GetRequestID(r.Context()),
				apperrors.CodeForbidden,
				"https is required",
				nil,
			)
		})
	}
}

func firstHeaderValue(value string) string {
	if value == "" {
		return ""
	}

	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return strings.TrimSpace(value)
	}

	return strings.TrimSpace(parts[0])
}

func normalizeRemoteAddr(ip string, fallback string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return fallback
	}

	if host, port, err := net.SplitHostPort(fallback); err == nil {
		if parsed := net.ParseIP(ip); parsed != nil {
			return net.JoinHostPort(ip, port)
		}
		return net.JoinHostPort(host, port)
	}

	return ip
}