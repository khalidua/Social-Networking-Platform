package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"social-networking-platform/posts-service/internal/apiresponse"
	"social-networking-platform/posts-service/internal/apperrors"
	"social-networking-platform/posts-service/internal/domain"
	"social-networking-platform/posts-service/internal/middleware"
	"social-networking-platform/posts-service/internal/service"
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

func NewPostHandler(svc service.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	userID := middleware.GetAuthenticatedUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, rid, apperrors.CodeUnauthenticated, "missing authenticated user")
		return
	}

	var body createPostBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, rid, apperrors.CodeBadRequest, "invalid JSON body")
		return
	}

	post, err := h.svc.CreatePost(r.Context(), userID, body.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			writeError(w, http.StatusBadRequest, rid, apperrors.CodeValidationError, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, rid, apperrors.CodeInternalError, err.Error())
		}
		return
	}

	writeSuccess(w, http.StatusCreated, rid, post, "")
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	postID, ok := postIDFromPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, rid, apperrors.CodeBadRequest, "invalid post id in path")
		return
	}

	post, err := h.svc.GetPost(r.Context(), postID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, rid, apperrors.CodeInternalError, err.Error())
		return
	}
	if post == nil {
		writeError(w, http.StatusNotFound, rid, apperrors.CodeNotFound, "post not found")
		return
	}

	writeSuccess(w, http.StatusOK, rid, post, "")
}

func (h *PostHandler) ListPostsByAuthor(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	authorID := strings.TrimSpace(r.URL.Query().Get("authorId"))
	if authorID == "" {
		writeError(w, http.StatusBadRequest, rid, apperrors.CodeBadRequest, "missing authorId query parameter")
		return
	}

	posts, err := h.svc.ListPostsByAuthor(r.Context(), authorID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, rid, apperrors.CodeInternalError, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, rid, posts, "")
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	userID := middleware.GetAuthenticatedUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, rid, apperrors.CodeUnauthenticated, "missing authenticated user")
		return
	}

	postID, ok := postIDFromPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, rid, apperrors.CodeBadRequest, "invalid post id in path")
		return
	}

	var body updatePostBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, rid, apperrors.CodeBadRequest, "invalid JSON body")
		return
	}

	post, err := h.svc.UpdatePost(r.Context(), userID, postID, body.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			writeError(w, http.StatusBadRequest, rid, apperrors.CodeValidationError, err.Error())
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, rid, apperrors.CodeForbidden, err.Error())
		case errors.Is(err, service.ErrPostNotFound):
			writeError(w, http.StatusNotFound, rid, apperrors.CodeNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, rid, apperrors.CodeInternalError, err.Error())
		}
		return
	}

	writeSuccess(w, http.StatusOK, rid, post, "")
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	userID := middleware.GetAuthenticatedUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, rid, apperrors.CodeUnauthenticated, "missing authenticated user")
		return
	}

	postID, ok := postIDFromPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, rid, apperrors.CodeBadRequest, "invalid post id in path")
		return
	}

	err := h.svc.DeletePost(r.Context(), userID, postID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			writeError(w, http.StatusForbidden, rid, apperrors.CodeForbidden, err.Error())
		case errors.Is(err, service.ErrPostNotFound):
			writeError(w, http.StatusNotFound, rid, apperrors.CodeNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, rid, apperrors.CodeInternalError, err.Error())
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
	apiresponse.Success(w, status, requestID, data, message)
}

func writeError(w http.ResponseWriter, status int, requestID string, code string, message string) {
	_ = requestID
	apiresponse.Error(w, status, code, message)
}
