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

// PostInteractedV1 is the JSON payload published by posts-service on topic post.interacted.
type PostInteractedV1 struct {
	PostID          string `json:"post_id"`
	PostAuthorID    string `json:"post_author_id"`
	ActorID         string `json:"actor_id"`
	InteractionType string `json:"interaction_type"`
	CreatedAt       int64  `json:"created_at"`
}

type InteractionConsumer interface {
	Run(ctx context.Context) error
	Close() error
}

type StubInteractionConsumer struct{}

func NewStubInteractionConsumer() *StubInteractionConsumer {
	return &StubInteractionConsumer{}
}

func (c *StubInteractionConsumer) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (c *StubInteractionConsumer) Close() error { return nil }

type kafkaInteractionConsumer struct {
	r         *kafkago.Reader
	dlq       *kafkago.Writer
	processor notificationProcessor
	prefix    string
}

func NewInteractionConsumer(cfg config.Config, processor notificationProcessor) (InteractionConsumer, error) {
	brokers := parseBrokers(cfg.KafkaBrokers)
	topic := strings.TrimSpace(cfg.KafkaTopicPostInteracted)
	if len(brokers) == 0 || topic == "" || processor == nil {
		return NewStubInteractionConsumer(), nil
	}
	groupID := strings.TrimSpace(cfg.ServiceName) + "-post-interacted"
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
	return &kafkaInteractionConsumer{r: r, dlq: dlq, processor: processor, prefix: "notification-service kafka post.interacted"}, nil
}

func (c *kafkaInteractionConsumer) Run(ctx context.Context) error {
	for {
		m, err := c.r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		var ev PostInteractedV1
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
		if strings.TrimSpace(ev.PostID) == "" ||
			strings.TrimSpace(ev.PostAuthorID) == "" ||
			strings.TrimSpace(ev.ActorID) == "" ||
			strings.TrimSpace(ev.InteractionType) == "" {
			err := errors.New("missing required post interaction fields")
			log.Printf("%s: skip empty required fields offset=%d", c.prefix, m.Offset)
			if dlqErr := c.publishDLQ(ctx, m, err); dlqErr != nil && !errors.Is(dlqErr, context.Canceled) {
				log.Printf("%s: dlq publish failed offset=%d err=%v", c.prefix, m.Offset, dlqErr)
			}
			if err := c.r.CommitMessages(ctx, m); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}
		if err := retry(ctx, 3, 100*time.Millisecond, func() error {
			return c.processor.CreatePostInteractionNotification(ctx, ev.PostID, ev.PostAuthorID, ev.ActorID, ev.InteractionType)
		}); err != nil {
			log.Printf("%s: notification processing failed offset=%d err=%v", c.prefix, m.Offset, err)
			if dlqErr := c.publishDLQ(ctx, m, err); dlqErr != nil && !errors.Is(dlqErr, context.Canceled) {
				log.Printf("%s: dlq publish failed offset=%d err=%v", c.prefix, m.Offset, dlqErr)
			}
		} else {
			log.Printf("%s: persisted post_id=%q actor_id=%q interaction_type=%q partition=%d offset=%d",
				c.prefix, ev.PostID, ev.ActorID, ev.InteractionType, m.Partition, m.Offset)
		}
		if err := c.r.CommitMessages(ctx, m); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}
}

func (c *kafkaInteractionConsumer) publishDLQ(ctx context.Context, m kafkago.Message, processingErr error) error {
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

func (c *kafkaInteractionConsumer) Close() error {
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
