package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	kafkago "github.com/segmentio/kafka-go"

	"social-networking-platform/feed-service/internal/config"
	"social-networking-platform/feed-service/internal/domain"
	redisrepo "social-networking-platform/feed-service/internal/repository/redis"
)

// PostCreatedV1 matches posts-service kafka payload (see posts-service/internal/repository/kafka/post_producer.go).
type PostCreatedV1 struct {
	PostID    string `json:"post_id"`
	AuthorID  string `json:"author_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

// FollowerIDsProvider resolves who should receive a post authored by authorID.
type FollowerIDsProvider interface {
	FollowerIDs(ctx context.Context, authorID string) ([]string, error)
}

// NopFollowerIDs returns no followers (no fan-out).
type NopFollowerIDs struct{}

func (NopFollowerIDs) FollowerIDs(ctx context.Context, authorID string) ([]string, error) {
	return nil, nil
}

type PostConsumer interface {
	Run(ctx context.Context) error
	Close() error
}

type stubPostConsumer struct{}

func (stubPostConsumer) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (stubPostConsumer) Close() error { return nil }

type kafkaPostConsumer struct {
	reader    *kafkago.Reader
	feedRepo  redisrepo.FeedRepository
	followers FollowerIDsProvider
	prefix    string
}

func NewPostConsumer(
	cfg config.Config,
	feedRepo redisrepo.FeedRepository,
	followers FollowerIDsProvider,
) (PostConsumer, error) {
	if followers == nil {
		followers = NopFollowerIDs{}
	}
	brokers := parseBrokers(cfg.KafkaBrokers)
	topic := strings.TrimSpace(cfg.KafkaTopicPostCreated)
	if len(brokers) == 0 || topic == "" {
		return stubPostConsumer{}, nil
	}
	groupID := strings.TrimSpace(cfg.ServiceName) + "-post-created"
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		Topic:       topic,
		MinBytes:    1,
		MaxBytes:    10 << 20,
		StartOffset: kafkago.LastOffset,
	})
	return &kafkaPostConsumer{
		reader:    r,
		feedRepo:  feedRepo,
		followers: followers,
		prefix:    "feed-service kafka post.created",
	}, nil
}

func (c *kafkaPostConsumer) Run(ctx context.Context) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if err := c.handleMessage(ctx, m); err != nil {
			log.Printf("%s: handle offset=%d err=%v", c.prefix, m.Offset, err)
		}
		if err := c.reader.CommitMessages(ctx, m); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}
}

func (c *kafkaPostConsumer) handleMessage(ctx context.Context, m kafkago.Message) error {
	var event PostCreatedV1
	if err := json.Unmarshal(m.Value, &event); err != nil {
		return err
	}
	if strings.TrimSpace(event.PostID) == "" || strings.TrimSpace(event.AuthorID) == "" {
		return errors.New("skip event: missing post_id or author_id")
	}
	item := domain.FeedItem{
		PostID:    event.PostID,
		AuthorID:  event.AuthorID,
		Content:   event.Content,
		CreatedAt: event.CreatedAt,
	}
	ids, err := c.followers.FollowerIDs(ctx, event.AuthorID)
	if err != nil {
		return err
	}
	for _, followerID := range ids {
		if strings.TrimSpace(followerID) == "" {
			continue
		}
		if err := c.feedRepo.AddFeedItem(followerID, item); err != nil {
			log.Printf("%s: redis add follower=%q post=%q err=%v", c.prefix, followerID, event.PostID, err)
		}
	}
	return nil
}

func (c *kafkaPostConsumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}
