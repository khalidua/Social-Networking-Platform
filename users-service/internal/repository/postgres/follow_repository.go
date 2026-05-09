package postgres

import (
	"context"
	"sort"
	"strings"
	"sync"

	"social-networking-platform/users-service/internal/domain"
)

type FollowRepository interface {
	Follow(ctx context.Context, rel domain.Follow) (added bool, err error)
	Unfollow(ctx context.Context, rel domain.Follow) error
	// ListFollowerIDs returns users who follow followeeID (following_id in DB).
	ListFollowerIDs(ctx context.Context, followeeID string) ([]string, error)
}

type StubFollowRepository struct{}

func NewStubFollowRepository() *StubFollowRepository {
	return &StubFollowRepository{}
}

func (r *StubFollowRepository) Follow(ctx context.Context, rel domain.Follow) (bool, error) {
	return true, nil
}

func (r *StubFollowRepository) Unfollow(ctx context.Context, rel domain.Follow) error {
	return nil
}

func (r *StubFollowRepository) ListFollowerIDs(ctx context.Context, followeeID string) ([]string, error) {
	return nil, nil
}

// InMemoryFollowRepository stores follower→followee edges in-process (dev / tests).
type InMemoryFollowRepository struct {
	mu      sync.RWMutex
	follows map[string]struct{}
}

func NewInMemoryFollowRepository() *InMemoryFollowRepository {
	return &InMemoryFollowRepository{
		follows: make(map[string]struct{}),
	}
}

func followKey(rel domain.Follow) string {
	return rel.FollowerID + "\x00" + rel.FolloweeID
}

func (r *InMemoryFollowRepository) Follow(ctx context.Context, rel domain.Follow) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := followKey(rel)
	if _, ok := r.follows[k]; ok {
		return false, nil
	}
	r.follows[k] = struct{}{}
	return true, nil
}

func (r *InMemoryFollowRepository) Unfollow(ctx context.Context, rel domain.Follow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.follows, followKey(rel))
	return nil
}

func (r *InMemoryFollowRepository) ListFollowerIDs(ctx context.Context, followeeID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	suffix := "\x00" + followeeID
	var out []string
	for k := range r.follows {
		if strings.HasSuffix(k, suffix) {
			followerID := strings.TrimSuffix(k, suffix)
			out = append(out, followerID)
		}
	}
	sort.Strings(out)
	return out, nil
}
