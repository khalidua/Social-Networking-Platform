package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"social-networking-platform/notification-service/internal/config"
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

type notificationProcessor interface {
	CreateFollowNotification(ctx context.Context, followerID string, followeeID string) error
	CreatePostInteractionNotification(ctx context.Context, postID string, postAuthorID string, actorID string, interactionType string) error
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
	r         *kafkago.Reader
	dlq       *kafkago.Writer
	processor notificationProcessor
	prefix    string
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

// NewFollowConsumer builds a Kafka reader for user.followed, or a stub if brokers are unset.
func NewFollowConsumer(cfg config.Config, processor notificationProcessor) (FollowConsumer, error) {
	brokers := parseBrokers(cfg.KafkaBrokers)
	topic := strings.TrimSpace(cfg.KafkaTopicFollowed)
	if len(brokers) == 0 || topic == "" || processor == nil {
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
	dlq := &kafkago.Writer{
		Addr:                   kafkago.TCP(brokers...),
		Topic:                  topic + ".dlq",
		Balancer:               &kafkago.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	return &kafkaFollowConsumer{r: r, dlq: dlq, processor: processor, prefix: "notification-service kafka user.followed"}, nil
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
			if dlqErr := c.publishDLQ(ctx, m, err); dlqErr != nil && !errors.Is(dlqErr, context.Canceled) {
				log.Printf("%s: dlq publish failed offset=%d err=%v", c.prefix, m.Offset, dlqErr)
			}
			if err := c.r.CommitMessages(ctx, m); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}
		if strings.TrimSpace(ev.FollowerID) == "" || strings.TrimSpace(ev.FolloweeID) == "" {
			err := errors.New("missing required follower_id or followee_id")
			log.Printf("%s: skip empty ids offset=%d", c.prefix, m.Offset)
			if dlqErr := c.publishDLQ(ctx, m, err); dlqErr != nil && !errors.Is(dlqErr, context.Canceled) {
				log.Printf("%s: dlq publish failed offset=%d err=%v", c.prefix, m.Offset, dlqErr)
			}
			if err := c.r.CommitMessages(ctx, m); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}
		if err := retry(ctx, 3, 100*time.Millisecond, func() error {
			return c.processor.CreateFollowNotification(ctx, ev.FollowerID, ev.FolloweeID)
		}); err != nil {
			log.Printf("%s: notification processing failed offset=%d err=%v", c.prefix, m.Offset, err)
			if dlqErr := c.publishDLQ(ctx, m, err); dlqErr != nil && !errors.Is(dlqErr, context.Canceled) {
				log.Printf("%s: dlq publish failed offset=%d err=%v", c.prefix, m.Offset, dlqErr)
			}
		} else {
			log.Printf("%s: persisted follower_id=%q followee_id=%q partition=%d offset=%d",
				c.prefix, ev.FollowerID, ev.FolloweeID, m.Partition, m.Offset)
		}
		if err := c.r.CommitMessages(ctx, m); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}
}

func (c *kafkaFollowConsumer) publishDLQ(ctx context.Context, m kafkago.Message, processingErr error) error {
	if c == nil || c.dlq == nil {
		return nil
	}
	payload := deadLetterMessage{
		SourceTopic: m.Topic,
		Partition:   m.Partition,
		Offset:      m.Offset,
		Error:       processingErr.Error(),
		Payload:     string(m.Value),
		FailedAt:    time.Now().UTC().UnixMilli(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.dlq.WriteMessages(ctx, kafkago.Message{Key: m.Key, Value: body})
}

func (c *kafkaFollowConsumer) Close() error {
	if c == nil || c.r == nil {
		return nil
	}
	if err := c.r.Close(); err != nil {
		return err
	}
	if c.dlq != nil {
		return c.dlq.Close()
	}
	return nil
}
