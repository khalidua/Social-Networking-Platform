package handlers

import "net/http"

type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    "NOT_IMPLEMENTED",
			"message": "GetNotifications is not implemented yet",
		},
	})
}
