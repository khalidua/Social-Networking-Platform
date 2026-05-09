package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"social-networking-platform/feed-service/internal/domain"

	goredis "github.com/redis/go-redis/v9"
)

const maxFeedItems = 1000

type FeedRepository interface {
	GetFeed(userID string) ([]domain.FeedItem, error)
	AddFeedItem(userID string, item domain.FeedItem) error
}

type redisFeedRepository struct {
	client *goredis.Client
	ctx    context.Context
}

func NewFeedRepository(client *goredis.Client) FeedRepository {
	return &redisFeedRepository{
		client: client,
		ctx:    context.Background(),
	}
}

func feedKey(userID string) string {
	return fmt.Sprintf("feed:home:%s", userID)
}

func (r *redisFeedRepository) AddFeedItem(
	userID string,
	item domain.FeedItem,
) error {
	key := feedKey(userID)

	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	pipe := r.client.TxPipeline()

	pipe.ZAdd(r.ctx, key, goredis.Z{
		Score:  float64(item.CreatedAt),
		Member: payload,
	})

	pipe.ZRemRangeByRank(r.ctx, key, 0, -maxFeedItems-1)
	_, err = pipe.Exec(r.ctx)
	return err
}

func (r *redisFeedRepository) GetFeed(
	userID string,
) ([]domain.FeedItem, error) {
	key := feedKey(userID)

	results, err := r.client.ZRangeArgs(r.ctx, goredis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  19,
		Rev:   true,
	}).Result()

	if err != nil {
		return nil, err
	}

	feed := make([]domain.FeedItem, 0, len(results))

	for _, raw := range results {
		var item domain.FeedItem
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			return nil, fmt.Errorf("decode feed item: %w", err)
		}
		feed = append(feed, item)
	}
	return feed, nil
}
