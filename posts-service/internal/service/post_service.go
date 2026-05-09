package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"social-networking-platform/posts-service/internal/domain"
	postkafka "social-networking-platform/posts-service/internal/repository/kafka"
	"social-networking-platform/posts-service/internal/repository/postgres"
)

var (
	ErrForbidden    = errors.New("forbidden")
	ErrPostNotFound = errors.New("post not found")
	ErrValidation   = errors.New("validation error")
)

const maxContentRunes = 2000

type PostService interface {
	CreatePost(ctx context.Context, authorID string, content string) (*domain.Post, error)
	GetPost(ctx context.Context, id string) (*domain.Post, error)
	ListPostsByAuthor(ctx context.Context, authorID string) ([]domain.Post, error)
	UpdatePost(ctx context.Context, requesterID string, postID string, content string) (*domain.Post, error)
	DeletePost(ctx context.Context, requesterID string, postID string) error
}

type postService struct {
	repo   postgres.PostRepository
	events postkafka.PostProducer
	newID  func() string
}

func NewPostService(repo postgres.PostRepository, publisher postkafka.PostProducer) PostService {
	if publisher == nil {
		publisher = postkafka.NewStubPostProducer()
	}
	return &postService{
		repo:   repo,
		events: publisher,
		newID:  newPostID,
	}
}

func (s *postService) CreatePost(ctx context.Context, authorID string, content string) (*domain.Post, error) {
	trimmedAuthorID := strings.TrimSpace(authorID)
	trimmedContent := strings.TrimSpace(content)
	if err := validatePostContent(trimmedContent); err != nil {
		return nil, err
	}

	post := &domain.Post{
		ID:       s.newID(),
		AuthorID: trimmedAuthorID,
		Content:  trimmedContent,
	}

	if err := s.repo.CreatePost(ctx, post); err != nil {
		return nil, err
	}
	if err := s.events.PublishCreated(ctx, *post); err != nil {
		log.Printf("posts-service: kafka publish post.created: %v", err)
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
	trimmedRequesterID := strings.TrimSpace(requesterID)
	trimmedContent := strings.TrimSpace(content)
	if err := validatePostContent(trimmedContent); err != nil {
		return nil, err
	}

	post, err := s.repo.GetPostByID(ctx, strings.TrimSpace(postID))
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}
	if post.AuthorID != trimmedRequesterID {
		return nil, ErrForbidden
	}

	post.Content = trimmedContent
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

func validatePostContent(content string) error {
	if content == "" {
		return validation("content is required")
	}
	if utf8.RuneCountInString(content) > maxContentRunes {
		return validation(fmt.Sprintf("content must be at most %d characters", maxContentRunes))
	}
	return nil
}

func validation(msg string) error {
	return fmt.Errorf("%s: %w", msg, ErrValidation)
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
