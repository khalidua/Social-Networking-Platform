package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"social-networking-platform/notification-service/internal/domain"
	"social-networking-platform/notification-service/internal/repository/postgres"
)

var ErrValidation = errors.New("validation error")

type NotificationService interface {
	GetNotifications(ctx context.Context, userID string) ([]domain.Notification, error)
	CreateFollowNotification(ctx context.Context, followerID string, followeeID string) error
	CreatePostInteractionNotification(ctx context.Context, postID string, postAuthorID string, actorID string, interactionType string) error
}

type Service struct {
	repo  postgres.NotificationRepository
	newID func() string
}

func NewService(repo postgres.NotificationRepository) *Service {
	return &Service{repo: repo, newID: newNotificationID}
}

func (s *Service) GetNotifications(ctx context.Context, userID string) ([]domain.Notification, error) {
	started := time.Now()
	status := businessStatusFailure
	defer func() {
		observeBusinessOperation("get_notifications", started, status)
	}()
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrValidation
	}

	notifications, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if notifications == nil {
		status = businessStatusSuccess
		return []domain.Notification{}, nil
	}
	status = businessStatusSuccess
	return notifications, nil
}

func (s *Service) CreateFollowNotification(ctx context.Context, followerID string, followeeID string) error {
	started := time.Now()
	status := businessStatusFailure
	defer func() {
		observeBusinessOperation("create_follow_notification", started, status)
	}()
	followerID = strings.TrimSpace(followerID)
	followeeID = strings.TrimSpace(followeeID)
	if followerID == "" || followeeID == "" {
		return ErrValidation
	}
	if followerID == followeeID {
		status = businessStatusSuccess
		return nil
	}

	if err := s.repo.Save(ctx, domain.Notification{
		ID:      s.newID(),
		UserID:  followeeID,
		Type:    "follow",
		Message: fmt.Sprintf("%s followed you", followerID),
		Read:    false,
	}); err != nil {
		return err
	}
	status = businessStatusSuccess
	return nil
}

func (s *Service) CreatePostInteractionNotification(ctx context.Context, postID string, postAuthorID string, actorID string, interactionType string) error {
	started := time.Now()
	status := businessStatusFailure
	defer func() {
		observeBusinessOperation("create_post_interaction_notification", started, status)
	}()
	postID = strings.TrimSpace(postID)
	postAuthorID = strings.TrimSpace(postAuthorID)
	actorID = strings.TrimSpace(actorID)
	interactionType = strings.TrimSpace(strings.ToLower(interactionType))
	if postID == "" || postAuthorID == "" || actorID == "" || interactionType == "" {
		return ErrValidation
	}
	if actorID == postAuthorID {
		status = businessStatusSuccess
		return nil
	}
	if interactionType != "like" {
		return fmt.Errorf("unsupported interaction type: %w", ErrValidation)
	}

	if err := s.repo.Save(ctx, domain.Notification{
		ID:      s.newID(),
		UserID:  postAuthorID,
		Type:    "post_like",
		Message: fmt.Sprintf("%s liked your post %s", actorID, postID),
		Read:    false,
	}); err != nil {
		return err
	}
	status = businessStatusSuccess
	return nil
}

func newNotificationID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("notification-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}
