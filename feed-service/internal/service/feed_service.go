package service

import (
	"log"
	"sync"
	"time"

	"social-networking-platform/feed-service/internal/domain"
	"social-networking-platform/feed-service/internal/repository/redis"
)

type FeedService interface {
	GetFeed(userID string) (FeedResult, error)
}

type FeedResult struct {
	Items    []domain.FeedItem
	Degraded bool
	Stale    bool
}

type feedService struct {
	repo redis.FeedRepository

	fallbackMu sync.RWMutex
	fallback   map[string][]domain.FeedItem
}

func NewFeedService(
	repo redis.FeedRepository,
) FeedService {
	return &feedService{
		repo:     repo,
		fallback: make(map[string][]domain.FeedItem),
	}
}

func (s *feedService) GetFeed(
	userID string,
) (FeedResult, error) {
	started := time.Now()
	status := businessStatusFailure
	defer func() {
		observeBusinessOperation("get_feed", started, status)
	}()
	feed, err := s.repo.GetFeed(userID)
	// Redis succeeded
	if err == nil {

		// Save latest successful feed in memory
		s.fallbackMu.Lock()
		s.fallback[userID] = feed
		s.fallbackMu.Unlock()

		status = businessStatusSuccess
		return FeedResult{
			Items:    feed,
			Degraded: false,
			Stale:    false,
		}, nil
	}

	// Redis failed: try fallback cache
	s.fallbackMu.RLock()
	cachedFeed, ok := s.fallback[userID]
	s.fallbackMu.RUnlock()

	if ok {

		log.Printf(
			"feed-service: serving stale fallback feed for user=%s due to redis error: %v",
			userID,
			err,
		)

		status = businessStatusSuccess
		return FeedResult{
			Items:    cachedFeed,
			Degraded: true,
			Stale:    true,
		}, nil
	}

	// No fallback available
	return FeedResult{}, err
}
