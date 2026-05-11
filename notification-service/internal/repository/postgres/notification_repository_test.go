package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"social-networking-platform/notification-service/internal/domain"
)

func TestSQLNotificationRepository_Save(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLNotificationRepository(db)
	createdAt := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	notification := domain.Notification{
		ID:      "notification-1",
		UserID:  "user-1",
		Type:    "follow",
		Message: "user-2 followed you",
		Read:    false,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
INSERT INTO notifications (id, user_id, type, message, is_read)
VALUES ($1, $2, $3, $4, $5)
RETURNING created_at
`)).
		WithArgs(notification.ID, notification.UserID, notification.Type, notification.Message, notification.Read).
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(createdAt))

	if err := repo.Save(context.Background(), notification); err != nil {
		t.Fatalf("Save: %v", err)
	}
	assertExpectations(t, mock)
}

func TestSQLNotificationRepository_ListByUser(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLNotificationRepository(db)
	firstCreatedAt := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	secondCreatedAt := firstCreatedAt.Add(-1 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id, user_id, type, message, is_read, created_at
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
`)).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "type", "message", "is_read", "created_at"}).
			AddRow("notification-2", "user-1", "follow", "newest", false, firstCreatedAt).
			AddRow("notification-1", "user-1", "follow", "older", true, secondCreatedAt))

	notifications, err := repo.ListByUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(notifications) != 2 {
		t.Fatalf("len(notifications) = %d, want 2", len(notifications))
	}
	if notifications[0].ID != "notification-2" || notifications[1].ID != "notification-1" {
		t.Fatalf("unexpected order: %+v", notifications)
	}
	if !notifications[1].Read {
		t.Fatalf("expected second notification to be read: %+v", notifications[1])
	}
	assertExpectations(t, mock)
}

func assertExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet: %v", err)
	}
}
