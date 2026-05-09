package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"social-networking-platform/posts-service/internal/domain"
	"social-networking-platform/posts-service/internal/repository/kafka"
	"social-networking-platform/posts-service/internal/repository/postgres"
)

const maxContentRunes = 10000

var (
	ErrEmptyContent   = errors.New("content is required")
	ErrContentTooLong = errors.New("content too long")
	ErrMissingAuthor  = errors.New("missing author id")
)

type PostService interface {
	CreatePost(ctx context.Context, authorID, content string) (*domain.Post, error)
}

type postService struct {
	repo postgres.PostRepository
	pub  kafka.PostProducer
}

func NewPostService(repo postgres.PostRepository, pub kafka.PostProducer) PostService {
	if pub == nil {
		pub = kafka.NoopPostProducer()
	}
	return &postService{repo: repo, pub: pub}
}

func newPostID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func (s *postService) CreatePost(ctx context.Context, authorID, content string) (*domain.Post, error) {
	authorID = strings.TrimSpace(authorID)
	content = strings.TrimSpace(content)
	if authorID == "" {
		return nil, ErrMissingAuthor
	}
	if content == "" {
		return nil, ErrEmptyContent
	}
	if utf8.RuneCountInString(content) > maxContentRunes {
		return nil, ErrContentTooLong
	}
	id, err := newPostID()
	if err != nil {
		return nil, err
	}
	post := domain.Post{
		ID:        id,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: time.Now().UnixMilli(),
	}
	if err := s.repo.Save(post); err != nil {
		return nil, err
	}
	if err := s.pub.PublishCreated(post); err != nil {
		log.Printf("posts-service: kafka publish post.created: %v", err)
	}
	return &post, nil
}
