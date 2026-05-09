package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"social-networking-platform/posts-service/internal/domain"
	"social-networking-platform/posts-service/internal/middleware"
	"social-networking-platform/posts-service/internal/service"
)

const (
	codeBadRequest      = "BAD_REQUEST"
	codeUnauthenticated = "UNAUTHENTICATED"
	codeForbidden       = "FORBIDDEN"
	codeNotFound        = "NOT_FOUND"
	codeInternalError   = "INTERNAL_ERROR"
)

type postService interface {
	CreatePost(ctx context.Context, authorID string, content string) (*domain.Post, error)
	GetPost(ctx context.Context, id string) (*domain.Post, error)
	ListPostsByAuthor(ctx context.Context, authorID string) ([]domain.Post, error)
	UpdatePost(ctx context.Context, requesterID string, postID string, content string) (*domain.Post, error)
	DeletePost(ctx context.Context, requesterID string, postID string) error
}

type PostHandler struct {
	svc postService
}

type createPostBody struct {
	Content string `json:"content"`
}

type updatePostBody struct {
	Content string `json:"content"`
}

type successEnvelope struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type errorEnvelope struct {
	Success   bool      `json:"success"`
	Error     errorBody `json:"error"`
	RequestID string    `json:"request_id,omitempty"`
}

type errorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func NewPostHandler(svc service.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	userID := middleware.GetAuthenticatedUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, rid, codeUnauthenticated, "missing authenticated user", nil)
		return
	}

	var body createPostBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, rid, codeBadRequest, "invalid JSON body", nil)
		return
	}

	post, err := h.svc.CreatePost(r.Context(), userID, body.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, rid, codeInternalError, err.Error(), nil)
		return
	}

	writeSuccess(w, http.StatusCreated, rid, post, "")
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	postID, ok := postIDFromPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, rid, codeBadRequest, "invalid post id in path", nil)
		return
	}

	post, err := h.svc.GetPost(r.Context(), postID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, rid, codeInternalError, err.Error(), nil)
		return
	}
	if post == nil {
		writeError(w, http.StatusNotFound, rid, codeNotFound, "post not found", nil)
		return
	}

	writeSuccess(w, http.StatusOK, rid, post, "")
}

func (h *PostHandler) ListPostsByAuthor(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	authorID := strings.TrimSpace(r.URL.Query().Get("authorId"))
	if authorID == "" {
		writeError(w, http.StatusBadRequest, rid, codeBadRequest, "missing authorId query parameter", nil)
		return
	}

	posts, err := h.svc.ListPostsByAuthor(r.Context(), authorID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, rid, codeInternalError, err.Error(), nil)
		return
	}

	writeSuccess(w, http.StatusOK, rid, posts, "")
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	userID := middleware.GetAuthenticatedUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, rid, codeUnauthenticated, "missing authenticated user", nil)
		return
	}

	postID, ok := postIDFromPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, rid, codeBadRequest, "invalid post id in path", nil)
		return
	}

	var body updatePostBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, rid, codeBadRequest, "invalid JSON body", nil)
		return
	}

	post, err := h.svc.UpdatePost(r.Context(), userID, postID, body.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, rid, codeForbidden, err.Error(), nil)
		case errors.Is(err, service.ErrPostNotFound):
			writeError(w, http.StatusNotFound, rid, codeNotFound, err.Error(), nil)
		default:
			writeError(w, http.StatusInternalServerError, rid, codeInternalError, err.Error(), nil)
		}
		return
	}

	writeSuccess(w, http.StatusOK, rid, post, "")
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	userID := middleware.GetAuthenticatedUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, rid, codeUnauthenticated, "missing authenticated user", nil)
		return
	}

	postID, ok := postIDFromPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, rid, codeBadRequest, "invalid post id in path", nil)
		return
	}

	err := h.svc.DeletePost(r.Context(), userID, postID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, rid, codeForbidden, err.Error(), nil)
		case errors.Is(err, service.ErrPostNotFound):
			writeError(w, http.StatusNotFound, rid, codeNotFound, err.Error(), nil)
		default:
			writeError(w, http.StatusInternalServerError, rid, codeInternalError, err.Error(), nil)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func postIDFromPath(path string) (string, bool) {
	const prefix = "/api/v1/posts/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if rest == "" || strings.Contains(rest, "/") {
		return "", false
	}
	return rest, true
}

func writeSuccess(w http.ResponseWriter, status int, requestID string, data interface{}, message string) {
	writeJSON(w, status, successEnvelope{
		Success:   true,
		Data:      data,
		Message:   message,
		RequestID: requestID,
	})
}

func writeError(w http.ResponseWriter, status int, requestID string, code string, message string, details interface{}) {
	writeJSON(w, status, errorEnvelope{
		Success: false,
		Error: errorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: requestID,
	})
}
