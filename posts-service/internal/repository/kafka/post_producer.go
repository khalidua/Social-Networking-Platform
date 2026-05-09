package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	kafkago "github.com/segmentio/kafka-go"

	"social-networking-platform/posts-service/internal/domain"
)

type PostProducer interface {
	PublishCreated(ctx context.Context, post domain.Post) error
	Close() error
}

type StubPostProducer struct{}

func NewStubPostProducer() *StubPostProducer {
	return &StubPostProducer{}
}

func (p *StubPostProducer) PublishCreated(ctx context.Context, post domain.Post) error {
	return nil
}

func (p *StubPostProducer) Close() error { return nil }

// KafkaPostProducer emits JSON messages to post.created (or configured topic).
type KafkaPostProducer struct {
	w *kafkago.Writer
}

type postCreatedEventPayload struct {
	PostID    string `json:"postId"`
	AuthorID  string `json:"authorId"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

func parseBrokers(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, broker := range parts {
		broker = strings.TrimSpace(broker)
		if broker != "" {
			out = append(out, broker)
		}
	}
	return out
}

func NewKafkaPostProducer(brokersCSV string, topic string) (*KafkaPostProducer, error) {
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

	return &KafkaPostProducer{w: w}, nil
}

func (p *KafkaPostProducer) PublishCreated(ctx context.Context, post domain.Post) error {
	if p == nil || p.w == nil {
		return errors.New("kafka writer not initialized")
	}

	body, err := json.Marshal(postCreatedEventPayload{
		PostID:    post.ID,
		AuthorID:  post.AuthorID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
	})
	if err != nil {
		return err
	}

	return p.w.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(post.AuthorID),
		Value: body,
	})
}

func (p *KafkaPostProducer) Close() error {
	if p == nil || p.w == nil {
		return nil
	}
	return p.w.Close()
}
