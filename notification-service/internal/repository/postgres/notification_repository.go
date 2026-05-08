package postgres

import "social-networking-platform/notification-service/internal/domain"

type NotificationRepository interface {
	Save(notification domain.Notification) error
	ListByUser(userID string) ([]domain.Notification, error)
}

type StubNotificationRepository struct{}

func NewStubNotificationRepository() *StubNotificationRepository {
	return &StubNotificationRepository{}
}

func (r *StubNotificationRepository) Save(notification domain.Notification) error { return nil }
func (r *StubNotificationRepository) ListByUser(userID string) ([]domain.Notification, error) {
	return nil, nil
}
