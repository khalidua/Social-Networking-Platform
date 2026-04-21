package handlers

import "net/http"

type UserHandler struct {}

func NewUserHandler() *UserHandler {
    return &UserHandler{}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusNotImplemented, map[string]any{
        "success": false,
        "error": map[string]any{
            "code": "NOT_IMPLEMENTED",
            "message": "GetMe is not implemented yet",
        },
    })
}


func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusNotImplemented, map[string]any{
        "success": false,
        "error": map[string]any{
            "code": "NOT_IMPLEMENTED",
            "message": "UpdateMe is not implemented yet",
        },
    })
}


func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusNotImplemented, map[string]any{
        "success": false,
        "error": map[string]any{
            "code": "NOT_IMPLEMENTED",
            "message": "GetByID is not implemented yet",
        },
    })
}


func (h *UserHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusNotImplemented, map[string]any{
        "success": false,
        "error": map[string]any{
            "code": "NOT_IMPLEMENTED",
            "message": "FollowUser is not implemented yet",
        },
    })
}


func (h *UserHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusNotImplemented, map[string]any{
        "success": false,
        "error": map[string]any{
            "code": "NOT_IMPLEMENTED",
            "message": "UnfollowUser is not implemented yet",
        },
    })
}
