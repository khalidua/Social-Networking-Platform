package kafka

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"social-networking-platform/posts-service/internal/domain"

	kafkago "github.com/segmentio/kafka-go"
)

// PostCreatedV1 must stay aligned with feed-service/internal/repository/kafka/post_consumer.go.
type PostCreatedV1 struct {
	PostID    string `json:"post_id"`
	AuthorID  string `json:"author_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

type PostProducer interface {
	PublishCreated(post domain.Post) error
	Close() error
}

type noopPostProducer struct{}

func (noopPostProducer) PublishCreated(post domain.Post) error { return nil }

func (noopPostProducer) Close() error { return nil }

// NoopPostProducer is used when Kafka is disabled or for tests.
func NoopPostProducer() PostProducer {
	return noopPostProducer{}
}

type kafkaPostProducer struct {
	writer *kafkago.Writer
}

func splitBrokers(csv string) []string {
	csv = strings.TrimSpace(csv)
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// NewPostProducer returns a Kafka writer, or a no-op producer when brokers or topic are unset.
func NewPostProducer(brokersCSV, topic string) PostProducer {
	brokers := splitBrokers(brokersCSV)
	topic = strings.TrimSpace(topic)
	if len(brokers) == 0 || topic == "" {
		return NoopPostProducer()
	}
	return &kafkaPostProducer{
		writer: &kafkago.Writer{
			Addr:  kafkago.TCP(brokers...),
			Topic: topic,
		},
	}
}

func (p *kafkaPostProducer) PublishCreated(post domain.Post) error {
	event := PostCreatedV1{
		PostID:    post.ID,
		AuthorID:  post.AuthorID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
	}
	if event.CreatedAt == 0 {
		event.CreatedAt = time.Now().UnixMilli()
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := kafkago.Message{
		Key:   []byte(post.AuthorID),
		Value: payload,
	}
	return p.writer.WriteMessages(context.Background(), msg)
}

func (p *kafkaPostProducer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
