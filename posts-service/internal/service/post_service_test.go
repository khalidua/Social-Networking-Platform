package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"social-networking-platform/posts-service/internal/domain"
)

type mockPostProducer struct {
	publishCreatedFunc    func(ctx context.Context, post domain.Post) error
	publishInteractedFunc func(ctx context.Context, interaction domain.PostInteraction) error
	publishedPosts        []domain.Post
	publishedInteractions []domain.PostInteraction
	publishCallCount      int
	interactionCallCount  int
}

func (m *mockPostProducer) PublishCreated(ctx context.Context, post domain.Post) error {
	m.publishCallCount++
	m.publishedPosts = append(m.publishedPosts, post)
	if m.publishCreatedFunc != nil {
		return m.publishCreatedFunc(ctx, post)
	}
	return nil
}

func (m *mockPostProducer) PublishInteracted(ctx context.Context, interaction domain.PostInteraction) error {
	m.interactionCallCount++
	m.publishedInteractions = append(m.publishedInteractions, interaction)
	if m.publishInteractedFunc != nil {
		return m.publishInteractedFunc(ctx, interaction)
	}
	return nil
}

func (m *mockPostProducer) Close() error { return nil }

type mockPostRepository struct {
	createPostFunc       func(ctx context.Context, post *domain.Post) error
	getPostByIDFunc      func(ctx context.Context, id string) (*domain.Post, error)
	getPostsByAuthorFunc func(ctx context.Context, authorID string) ([]domain.Post, error)
	updatePostFunc       func(ctx context.Context, post *domain.Post) error
	deletePostFunc       func(ctx context.Context, id string) error

	lastCreated     *domain.Post
	lastFetchedID   string
	lastAuthorID    string
	lastUpdated     *domain.Post
	lastDeletedID   string
	updateCallCount int
	deleteCallCount int
}

func (m *mockPostRepository) CreatePost(ctx context.Context, post *domain.Post) error {
	cp := *post
	m.lastCreated = &cp
	if m.createPostFunc != nil {
		return m.createPostFunc(ctx, post)
	}
	return nil
}

func (m *mockPostRepository) GetPostByID(ctx context.Context, id string) (*domain.Post, error) {
	m.lastFetchedID = id
	if m.getPostByIDFunc != nil {
		return m.getPostByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockPostRepository) GetPostsByAuthor(ctx context.Context, authorID string) ([]domain.Post, error) {
	m.lastAuthorID = authorID
	if m.getPostsByAuthorFunc != nil {
		return m.getPostsByAuthorFunc(ctx, authorID)
	}
	return nil, nil
}

func (m *mockPostRepository) UpdatePost(ctx context.Context, post *domain.Post) error {
	m.updateCallCount++
	cp := *post
	m.lastUpdated = &cp
	if m.updatePostFunc != nil {
		return m.updatePostFunc(ctx, post)
	}
	return nil
}

func (m *mockPostRepository) DeletePost(ctx context.Context, id string) error {
	m.deleteCallCount++
	m.lastDeletedID = id
	if m.deletePostFunc != nil {
		return m.deletePostFunc(ctx, id)
	}
	return nil
}

func TestCreatePost_PersistsGeneratedPost(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		createPostFunc: func(ctx context.Context, post *domain.Post) error {
			post.CreatedAt = time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
			post.UpdatedAt = post.CreatedAt
			return nil
		},
	}
	producer := &mockPostProducer{}
	svc := &postService{
		repo:   repo,
		events: producer,
		newID: func() string {
			return "post-1"
		},
	}

	post, err := svc.CreatePost(ctx, " author-1 ", " hello world ")
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if post == nil {
		t.Fatal("expected post, got nil")
	}
	if post.ID != "post-1" || post.AuthorID != "author-1" || post.Content != "hello world" {
		t.Fatalf("unexpected post: %+v", post)
	}
	if repo.lastCreated == nil {
		t.Fatal("expected repository create call")
	}
	if repo.lastCreated.ID != "post-1" || repo.lastCreated.AuthorID != "author-1" || repo.lastCreated.Content != "hello world" {
		t.Fatalf("unexpected repository payload: %+v", repo.lastCreated)
	}
	if producer.publishCallCount != 1 {
		t.Fatalf("publishCallCount = %d, want 1", producer.publishCallCount)
	}
	if len(producer.publishedPosts) != 1 || producer.publishedPosts[0].ID != "post-1" {
		t.Fatalf("unexpected published posts: %+v", producer.publishedPosts)
	}
}

func TestCreatePost_ValidationErrorWhenContentBlank(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{}
	producer := &mockPostProducer{}
	svc := &postService{
		repo:   repo,
		events: producer,
		newID: func() string {
			return "post-1"
		},
	}

	post, err := svc.CreatePost(ctx, "author-1", "   ")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("CreatePost error = %v, want %v", err, ErrValidation)
	}
	if post != nil {
		t.Fatalf("expected nil post, got %+v", post)
	}
	if repo.lastCreated != nil {
		t.Fatalf("expected repository not to be called, got %+v", repo.lastCreated)
	}
	if producer.publishCallCount != 0 {
		t.Fatalf("publishCallCount = %d, want 0", producer.publishCallCount)
	}
}

func TestCreatePost_ValidationErrorWhenContentTooLong(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{}
	producer := &mockPostProducer{}
	svc := &postService{
		repo:   repo,
		events: producer,
		newID: func() string {
			return "post-1"
		},
	}

	post, err := svc.CreatePost(ctx, "author-1", strings.Repeat("a", maxContentRunes+1))
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("CreatePost error = %v, want %v", err, ErrValidation)
	}
	if post != nil {
		t.Fatalf("expected nil post, got %+v", post)
	}
	if repo.lastCreated != nil {
		t.Fatalf("expected repository not to be called, got %+v", repo.lastCreated)
	}
	if producer.publishCallCount != 0 {
		t.Fatalf("publishCallCount = %d, want 0", producer.publishCallCount)
	}
}

func TestCreatePost_DoesNotPublishWhenRepositoryFails(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		createPostFunc: func(ctx context.Context, post *domain.Post) error {
			return errors.New("insert failed")
		},
	}
	producer := &mockPostProducer{}
	svc := &postService{
		repo:   repo,
		events: producer,
		newID: func() string {
			return "post-1"
		},
	}

	post, err := svc.CreatePost(ctx, "author-1", "hello world")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if post != nil {
		t.Fatalf("expected nil post, got %+v", post)
	}
	if producer.publishCallCount != 0 {
		t.Fatalf("publishCallCount = %d, want 0", producer.publishCallCount)
	}
}

func TestCreatePost_IgnoresPublishFailureAfterInsert(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		createPostFunc: func(ctx context.Context, post *domain.Post) error {
			post.CreatedAt = time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
			post.UpdatedAt = post.CreatedAt
			return nil
		},
	}
	producer := &mockPostProducer{
		publishCreatedFunc: func(ctx context.Context, post domain.Post) error {
			return errors.New("publish failed")
		},
	}
	svc := &postService{
		repo:   repo,
		events: producer,
		newID: func() string {
			return "post-1"
		},
	}

	post, err := svc.CreatePost(ctx, "author-1", "hello world")
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if post == nil {
		t.Fatal("expected post, got nil")
	}
	if producer.publishCallCount != 1 {
		t.Fatalf("publishCallCount = %d, want 1", producer.publishCallCount)
	}
}

func TestGetPost_DelegatesToRepository(t *testing.T) {
	ctx := context.Background()
	want := &domain.Post{ID: "post-1", AuthorID: "author-1", Content: "hello"}
	repo := &mockPostRepository{
		getPostByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return want, nil
		},
	}
	svc := NewPostService(repo, nil)

	post, err := svc.GetPost(ctx, " post-1 ")
	if err != nil {
		t.Fatalf("GetPost: %v", err)
	}
	if post != want {
		t.Fatalf("post = %+v, want %+v", post, want)
	}
	if repo.lastFetchedID != "post-1" {
		t.Fatalf("lastFetchedID = %q, want %q", repo.lastFetchedID, "post-1")
	}
}

func TestListPostsByAuthor_DelegatesToRepository(t *testing.T) {
	ctx := context.Background()
	want := []domain.Post{{ID: "post-1"}, {ID: "post-2"}}
	repo := &mockPostRepository{
		getPostsByAuthorFunc: func(ctx context.Context, authorID string) ([]domain.Post, error) {
			return want, nil
		},
	}
	svc := NewPostService(repo, nil)

	posts, err := svc.ListPostsByAuthor(ctx, " author-1 ")
	if err != nil {
		t.Fatalf("ListPostsByAuthor: %v", err)
	}
	if len(posts) != 2 || posts[0].ID != "post-1" || posts[1].ID != "post-2" {
		t.Fatalf("unexpected posts: %+v", posts)
	}
	if repo.lastAuthorID != "author-1" {
		t.Fatalf("lastAuthorID = %q, want %q", repo.lastAuthorID, "author-1")
	}
}

func TestUpdatePost_UpdatesOwnedPost(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		getPostByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return &domain.Post{
				ID:        "post-1",
				AuthorID:  "author-1",
				Content:   "old",
				CreatedAt: time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
			}, nil
		},
		updatePostFunc: func(ctx context.Context, post *domain.Post) error {
			post.UpdatedAt = time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
			return nil
		},
	}
	svc := NewPostService(repo, nil)

	post, err := svc.UpdatePost(ctx, "author-1", "post-1", " edited ")
	if err != nil {
		t.Fatalf("UpdatePost: %v", err)
	}
	if post.Content != "edited" {
		t.Fatalf("content = %q, want %q", post.Content, "edited")
	}
	if repo.updateCallCount != 1 {
		t.Fatalf("updateCallCount = %d, want 1", repo.updateCallCount)
	}
	if repo.lastUpdated == nil || repo.lastUpdated.Content != "edited" {
		t.Fatalf("unexpected updated payload: %+v", repo.lastUpdated)
	}
}

func TestUpdatePost_RejectsNonOwner(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		getPostByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return &domain.Post{ID: "post-1", AuthorID: "author-2", Content: "old"}, nil
		},
	}
	svc := NewPostService(repo, nil)

	post, err := svc.UpdatePost(ctx, "author-1", "post-1", "edited")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("UpdatePost error = %v, want %v", err, ErrForbidden)
	}
	if post != nil {
		t.Fatalf("expected nil post, got %+v", post)
	}
	if repo.updateCallCount != 0 {
		t.Fatalf("updateCallCount = %d, want 0", repo.updateCallCount)
	}
}

func TestUpdatePost_ValidationErrorWhenContentBlank(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{}
	svc := NewPostService(repo, nil)

	post, err := svc.UpdatePost(ctx, "author-1", "post-1", "   ")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("UpdatePost error = %v, want %v", err, ErrValidation)
	}
	if post != nil {
		t.Fatalf("expected nil post, got %+v", post)
	}
	if repo.lastFetchedID != "" {
		t.Fatalf("expected repository not to be queried, got %q", repo.lastFetchedID)
	}
}

func TestUpdatePost_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		getPostByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return nil, nil
		},
	}
	svc := NewPostService(repo, nil)

	post, err := svc.UpdatePost(ctx, "author-1", "missing", "edited")
	if !errors.Is(err, ErrPostNotFound) {
		t.Fatalf("UpdatePost error = %v, want %v", err, ErrPostNotFound)
	}
	if post != nil {
		t.Fatalf("expected nil post, got %+v", post)
	}
}

func TestUpdatePost_MapsRepositoryNotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		getPostByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return &domain.Post{ID: "post-1", AuthorID: "author-1", Content: "old"}, nil
		},
		updatePostFunc: func(ctx context.Context, post *domain.Post) error {
			return sql.ErrNoRows
		},
	}
	svc := NewPostService(repo, nil)

	post, err := svc.UpdatePost(ctx, "author-1", "post-1", "edited")
	if !errors.Is(err, ErrPostNotFound) {
		t.Fatalf("UpdatePost error = %v, want %v", err, ErrPostNotFound)
	}
	if post != nil {
		t.Fatalf("expected nil post, got %+v", post)
	}
}

func TestDeletePost_DeletesOwnedPost(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		getPostByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return &domain.Post{ID: "post-1", AuthorID: "author-1"}, nil
		},
	}
	svc := NewPostService(repo, nil)

	if err := svc.DeletePost(ctx, "author-1", "post-1"); err != nil {
		t.Fatalf("DeletePost: %v", err)
	}
	if repo.deleteCallCount != 1 {
		t.Fatalf("deleteCallCount = %d, want 1", repo.deleteCallCount)
	}
	if repo.lastDeletedID != "post-1" {
		t.Fatalf("lastDeletedID = %q, want %q", repo.lastDeletedID, "post-1")
	}
}

func TestDeletePost_RejectsNonOwner(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		getPostByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return &domain.Post{ID: "post-1", AuthorID: "author-2"}, nil
		},
	}
	svc := NewPostService(repo, nil)

	err := svc.DeletePost(ctx, "author-1", "post-1")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("DeletePost error = %v, want %v", err, ErrForbidden)
	}
	if repo.deleteCallCount != 0 {
		t.Fatalf("deleteCallCount = %d, want 0", repo.deleteCallCount)
	}
}

func TestDeletePost_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		getPostByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return nil, nil
		},
	}
	svc := NewPostService(repo, nil)

	err := svc.DeletePost(ctx, "author-1", "missing")
	if !errors.Is(err, ErrPostNotFound) {
		t.Fatalf("DeletePost error = %v, want %v", err, ErrPostNotFound)
	}
}

func TestDeletePost_MapsRepositoryNotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		getPostByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return &domain.Post{ID: "post-1", AuthorID: "author-1"}, nil
		},
		deletePostFunc: func(ctx context.Context, id string) error {
			return sql.ErrNoRows
		},
	}
	svc := NewPostService(repo, nil)

	err := svc.DeletePost(ctx, "author-1", "post-1")
	if !errors.Is(err, ErrPostNotFound) {
		t.Fatalf("DeletePost error = %v, want %v", err, ErrPostNotFound)
	}
}

func TestInteractWithPostPublishesLike(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		getPostByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return &domain.Post{ID: "post-1", AuthorID: "author-1"}, nil
		},
	}
	producer := &mockPostProducer{}
	svc := &postService{
		repo:   repo,
		events: producer,
		newID:  func() string { return "post-1" },
	}
	oldNow := timeNowMillis
	timeNowMillis = func() int64 { return 1234 }
	defer func() { timeNowMillis = oldNow }()

	interaction, err := svc.InteractWithPost(ctx, " actor-1 ", " post-1 ", "like")
	if err != nil {
		t.Fatalf("InteractWithPost: %v", err)
	}
	if interaction.PostAuthorID != "author-1" || interaction.ActorID != "actor-1" || interaction.CreatedAt != 1234 {
		t.Fatalf("unexpected interaction: %+v", interaction)
	}
	if producer.interactionCallCount != 1 {
		t.Fatalf("interactionCallCount = %d, want 1", producer.interactionCallCount)
	}
	if len(producer.publishedInteractions) != 1 || producer.publishedInteractions[0].PostID != "post-1" {
		t.Fatalf("unexpected published interactions: %+v", producer.publishedInteractions)
	}
}

func TestInteractWithPostRejectsSelfInteraction(t *testing.T) {
	ctx := context.Background()
	repo := &mockPostRepository{
		getPostByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return &domain.Post{ID: "post-1", AuthorID: "author-1"}, nil
		},
	}
	producer := &mockPostProducer{}
	svc := NewPostService(repo, producer)

	interaction, err := svc.InteractWithPost(ctx, "author-1", "post-1", "like")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("InteractWithPost error = %v, want %v", err, ErrForbidden)
	}
	if interaction != nil {
		t.Fatalf("expected nil interaction, got %+v", interaction)
	}
	if producer.interactionCallCount != 0 {
		t.Fatalf("interactionCallCount = %d, want 0", producer.interactionCallCount)
	}
}

func TestInteractWithPostRejectsUnsupportedType(t *testing.T) {
	ctx := context.Background()
	svc := NewPostService(&mockPostRepository{}, &mockPostProducer{})

	interaction, err := svc.InteractWithPost(ctx, "actor-1", "post-1", "share")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("InteractWithPost error = %v, want %v", err, ErrValidation)
	}
	if interaction != nil {
		t.Fatalf("expected nil interaction, got %+v", interaction)
	}
}
