package handlers

import (
	"errors"
	"net/http"
	"strings"

	"social-networking-platform/notification-service/internal/middleware"
	"social-networking-platform/notification-service/internal/service"
)

const headerXUserID = "X-User-ID"

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, requestID, "BAD_REQUEST", "method not supported for this route", nil)
		return
	}

	userID := strings.TrimSpace(r.Header.Get(headerXUserID))
	if userID == "" {
		writeError(w, http.StatusUnauthorized, requestID, "UNAUTHENTICATED", "missing X-User-ID", nil)
		return
	}

	notifications, err := h.svc.GetNotifications(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, requestID, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		writeError(w, http.StatusInternalServerError, requestID, "INTERNAL_ERROR", "failed to get notifications", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":    true,
		"data":       notifications,
		"request_id": requestID,
	})
}

func writeError(w http.ResponseWriter, status int, requestID string, code string, message string, details any) {
	writeJSON(w, status, map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    code,
			"message": message,
			"details": details,
		},
		"request_id": requestID,
	})
}
