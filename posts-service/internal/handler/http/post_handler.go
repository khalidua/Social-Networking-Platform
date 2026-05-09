package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"social-networking-platform/posts-service/internal/service"
)

type PostHandler struct {
	svc service.PostService
}

func NewPostHandler(svc service.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

type createPostBody struct {
	Content string `json:"content"`
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error": map[string]any{
				"code":    "METHOD_NOT_ALLOWED",
				"message": "only POST is supported",
			},
		})
		return
	}

	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
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

	var body createPostBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error": map[string]any{
				"code":    "BAD_REQUEST",
				"message": "invalid JSON body",
			},
		})
		return
	}

	post, err := h.svc.CreatePost(r.Context(), userID, body.Content)
	if err != nil {
		status := http.StatusInternalServerError
		code := "INTERNAL_ERROR"
		switch {
		case errors.Is(err, service.ErrEmptyContent),
			errors.Is(err, service.ErrContentTooLong),
			errors.Is(err, service.ErrMissingAuthor):
			status = http.StatusBadRequest
			code = "VALIDATION_ERROR"
			if errors.Is(err, service.ErrMissingAuthor) {
				code = "BAD_REQUEST"
			}
		}
		writeJSON(w, status, map[string]any{
			"success": false,
			"error": map[string]any{
				"code":    code,
				"message": err.Error(),
			},
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"data": map[string]any{
			"id":         post.ID,
			"author_id":  post.AuthorID,
			"content":    post.Content,
			"created_at": post.CreatedAt,
		},
	})
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    "NOT_IMPLEMENTED",
			"message": "GetPost is not implemented yet",
		},
	})
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    "NOT_IMPLEMENTED",
			"message": "UpdatePost is not implemented yet",
		},
	})
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    "NOT_IMPLEMENTED",
			"message": "DeletePost is not implemented yet",
		},
	})
}
