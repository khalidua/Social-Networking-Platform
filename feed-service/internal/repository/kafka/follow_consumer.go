package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	kafkago "github.com/segmentio/kafka-go"

	"social-networking-platform/feed-service/internal/config"
)

// UserFollowedV1 is the JSON payload published by users-service on topic user.followed.
type UserFollowedV1 struct {
	FollowerID string `json:"follower_id"`
	FolloweeID string `json:"followee_id"`
}

// FollowConsumer runs until ctx is cancelled.
type FollowConsumer interface {
	Run(ctx context.Context) error
	Close() error
}

type StubFollowConsumer struct{}

func NewStubFollowConsumer() *StubFollowConsumer {
	return &StubFollowConsumer{}
}

func (c *StubFollowConsumer) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (c *StubFollowConsumer) Close() error { return nil }

type kafkaFollowConsumer struct {
	r      *kafkago.Reader
	prefix string
}

func parseBrokers(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, b := range parts {
		b = strings.TrimSpace(b)
		if b != "" {
			out = append(out, b)
		}
	}
	return out
}

// NewFollowConsumer builds a Kafka reader for user.followed, or nil consumer if brokers are unset.
func NewFollowConsumer(cfg config.Config) (FollowConsumer, error) {
	brokers := parseBrokers(cfg.KafkaBrokers)
	topic := strings.TrimSpace(cfg.KafkaTopicFollowed)
	if len(brokers) == 0 || topic == "" {
		return NewStubFollowConsumer(), nil
	}
	groupID := strings.TrimSpace(cfg.ServiceName) + "-user-followed"
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		Topic:       topic,
		MinBytes:    1,
		MaxBytes:    10 << 20,
		StartOffset: kafkago.LastOffset,
	})
	return &kafkaFollowConsumer{r: r, prefix: "feed-service kafka user.followed"}, nil
}

func (c *kafkaFollowConsumer) Run(ctx context.Context) error {
	for {
		m, err := c.r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		var ev UserFollowedV1
		if err := json.Unmarshal(m.Value, &ev); err != nil {
			log.Printf("%s: skip invalid JSON offset=%d err=%v", c.prefix, m.Offset, err)
			if err := c.r.CommitMessages(ctx, m); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}
		if strings.TrimSpace(ev.FollowerID) == "" || strings.TrimSpace(ev.FolloweeID) == "" {
			log.Printf("%s: skip empty ids offset=%d", c.prefix, m.Offset)
			if err := c.r.CommitMessages(ctx, m); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}
		if err := c.handleFollowed(ctx, ev); err != nil {
			log.Printf("%s: handle follow err=%v", c.prefix, err)
		}

		if err := c.r.CommitMessages(ctx, m); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}
}

func (c *kafkaFollowConsumer) handleFollowed(
	ctx context.Context,
	ev UserFollowedV1,
) error {

	log.Printf(
		"%s: reconcile follow follower=%q followee=%q policy=future-posts-only",
		c.prefix,
		ev.FollowerID,
		ev.FolloweeID,
	)

	return nil
}

func (c *kafkaFollowConsumer) Close() error {
	if c == nil || c.r == nil {
		return nil
	}
	return c.r.Close()
}
