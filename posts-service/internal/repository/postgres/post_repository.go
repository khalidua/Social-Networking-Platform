package postgres

import (
	"context"
	"sort"
	"sync"

	"social-networking-platform/posts-service/internal/domain"
)

type PostRepository interface {
	CreatePost(ctx context.Context, post *domain.Post) error
	GetPostByID(ctx context.Context, id string) (*domain.Post, error)
	GetPostsByAuthor(ctx context.Context, authorID string) ([]domain.Post, error)
	UpdatePost(ctx context.Context, post *domain.Post) error
	DeletePost(ctx context.Context, id string) error
}

type StubPostRepository struct{}

func NewStubPostRepository() *StubPostRepository {
	return &StubPostRepository{}
}

func (r *StubPostRepository) CreatePost(_ context.Context, post *domain.Post) error {
	return nil
}

func (r *StubPostRepository) GetPostByID(_ context.Context, id string) (*domain.Post, error) {
	return nil, nil
}

func (r *StubPostRepository) GetPostsByAuthor(_ context.Context, authorID string) ([]domain.Post, error) {
	return nil, nil
}

func (r *StubPostRepository) UpdatePost(_ context.Context, post *domain.Post) error {
	return nil
}

func (r *StubPostRepository) DeletePost(_ context.Context, id string) error {
	return nil
}

// InMemoryPostRepository keeps posts in-process for lightweight tests and local wiring.
type InMemoryPostRepository struct {
	mu    sync.RWMutex
	posts map[string]*domain.Post
}

func NewInMemoryPostRepository() *InMemoryPostRepository {
	return &InMemoryPostRepository{
		posts: make(map[string]*domain.Post),
	}
}

func (r *InMemoryPostRepository) CreatePost(_ context.Context, post *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *post
	r.posts[post.ID] = &cp
	return nil
}

func (r *InMemoryPostRepository) GetPostByID(_ context.Context, id string) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, ok := r.posts[id]
	if !ok {
		return nil, nil
	}

	cp := *post
	return &cp, nil
}

func (r *InMemoryPostRepository) GetPostsByAuthor(_ context.Context, authorID string) ([]domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	posts := make([]domain.Post, 0)
	for _, post := range r.posts {
		if post.AuthorID != authorID {
			continue
		}
		posts = append(posts, *post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	return posts, nil
}

func (r *InMemoryPostRepository) UpdatePost(_ context.Context, post *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *post
	r.posts[post.ID] = &cp
	return nil
}

func (r *InMemoryPostRepository) DeletePost(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.posts, id)
	return nil
}
