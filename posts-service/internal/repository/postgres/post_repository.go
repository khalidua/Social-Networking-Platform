package postgres

import (
	"sync"

	"social-networking-platform/posts-service/internal/domain"
)

type PostRepository interface {
	Save(post domain.Post) error
	GetByID(id string) (*domain.Post, error)
	Update(post domain.Post) error
	Delete(id string) error
}

type StubPostRepository struct{}

func NewStubPostRepository() *StubPostRepository {
	return &StubPostRepository{}
}

func (r *StubPostRepository) Save(post domain.Post) error { return nil }

func (r *StubPostRepository) GetByID(id string) (*domain.Post, error) { return nil, nil }

func (r *StubPostRepository) Update(post domain.Post) error { return nil }

func (r *StubPostRepository) Delete(id string) error { return nil }

// InMemoryPostRepository persists posts in-process (dev / single-node until SQL repo exists).
type InMemoryPostRepository struct {
	mu   sync.RWMutex
	byID map[string]domain.Post
}

func NewInMemoryPostRepository() *InMemoryPostRepository {
	return &InMemoryPostRepository{byID: make(map[string]domain.Post)}
}

func (r *InMemoryPostRepository) Save(post domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[post.ID] = post
	return nil
}

func (r *InMemoryPostRepository) GetByID(id string) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	copy := p
	return &copy, nil
}

func (r *InMemoryPostRepository) Update(post domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[post.ID]; !ok {
		return nil
	}
	r.byID[post.ID] = post
	return nil
}

func (r *InMemoryPostRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byID, id)
	return nil
}
