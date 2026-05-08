package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	kafkago "github.com/segmentio/kafka-go"

	"social-networking-platform/users-service/internal/domain"
)

type FollowProducer interface {
	PublishFollowed(ctx context.Context, rel domain.Follow) error
	Close() error
}

type StubFollowProducer struct{}

func NewStubFollowProducer() *StubFollowProducer {
	return &StubFollowProducer{}
}

func (p *StubFollowProducer) PublishFollowed(ctx context.Context, rel domain.Follow) error {
	return nil
}

func (p *StubFollowProducer) Close() error { return nil }

// KafkaFollowProducer emits JSON messages to user.followed (or configured topic).
type KafkaFollowProducer struct {
	w *kafkago.Writer
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

func NewKafkaFollowProducer(brokersCSV string, topic string) (*KafkaFollowProducer, error) {
	brokers := parseBrokers(brokersCSV)
	if len(brokers) == 0 || strings.TrimSpace(topic) == "" {
		return nil, errors.New("kafka brokers and topic are required")
	}
	w := &kafkago.Writer{
		Addr:                   kafkago.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafkago.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	return &KafkaFollowProducer{w: w}, nil
}

type followEventPayload struct {
	FollowerID string `json:"follower_id"`
	FolloweeID string `json:"followee_id"`
}

func (p *KafkaFollowProducer) PublishFollowed(ctx context.Context, rel domain.Follow) error {
	if p == nil || p.w == nil {
		return errors.New("kafka writer not initialized")
	}
	body, err := json.Marshal(followEventPayload{
		FollowerID: rel.FollowerID,
		FolloweeID: rel.FolloweeID,
	})
	if err != nil {
		return err
	}
	return p.w.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(rel.FollowerID),
		Value: body,
	})
}

func (p *KafkaFollowProducer) Close() error {
	if p == nil || p.w == nil {
		return nil
	}
	return p.w.Close()
}
