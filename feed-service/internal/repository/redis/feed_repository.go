package redis

import "social-networking-platform/feed-service/internal/domain"

type FeedRepository interface {
	GetFeed(userID string) ([]domain.FeedItem, error)
	AddFeedItem(userID string, item domain.FeedItem) error
}

type StubFeedRepository struct{}

func NewStubFeedRepository() *StubFeedRepository {
	return &StubFeedRepository{}
}

func (r *StubFeedRepository) GetFeed(userID string) ([]domain.FeedItem, error)      { return nil, nil }
func (r *StubFeedRepository) AddFeedItem(userID string, item domain.FeedItem) error { return nil }
