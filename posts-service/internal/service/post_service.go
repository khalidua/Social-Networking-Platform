package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"

	"social-networking-platform/posts-service/internal/domain"
	"social-networking-platform/posts-service/internal/repository/postgres"
)

var (
	ErrForbidden    = errors.New("forbidden")
	ErrPostNotFound = errors.New("post not found")
)

type PostService interface {
	CreatePost(ctx context.Context, authorID string, content string) (*domain.Post, error)
	GetPost(ctx context.Context, id string) (*domain.Post, error)
	ListPostsByAuthor(ctx context.Context, authorID string) ([]domain.Post, error)
	UpdatePost(ctx context.Context, requesterID string, postID string, content string) (*domain.Post, error)
	DeletePost(ctx context.Context, requesterID string, postID string) error
}

type postService struct {
	repo  postgres.PostRepository
	newID func() string
}

func NewPostService(repo postgres.PostRepository) PostService {
	return &postService{
		repo:  repo,
		newID: newPostID,
	}
}

func (s *postService) CreatePost(ctx context.Context, authorID string, content string) (*domain.Post, error) {
	post := &domain.Post{
		ID:       s.newID(),
		AuthorID: strings.TrimSpace(authorID),
		Content:  strings.TrimSpace(content),
	}

	if err := s.repo.CreatePost(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *postService) GetPost(ctx context.Context, id string) (*domain.Post, error) {
	return s.repo.GetPostByID(ctx, strings.TrimSpace(id))
}

func (s *postService) ListPostsByAuthor(ctx context.Context, authorID string) ([]domain.Post, error) {
	return s.repo.GetPostsByAuthor(ctx, strings.TrimSpace(authorID))
}

func (s *postService) UpdatePost(ctx context.Context, requesterID string, postID string, content string) (*domain.Post, error) {
	post, err := s.repo.GetPostByID(ctx, strings.TrimSpace(postID))
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}
	if post.AuthorID != strings.TrimSpace(requesterID) {
		return nil, ErrForbidden
	}

	post.Content = strings.TrimSpace(content)
	if err := s.repo.UpdatePost(ctx, post); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	return post, nil
}

func (s *postService) DeletePost(ctx context.Context, requesterID string, postID string) error {
	post, err := s.repo.GetPostByID(ctx, strings.TrimSpace(postID))
	if err != nil {
		return err
	}
	if post == nil {
		return ErrPostNotFound
	}
	if post.AuthorID != strings.TrimSpace(requesterID) {
		return ErrForbidden
	}

	if err := s.repo.DeletePost(ctx, post.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPostNotFound
		}
		return err
	}

	return nil
}

func newPostID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic("posts-service: unable to generate post id")
	}
	return hex.EncodeToString(buf[:])
}

type StubPostService struct{}

func NewStubPostService() *StubPostService {
	return &StubPostService{}
}

func (s *StubPostService) CreatePost(ctx context.Context, authorID string, content string) (*domain.Post, error) {
	return nil, nil
}

func (s *StubPostService) GetPost(ctx context.Context, id string) (*domain.Post, error) {
	return nil, nil
}

func (s *StubPostService) ListPostsByAuthor(ctx context.Context, authorID string) ([]domain.Post, error) {
	return nil, nil
}

func (s *StubPostService) UpdatePost(ctx context.Context, requesterID string, postID string, content string) (*domain.Post, error) {
	return nil, nil
}

func (s *StubPostService) DeletePost(ctx context.Context, requesterID string, postID string) error {
	return nil
}
