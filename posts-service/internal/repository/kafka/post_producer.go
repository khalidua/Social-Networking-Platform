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
	PublishInteracted(ctx context.Context, interaction domain.PostInteraction) error
	Close() error
}

type StubPostProducer struct{}

func NewStubPostProducer() *StubPostProducer {
	return &StubPostProducer{}
}

func (p *StubPostProducer) PublishCreated(ctx context.Context, post domain.Post) error {
	return nil
}

func (p *StubPostProducer) PublishInteracted(ctx context.Context, interaction domain.PostInteraction) error {
	return nil
}

func (p *StubPostProducer) Close() error { return nil }

type KafkaPostProducer struct {
	createdWriter    *kafkago.Writer
	interactedWriter *kafkago.Writer
}

// postCreatedFeedV1 matches feed-service/internal/repository/kafka/post_consumer.go.
type postCreatedFeedV1 struct {
	PostID    string `json:"post_id"`
	AuthorID  string `json:"author_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

type postInteractedV1 struct {
	PostID          string `json:"post_id"`
	PostAuthorID    string `json:"post_author_id"`
	ActorID         string `json:"actor_id"`
	InteractionType string `json:"interaction_type"`
	CreatedAt       int64  `json:"created_at"`
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

func NewKafkaPostProducer(brokersCSV string, postCreatedTopic string, postInteractedTopic string) (*KafkaPostProducer, error) {
	brokers := parseBrokers(brokersCSV)
	if len(brokers) == 0 || strings.TrimSpace(postCreatedTopic) == "" || strings.TrimSpace(postInteractedTopic) == "" {
		return nil, errors.New("kafka brokers and topics are required")
	}

	createdWriter := &kafkago.Writer{
		Addr:                   kafkago.TCP(brokers...),
		Topic:                  postCreatedTopic,
		Balancer:               &kafkago.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	interactedWriter := &kafkago.Writer{
		Addr:                   kafkago.TCP(brokers...),
		Topic:                  postInteractedTopic,
		Balancer:               &kafkago.LeastBytes{},
		AllowAutoTopicCreation: true,
	}

	return &KafkaPostProducer{createdWriter: createdWriter, interactedWriter: interactedWriter}, nil
}

func (p *KafkaPostProducer) PublishCreated(ctx context.Context, post domain.Post) error {
	if p == nil || p.createdWriter == nil {
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

	return p.createdWriter.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(post.AuthorID),
		Value: body,
	})
}

func (p *KafkaPostProducer) PublishInteracted(ctx context.Context, interaction domain.PostInteraction) error {
	if p == nil || p.interactedWriter == nil {
		return errors.New("kafka writer not initialized")
	}

	ts := interaction.CreatedAt
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}

	body, err := json.Marshal(postInteractedV1{
		PostID:          interaction.PostID,
		PostAuthorID:    interaction.PostAuthorID,
		ActorID:         interaction.ActorID,
		InteractionType: interaction.InteractionType,
		CreatedAt:       ts,
	})
	if err != nil {
		return err
	}

	return p.interactedWriter.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(interaction.PostID),
		Value: body,
	})
}

func (p *KafkaPostProducer) Close() error {
	if p == nil {
		return nil
	}
	if p.createdWriter != nil {
		if err := p.createdWriter.Close(); err != nil {
			return err
		}
	}
	if p.interactedWriter != nil {
		return p.interactedWriter.Close()
	}
	return nil
}
