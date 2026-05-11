package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"social-networking-platform/notification-service/internal/domain"
	"social-networking-platform/notification-service/internal/middleware"
)

type fakeNotificationService struct {
	notifications []domain.Notification
	err           error
	userID        string
}

func (s *fakeNotificationService) GetNotifications(ctx context.Context, userID string) ([]domain.Notification, error) {
	s.userID = userID
	if s.err != nil {
		return nil, s.err
	}
	return s.notifications, nil
}

func (s *fakeNotificationService) CreateFollowNotification(ctx context.Context, followerID string, followeeID string) error {
	return nil
}

func (s *fakeNotificationService) CreatePostInteractionNotification(ctx context.Context, postID string, postAuthorID string, actorID string, interactionType string) error {
	return nil
}

func TestGetNotificationsRequiresUserID(t *testing.T) {
	handler := NewNotificationHandler(&fakeNotificationService{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "req-1"))
	rec := httptest.NewRecorder()

	handler.GetNotifications(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["success"] != false {
		t.Fatalf("expected success false, got %v", body["success"])
	}
	if body["request_id"] != "req-1" {
		t.Fatalf("expected request_id req-1, got %v", body["request_id"])
	}
}

func TestGetNotificationsRejectsUnsupportedMethod(t *testing.T) {
	handler := NewNotificationHandler(&fakeNotificationService{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications", nil)
	req.Header.Set("X-User-ID", "user-1")
	req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "req-1"))
	rec := httptest.NewRecorder()

	handler.GetNotifications(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["success"] != false {
		t.Fatalf("expected success false, got %v", body["success"])
	}
}

func TestGetNotificationsReturnsUserNotifications(t *testing.T) {
	createdAt := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	svc := &fakeNotificationService{
		notifications: []domain.Notification{
			{
				ID:        "notification-1",
				UserID:    "user-1",
				Type:      "follow",
				Message:   "user-2 followed you",
				Read:      false,
				CreatedAt: createdAt,
			},
		},
	}
	handler := NewNotificationHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req.Header.Set("X-User-ID", "user-1")
	req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "req-1"))
	rec := httptest.NewRecorder()

	handler.GetNotifications(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if svc.userID != "user-1" {
		t.Fatalf("expected service user id user-1, got %q", svc.userID)
	}
	var body struct {
		Success   bool                  `json:"success"`
		Data      []domain.Notification `json:"data"`
		RequestID string                `json:"request_id"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success {
		t.Fatal("expected success true")
	}
	if body.RequestID != "req-1" {
		t.Fatalf("expected request_id req-1, got %q", body.RequestID)
	}
	if len(body.Data) != 1 || body.Data[0].ID != "notification-1" {
		t.Fatalf("unexpected notifications: %+v", body.Data)
	}
}

func TestGetNotificationsHandlesServiceError(t *testing.T) {
	handler := NewNotificationHandler(&fakeNotificationService{err: errors.New("database unavailable")})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req.Header.Set("X-User-ID", "user-1")
	req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "req-1"))
	rec := httptest.NewRecorder()

	handler.GetNotifications(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["success"] != false {
		t.Fatalf("expected success false, got %v", body["success"])
	}
}
