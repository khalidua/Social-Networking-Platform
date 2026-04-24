package handlers

import (
	"net/http"

	"social-networking-platform/auth-service/internal/apiresponse"
	"social-networking-platform/auth-service/internal/middleware"
	"social-networking-platform/auth-service/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	redirectURL, err := h.authService.BeginLogin(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	result, err := h.authService.HandleCallback(
		r.Context(),
		r.URL.Query().Get("code"),
		r.URL.Query().Get("state"),
	)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	apiresponse.Success(
		w,
		http.StatusOK,
		middleware.GetRequestID(r.Context()),
		result,
		"Google login completed successfully",
	)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.authService.Logout(r.Context(), r.Header.Get("Authorization")); err != nil {
		writeServiceError(w, r, err)
		return
	}

	apiresponse.Success(
		w,
		http.StatusOK,
		middleware.GetRequestID(r.Context()),
		map[string]string{"status": "logged_out"},
		"session invalidated successfully",
	)
}

func (h *AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	result, err := h.authService.ValidateSession(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	apiresponse.Success(
		w,
		http.StatusOK,
		middleware.GetRequestID(r.Context()),
		result,
		"session is valid",
	)
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	serviceErr := service.AsServiceError(err)
	apiresponse.Error(
		w,
		serviceErr.Status,
		middleware.GetRequestID(r.Context()),
		serviceErr.Code,
		serviceErr.Message,
		serviceErr.Details,
	)
}
