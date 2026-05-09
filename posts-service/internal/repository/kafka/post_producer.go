package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"social-networking-platform/posts-service/internal/domain"
)

// PostProducer emits post.created for feed-service consumption.
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

type KafkaPostProducer struct {
	w *kafkago.Writer
}

// postCreatedFeedV1 matches feed-service/internal/repository/kafka/post_consumer.go.
type postCreatedFeedV1 struct {
	PostID    string `json:"post_id"`
	AuthorID  string `json:"author_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
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

	ts := post.CreatedAt.UnixMilli()
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}

	body, err := json.Marshal(postCreatedFeedV1{
		PostID:    post.ID,
		AuthorID:  post.AuthorID,
		Content:   post.Content,
		CreatedAt: ts,
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
