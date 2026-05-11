package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"social-networking-platform/notification-service/internal/domain"
)

type fakeNotificationRepository struct {
	notifications []domain.Notification
	err           error
	userID        string
	saved         []domain.Notification
}

func (r *fakeNotificationRepository) Save(ctx context.Context, notification domain.Notification) error {
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, notification)
	return nil
}

func (r *fakeNotificationRepository) ListByUser(ctx context.Context, userID string) ([]domain.Notification, error) {
	r.userID = userID
	if r.err != nil {
		return nil, r.err
	}
	return r.notifications, nil
}

func TestGetNotificationsReturnsUserNotifications(t *testing.T) {
	createdAt := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	repo := &fakeNotificationRepository{
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
	svc := NewService(repo)

	notifications, err := svc.GetNotifications(context.Background(), " user-1 ")
	if err != nil {
		t.Fatalf("GetNotifications returned error: %v", err)
	}
	if repo.userID != "user-1" {
		t.Fatalf("expected trimmed user id user-1, got %q", repo.userID)
	}
	if len(notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifications))
	}
	if notifications[0].ID != "notification-1" {
		t.Fatalf("expected notification-1, got %q", notifications[0].ID)
	}
}

func TestGetNotificationsNormalizesNilResult(t *testing.T) {
	svc := NewService(&fakeNotificationRepository{})

	notifications, err := svc.GetNotifications(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetNotifications returned error: %v", err)
	}
	if notifications == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(notifications) != 0 {
		t.Fatalf("expected no notifications, got %d", len(notifications))
	}
}

func TestGetNotificationsRejectsEmptyUserID(t *testing.T) {
	svc := NewService(&fakeNotificationRepository{})

	_, err := svc.GetNotifications(context.Background(), " ")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestGetNotificationsPropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	svc := NewService(&fakeNotificationRepository{err: repoErr})

	_, err := svc.GetNotifications(context.Background(), "user-1")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestCreateFollowNotificationPersistsForFollowee(t *testing.T) {
	repo := &fakeNotificationRepository{}
	svc := NewService(repo)
	svc.newID = func() string { return "notification-1" }

	err := svc.CreateFollowNotification(context.Background(), " follower-1 ", " followee-1 ")
	if err != nil {
		t.Fatalf("CreateFollowNotification returned error: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 saved notification, got %d", len(repo.saved))
	}
	got := repo.saved[0]
	if got.ID != "notification-1" || got.UserID != "followee-1" || got.Type != "follow" {
		t.Fatalf("unexpected notification: %+v", got)
	}
}

func TestCreateFollowNotificationIgnoresSelfFollow(t *testing.T) {
	repo := &fakeNotificationRepository{}
	svc := NewService(repo)

	err := svc.CreateFollowNotification(context.Background(), "user-1", "user-1")
	if err != nil {
		t.Fatalf("CreateFollowNotification returned error: %v", err)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("expected no saved notifications, got %+v", repo.saved)
	}
}

func TestCreatePostInteractionNotificationPersistsLikeForAuthor(t *testing.T) {
	repo := &fakeNotificationRepository{}
	svc := NewService(repo)
	svc.newID = func() string { return "notification-1" }

	err := svc.CreatePostInteractionNotification(context.Background(), "post-1", "author-1", "actor-1", "like")
	if err != nil {
		t.Fatalf("CreatePostInteractionNotification returned error: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 saved notification, got %d", len(repo.saved))
	}
	got := repo.saved[0]
	if got.ID != "notification-1" || got.UserID != "author-1" || got.Type != "post_like" {
		t.Fatalf("unexpected notification: %+v", got)
	}
}

func TestCreatePostInteractionNotificationRejectsUnsupportedType(t *testing.T) {
	svc := NewService(&fakeNotificationRepository{})

	err := svc.CreatePostInteractionNotification(context.Background(), "post-1", "author-1", "actor-1", "share")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}
