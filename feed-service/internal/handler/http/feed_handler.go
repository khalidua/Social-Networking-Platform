package handlers

import (
	"net/http"
	"social-networking-platform/feed-service/internal/service"
)

type FeedHandler struct{
	service service.FeedService
}

func NewFeedHandler(
	service service.FeedService,
) *FeedHandler {
	return &FeedHandler{
		service: service,
	}
}

func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error": map[string]any{
				"code":    "METHOD_NOT_ALLOWED",
				"message": "only GET is supported",
			},
		})
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"success": false,
			"error": map[string]any{
				"code":    "UNAUTHORIZED",
				"message": "missing user id",
			},
		})
		return
	}

	feed, err := h.service.GetFeed(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error": map[string]any{
				"code":    "FEED_ERROR",
				"message": err.Error(),
			},
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    feed,
	})
}