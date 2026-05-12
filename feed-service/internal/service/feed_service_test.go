package service

import (
	"errors"
	"testing"

	"social-networking-platform/feed-service/internal/domain"
)

type fakeFeedRepository struct {
	feed []domain.FeedItem
	err  error
}

func (r *fakeFeedRepository) GetFeed(userID string) ([]domain.FeedItem, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.feed, nil
}

func (r *fakeFeedRepository) AddFeedItem(userID string, item domain.FeedItem) error {
	return nil
}

func TestGetFeedReturnsFreshItemsAndCachesFallback(t *testing.T) {
	repo := &fakeFeedRepository{
		feed: []domain.FeedItem{
			{PostID: "post-1", AuthorID: "author-1", Content: "hello", CreatedAt: 100},
		},
	}
	svc := NewFeedService(repo)

	result, err := svc.GetFeed("user-1")
	if err != nil {
		t.Fatalf("GetFeed returned error: %v", err)
	}
	if result.Degraded || result.Stale {
		t.Fatalf("expected fresh feed, got degraded=%v stale=%v", result.Degraded, result.Stale)
	}
	if len(result.Items) != 1 || result.Items[0].PostID != "post-1" {
		t.Fatalf("unexpected feed items: %+v", result.Items)
	}

	repo.err = errors.New("redis unavailable")
	repo.feed = nil
	result, err = svc.GetFeed("user-1")
	if err != nil {
		t.Fatalf("expected fallback feed, got error: %v", err)
	}
	if !result.Degraded || !result.Stale {
		t.Fatalf("expected stale degraded feed, got degraded=%v stale=%v", result.Degraded, result.Stale)
	}
	if len(result.Items) != 1 || result.Items[0].PostID != "post-1" {
		t.Fatalf("unexpected fallback items: %+v", result.Items)
	}
}

func TestGetFeedReturnsRepositoryErrorWhenNoFallbackExists(t *testing.T) {
	repoErr := errors.New("redis unavailable")
	svc := NewFeedService(&fakeFeedRepository{err: repoErr})

	result, err := svc.GetFeed("user-1")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
	if len(result.Items) != 0 || result.Degraded || result.Stale {
		t.Fatalf("expected empty result on hard failure, got %+v", result)
	}
}
