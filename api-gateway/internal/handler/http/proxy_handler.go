package handlers

import (
	"net/http"

	"social-networking-platform/api-gateway/internal/middleware"
	"social-networking-platform/api-gateway/internal/apperrors"
	"social-networking-platform/api-gateway/internal/apiresponse"
)

type ProxyHandler struct{}

func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{}
}

func (h *ProxyHandler) ProxyAuth(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r, "gateway auth proxy not implemented yet")
}

func (h *ProxyHandler) ProxyUsers(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r, "gateway users proxy not implemented yet")
}

func (h *ProxyHandler) ProxyPosts(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r, "gateway posts proxy not implemented yet")
}

func (h *ProxyHandler) ProxyFeed(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r, "gateway feed proxy not implemented yet")
}

func (h *ProxyHandler) ProxyNotifications(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r, "gateway notifications proxy not implemented yet")
}

func notImplemented(w http.ResponseWriter, r *http.Request, message string) {
	apiresponse.Error(
		w,
		http.StatusNotImplemented,
		middleware.GetRequestID(r.Context()),
		apperrors.CodeNotImplemented,
		message,
		nil,
	)
}