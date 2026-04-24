package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"social-networking-platform/api-gateway/internal/domain"
)

type SessionRepository struct {
	client    *Client
	keyPrefix string
}

func NewSessionRepository(client *Client) *SessionRepository {
	return &SessionRepository{
		client:    client,
		keyPrefix: "auth:sessions:",
	}
}

func (r *SessionRepository) GetByID(ctx context.Context, sessionID string) (*domain.Session, error) {
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

func (r *SessionRepository) key(sessionID string) string {
	return r.keyPrefix + sessionID
}
