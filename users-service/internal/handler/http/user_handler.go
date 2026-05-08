package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"social-networking-platform/users-service/internal/apiresponse"
	"social-networking-platform/users-service/internal/apperrors"
	"social-networking-platform/users-service/internal/middleware"
	"social-networking-platform/users-service/internal/service"
)

const headerXUserID = "X-User-ID"

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

type updateMeBody struct {
	Name           *string `json:"name"`
	Bio            *string `json:"bio"`
	ProfilePicture *string `json:"profile_picture"`
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	userID := strings.TrimSpace(r.Header.Get(headerXUserID))
	if userID == "" {
		apiresponse.Error(w, http.StatusUnauthorized, rid, apperrors.CodeUnauthenticated, "missing X-User-ID", nil)
		return
	}
	u, err := h.svc.GetMe(r.Context(), userID)
	if err != nil {
		apiresponse.Error(w, http.StatusInternalServerError, rid, apperrors.CodeInternalError, err.Error(), nil)
		return
	}
	apiresponse.Success(w, http.StatusOK, rid, u, "")
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	userID := strings.TrimSpace(r.Header.Get(headerXUserID))
	if userID == "" {
		apiresponse.Error(w, http.StatusUnauthorized, rid, apperrors.CodeUnauthenticated, "missing X-User-ID", nil)
		return
	}

	var body updateMeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiresponse.Error(w, http.StatusBadRequest, rid, apperrors.CodeBadRequest, "invalid JSON body", nil)
		return
	}

	u, err := h.svc.UpdateMe(r.Context(), userID, body.Name, body.Bio, body.ProfilePicture)
	if err != nil {
		status := http.StatusInternalServerError
		code := apperrors.CodeInternalError
		if errors.Is(err, service.ErrValidation) {
			status = http.StatusBadRequest
			code = apperrors.CodeValidationError
		}
		apiresponse.Error(w, status, rid, code, err.Error(), nil)
		return
	}
	apiresponse.Success(w, http.StatusOK, rid, u, "")
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	rid := middleware.GetRequestID(r.Context())
	id, ok := userIDFromResourcePath(r.URL.Path, false)
	if !ok {
		apiresponse.Error(w, http.StatusBadRequest, rid, apperrors.CodeBadRequest, "invalid user id in path", nil)
		return
	}
	u, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		apiresponse.Error(w, http.StatusInternalServerError, rid, apperrors.CodeInternalError, err.Error(), nil)
		return
	}
	if u == nil {
		apiresponse.Error(w, http.StatusNotFound, rid, apperrors.CodeNotFound, "user not found", nil)
		return
	}
	apiresponse.Success(w, http.StatusOK, rid, u, "")
}

func (h *UserHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	h.followMutation(w, r, true)
}

func (h *UserHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	h.followMutation(w, r, false)
}

func (h *UserHandler) followMutation(w http.ResponseWriter, r *http.Request, follow bool) {
	rid := middleware.GetRequestID(r.Context())
	followerID := strings.TrimSpace(r.Header.Get(headerXUserID))
	if followerID == "" {
		apiresponse.Error(w, http.StatusUnauthorized, rid, apperrors.CodeUnauthenticated, "missing X-User-ID", nil)
		return
	}
	followeeID, ok := userIDFromResourcePath(r.URL.Path, true)
	if !ok {
		apiresponse.Error(w, http.StatusBadRequest, rid, apperrors.CodeBadRequest, "invalid user id in path", nil)
		return
	}

	var err error
	if follow {
		err = h.svc.FollowUser(r.Context(), followerID, followeeID)
	} else {
		err = h.svc.UnfollowUser(r.Context(), followerID, followeeID)
	}
	if err != nil {
		status := http.StatusBadRequest
		code := apperrors.CodeBadRequest
		if errors.Is(err, service.ErrCannotFollowSelf) {
			status = http.StatusForbidden
			code = apperrors.CodeForbidden
		}
		apiresponse.Error(w, status, rid, code, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// userIDFromResourcePath parses /api/v1/users/{id} or /api/v1/users/{id}/follow (when stripFollowSuffix).
func userIDFromResourcePath(path string, stripFollowSuffix bool) (string, bool) {
	const prefix = "/api/v1/users/"
	if stripFollowSuffix {
		const suf = "/follow"
		if !strings.HasSuffix(path, suf) {
			return "", false
		}
		path = strings.TrimSuffix(path, suf)
	}
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if rest == "" || rest == "me" {
		return "", false
	}
	if strings.Contains(rest, "/") {
		return "", false
	}
	return rest, true
}
