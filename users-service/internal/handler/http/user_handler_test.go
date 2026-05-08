package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"social-networking-platform/users-service/internal/apiresponse"
	"social-networking-platform/users-service/internal/domain"
	handlers "social-networking-platform/users-service/internal/handler/http"
	"social-networking-platform/users-service/internal/middleware"
	"social-networking-platform/users-service/internal/service"
)

type mockUserService struct {
	getMeResp     *domain.User
	getMeErr      error
	updateMeResp  *domain.User
	updateMeErr   error
	getByIDResp   *domain.User
	getByIDErr    error
	followErr     error
	unfollowErr   error
	lastFollowIDs [2]string
}

func (m *mockUserService) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	if m.getMeErr != nil {
		return nil, m.getMeErr
	}
	return m.getMeResp, nil
}

func (m *mockUserService) UpdateMe(ctx context.Context, userID string, name *string, bio *string, profilePicture *string) (*domain.User, error) {
	if m.updateMeErr != nil {
		return nil, m.updateMeErr
	}
	return m.updateMeResp, nil
}

func (m *mockUserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResp, nil
}

func (m *mockUserService) FollowUser(ctx context.Context, followerID, followeeID string) error {
	m.lastFollowIDs[0], m.lastFollowIDs[1] = followerID, followeeID
	return m.followErr
}

func (m *mockUserService) UnfollowUser(ctx context.Context, followerID, followeeID string) error {
	m.lastFollowIDs[0], m.lastFollowIDs[1] = followerID, followeeID
	return m.unfollowErr
}

func ctxWithRequestID(t *testing.T) context.Context {
	t.Helper()
	return context.WithValue(context.Background(), middleware.RequestIDKey, "test-req")
}

func TestGetMe_MissingHeader(t *testing.T) {
	h := handlers.NewUserHandler(&mockUserService{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	r = r.WithContext(ctxWithRequestID(t))
	h.GetMe(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", w.Code)
	}
}

func TestGetMe_OK(t *testing.T) {
	svc := &mockUserService{getMeResp: &domain.User{ID: "u1", Name: "A"}}
	h := handlers.NewUserHandler(svc)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	r.Header.Set("X-User-ID", "u1")
	r = r.WithContext(ctxWithRequestID(t))
	h.GetMe(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var env apiresponse.SuccessEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.Success {
		t.Fatal("expected success envelope")
	}
}

func TestUpdateMe_ValidationError(t *testing.T) {
	svc := &mockUserService{updateMeErr: service.ErrValidation}
	h := handlers.NewUserHandler(svc)
	w := httptest.NewRecorder()
	body := `{"name":"x"}`
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", strings.NewReader(body))
	r.Header.Set("X-User-ID", "u1")
	r = r.WithContext(ctxWithRequestID(t))
	h.UpdateMe(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d %s", w.Code, w.Body.String())
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := &mockUserService{getByIDResp: nil, getByIDErr: nil}
	h := handlers.NewUserHandler(svc)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users/missing-person", nil)
	r = r.WithContext(ctxWithRequestID(t))
	h.GetByID(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status %d", w.Code)
	}
}

func TestFollowUser_SelfForbidden(t *testing.T) {
	svc := &mockUserService{followErr: service.ErrCannotFollowSelf}
	h := handlers.NewUserHandler(svc)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/users/alice/follow", nil)
	r.Header.Set("X-User-ID", "alice")
	r = r.WithContext(ctxWithRequestID(t))
	h.FollowUser(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status %d", w.Code)
	}
}

func TestFollowUser_InvalidPath(t *testing.T) {
	h := handlers.NewUserHandler(&mockUserService{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/follow", nil)
	r.Header.Set("X-User-ID", "alice")
	r = r.WithContext(ctxWithRequestID(t))
	h.FollowUser(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d", w.Code)
	}
}

func TestFollowUser_NoContent(t *testing.T) {
	svc := &mockUserService{}
	h := handlers.NewUserHandler(svc)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/users/bob/follow", nil)
	r.Header.Set("X-User-ID", "alice")
	r = r.WithContext(ctxWithRequestID(t))
	h.FollowUser(w, r)
	if w.Code != http.StatusNoContent || w.Body.Len() != 0 {
		t.Fatalf("want 204 empty body got %d %q", w.Code, w.Body.String())
	}
	if svc.lastFollowIDs[0] != "alice" || svc.lastFollowIDs[1] != "bob" {
		t.Fatalf("wrong ids %+v", svc.lastFollowIDs)
	}
}

func TestUpdateMe_InvalidJSON(t *testing.T) {
	h := handlers.NewUserHandler(&mockUserService{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewReader([]byte("{")))
	r.Header.Set("X-User-ID", "u1")
	r = r.WithContext(ctxWithRequestID(t))
	h.UpdateMe(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func TestGetMe_ServiceError500(t *testing.T) {
	svc := &mockUserService{getMeErr: errors.New("boom")}
	h := handlers.NewUserHandler(svc)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	r.Header.Set("X-User-ID", "u1")
	r = r.WithContext(ctxWithRequestID(t))
	h.GetMe(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", w.Code)
	}
}
