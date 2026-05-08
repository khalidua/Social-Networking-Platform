package handlers

import "net/http"

type FeedHandler struct{}

func NewFeedHandler() *FeedHandler {
	return &FeedHandler{}
}

func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    "NOT_IMPLEMENTED",
			"message": "GetFeed is not implemented yet",
		},
	})
}
