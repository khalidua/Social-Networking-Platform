package service

import (
	"social-networking-platform/feed-service/internal/domain"
	"social-networking-platform/feed-service/internal/repository/redis"
)

type FeedService interface {
	GetFeed(userID string) ([]domain.FeedItem, error)
}

type feedService struct{
	repo redis.FeedRepository
}

func NewFeedService(
	repo redis.FeedRepository,
) FeedService {
	return &feedService{
		repo: repo,
	}
}

func (s *feedService) GetFeed(
	userID string,
) ([]domain.FeedItem, error) {

	return s.repo.GetFeed(userID)
}