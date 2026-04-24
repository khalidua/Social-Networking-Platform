package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"social-networking-platform/auth-service/internal/domain"
)

type SessionRepository interface {
	Save(ctx context.Context, session domain.Session) error
	GetByID(ctx context.Context, sessionID string) (*domain.Session, error)
	DeleteByID(ctx context.Context, sessionID string) error
}

type RedisSessionRepository struct {
	client    *Client
	keyPrefix string
}

func NewSessionRepository(client *Client) *RedisSessionRepository {
	return &RedisSessionRepository{
		client:    client,
		keyPrefix: "auth:sessions:",
	}
}

func (r *RedisSessionRepository) Save(ctx context.Context, session domain.Session) error {
	payload, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt.UTC())
	if ttl <= 0 {
		return fmt.Errorf("session TTL must be positive")
	}

	_, err = r.client.Do(
		ctx,
		"SET",
		r.key(session.ID),
		string(payload),
		"EX",
		strconv.FormatInt(int64(ttl.Seconds()), 10),
	)
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

func (r *RedisSessionRepository) GetByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	reply, err := r.client.Do(ctx, "GET", r.key(sessionID))
	if err != nil {
		if errors.Is(err, ErrNil) {
			return nil, nil
		}
		return nil, fmt.Errorf("load session: %w", err)
	}

	raw, ok := reply.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected session payload type %T", reply)
	}

	var session domain.Session
	if err := json.Unmarshal([]byte(raw), &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &session, nil
}

func (r *RedisSessionRepository) DeleteByID(ctx context.Context, sessionID string) error {
	if _, err := r.client.Do(ctx, "DEL", r.key(sessionID)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *RedisSessionRepository) key(sessionID string) string {
	return r.keyPrefix + sessionID
}
