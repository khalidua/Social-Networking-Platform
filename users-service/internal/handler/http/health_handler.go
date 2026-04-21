package handlers

import (
    "encoding/json"
    "net/http"
)

type HealthHandler struct {
    ServiceName string
}

func NewHealthHandler(serviceName string) *HealthHandler {
    return &HealthHandler{ServiceName: serviceName}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]any{
        "status":  "ok",
        "service": h.ServiceName,
    })
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(payload)
}
