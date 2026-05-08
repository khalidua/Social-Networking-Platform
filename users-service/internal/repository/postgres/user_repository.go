package postgres

import (
	"context"
	"sync"

	"social-networking-platform/users-service/internal/domain"
)

type UserRepository interface {
	Save(ctx context.Context, user domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type StubUserRepository struct{}

func NewStubUserRepository() *StubUserRepository {
	return &StubUserRepository{}
}

func (r *StubUserRepository) Save(ctx context.Context, user domain.User) error {
	return nil
}

func (r *StubUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, nil
}

// InMemoryUserRepository holds users in-process (dev / tests).
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{users: make(map[string]*domain.User)}
}

func (r *InMemoryUserRepository) Save(ctx context.Context, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := user
	r.users[user.ID] = &cp
	return nil
}

func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.users[id]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, nil
}
