package postgres

import (
	"context"
	"database/sql"
	"time"

	"social-networking-platform/notification-service/internal/domain"
)

type NotificationRepository interface {
	Save(ctx context.Context, notification domain.Notification) error
	ListByUser(ctx context.Context, userID string) ([]domain.Notification, error)
}

type SQLNotificationRepository struct {
	db *sql.DB
}

func NewSQLNotificationRepository(db *sql.DB) *SQLNotificationRepository {
	return &SQLNotificationRepository{db: db}
}

func (r *SQLNotificationRepository) Save(ctx context.Context, notification domain.Notification) error {
	started := time.Now()
	err := r.db.QueryRowContext(ctx, `
INSERT INTO notifications (id, user_id, type, message, is_read)
VALUES ($1, $2, $3, $4, $5)
RETURNING created_at
`, notification.ID, notification.UserID, notification.Type, notification.Message, notification.Read).Scan(&notification.CreatedAt)
	observeDBOperation("insert_notification", started, err)
	return err
}

func (r *SQLNotificationRepository) ListByUser(ctx context.Context, userID string) ([]domain.Notification, error) {
	started := time.Now()
	var opErr error
	defer func() {
		observeDBOperation("list_notifications", started, opErr)
	}()
	rows, err := r.db.QueryContext(ctx, `
SELECT id, user_id, type, message, is_read, created_at
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
`, userID)
	if err != nil {
		opErr = err
		return nil, err
	}
	defer rows.Close()

	notifications := make([]domain.Notification, 0)
	for rows.Next() {
		var notification domain.Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Message,
			&notification.Read,
			&notification.CreatedAt,
		); err != nil {
			opErr = err
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		opErr = err
		return nil, err
	}

	return notifications, nil
}
