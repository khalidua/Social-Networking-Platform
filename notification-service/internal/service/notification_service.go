package service

type NotificationService interface {
	GetNotifications() error
}

type StubNotificationService struct{}

func NewStubNotificationService() *StubNotificationService {
	return &StubNotificationService{}
}

func (s *StubNotificationService) GetNotifications() error { return nil }
