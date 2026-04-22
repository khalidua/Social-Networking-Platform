package handlers

import (
	"net/http"

	"social-networking-platform/users-service/internal/middleware"
	"social-networking-platform/users-service/internal/apiresponse"
)

type HealthHandler struct {
	ServiceName string
}

func NewHealthHandler(serviceName string) *HealthHandler {
	return &HealthHandler{ServiceName: serviceName}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	apiresponse.Success(
		w,
		http.StatusOK,
		middleware.GetRequestID(r.Context()),
		map[string]interface{}{
			"status":  "ok",
			"service": h.ServiceName,
		},
		"health check passed",
	)
}