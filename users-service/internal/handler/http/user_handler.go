package handlers

import (
	"net/http"

	"social-networking-platform/users-service/internal/middleware"
	"social-networking-platform/users-service/internal/apperrors"
	"social-networking-platform/users-service/internal/apiresponse"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r, "get current user not implemented yet")
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r, "update current user not implemented yet")
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r, "get user by id not implemented yet")
}

func (h *UserHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r, "follow user not implemented yet")
}

func (h *UserHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r, "unfollow user not implemented yet")
}

func notImplemented(w http.ResponseWriter, r *http.Request, message string) {
	apiresponse.Error(
		w,
		http.StatusNotImplemented,
		middleware.GetRequestID(r.Context()),
		apperrors.CodeNotImplemented,
		message,
		nil,
	)
}