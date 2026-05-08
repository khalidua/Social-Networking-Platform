package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"social-networking-platform/users-service/internal/domain"
	"social-networking-platform/users-service/internal/repository/postgres"
)

// ErrCannotFollowSelf is returned when follower and followee IDs are equal.
var ErrCannotFollowSelf = errors.New("cannot follow or unfollow yourself")

// ErrValidation indicates client-side field validation failures.
var ErrValidation = errors.New("validation error")

const (
	maxNameRunes           = 120
	maxBioRunes            = 2000
	maxProfilePictureRunes = 2048
)

type followEventPublisher interface {
	PublishFollowed(ctx context.Context, rel domain.Follow) error
}

type UserService interface {
	GetMe(ctx context.Context, userID string) (*domain.User, error)
	UpdateMe(ctx context.Context, userID string, name *string, bio *string, profilePicture *string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	FollowUser(ctx context.Context, followerID, followeeID string) error
	UnfollowUser(ctx context.Context, followerID, followeeID string) error
}

type userService struct {
	users   postgres.UserRepository
	follows postgres.FollowRepository
	events  followEventPublisher
}

func NewUserService(users postgres.UserRepository, follows postgres.FollowRepository, publisher followEventPublisher) UserService {
	if publisher == nil {
		publisher = noopPublisher{}
	}
	return &userService{
		users:   users,
		follows: follows,
		events:  publisher,
	}
}

type noopPublisher struct{}

func (noopPublisher) PublishFollowed(ctx context.Context, rel domain.Follow) error {
	return nil
}

func validation(msg string) error {
	return fmt.Errorf("%s: %w", msg, ErrValidation)
}

func (s *userService) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("missing user id")
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		u = &domain.User{ID: userID}
		if err := s.users.Save(ctx, *u); err != nil {
			return nil, err
		}
	}
	return u, nil
}

func (s *userService) UpdateMe(ctx context.Context, userID string, name *string, bio *string, profilePicture *string) (*domain.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("missing user id")
	}
	if err := validateProfilePatch(name, bio, profilePicture); err != nil {
		return nil, err
	}
	u, err := s.GetMe(ctx, userID)
	if err != nil {
		return nil, err
	}
	if name != nil {
		u.Name = strings.TrimSpace(*name)
	}
	if bio != nil {
		u.Bio = strings.TrimSpace(*bio)
	}
	if profilePicture != nil {
		u.ProfilePicture = strings.TrimSpace(*profilePicture)
	}
	if err := s.users.Save(ctx, *u); err != nil {
		return nil, err
	}
	return u, nil
}

func validateProfilePatch(name *string, bio *string, profilePicture *string) error {
	if name != nil && utf8.RuneCountInString(*name) > maxNameRunes {
		return validation(fmt.Sprintf("name must be at most %d characters", maxNameRunes))
	}
	if bio != nil && utf8.RuneCountInString(*bio) > maxBioRunes {
		return validation(fmt.Sprintf("bio must be at most %d characters", maxBioRunes))
	}
	if profilePicture != nil && utf8.RuneCountInString(*profilePicture) > maxProfilePictureRunes {
		return validation(fmt.Sprintf("profile_picture must be at most %d characters", maxProfilePictureRunes))
	}
	return nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("missing user id")
	}
	return s.users.GetByID(ctx, id)
}

func (s *userService) ensureUser(ctx context.Context, id string) error {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if u != nil {
		return nil
	}
	return s.users.Save(ctx, domain.User{ID: id, Name: ""})
}

func (s *userService) FollowUser(ctx context.Context, followerID, followeeID string) error {
	if strings.TrimSpace(followerID) == "" || strings.TrimSpace(followeeID) == "" {
		return errors.New("missing user id")
	}
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}
	if err := s.ensureUser(ctx, followerID); err != nil {
		return err
	}
	if err := s.ensureUser(ctx, followeeID); err != nil {
		return err
	}
	rel := domain.Follow{FollowerID: followerID, FolloweeID: followeeID}
	added, err := s.follows.Follow(ctx, rel)
	if err != nil {
		return err
	}
	if added {
		if err := s.events.PublishFollowed(ctx, rel); err != nil {
			log.Printf("users-service: kafka publish user.followed: %v", err)
		}
	}
	return nil
}

func (s *userService) UnfollowUser(ctx context.Context, followerID, followeeID string) error {
	if strings.TrimSpace(followerID) == "" || strings.TrimSpace(followeeID) == "" {
		return errors.New("missing user id")
	}
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}
	return s.follows.Unfollow(ctx, domain.Follow{FollowerID: followerID, FolloweeID: followeeID})
}

type StubUserService struct{}

func NewStubUserService() *StubUserService {
	return &StubUserService{}
}

func (s *StubUserService) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	return nil, nil
}
func (s *StubUserService) UpdateMe(ctx context.Context, userID string, name *string, bio *string, profilePicture *string) (*domain.User, error) {
	return nil, nil
}
func (s *StubUserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, nil
}
func (s *StubUserService) FollowUser(ctx context.Context, followerID, followeeID string) error {
	return nil
}
func (s *StubUserService) UnfollowUser(ctx context.Context, followerID, followeeID string) error {
	return nil
}
