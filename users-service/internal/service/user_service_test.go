package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"social-networking-platform/users-service/internal/domain"
	"social-networking-platform/users-service/internal/repository/postgres"
	"social-networking-platform/users-service/internal/service"
)

type recordPublisher struct {
	events []domain.Follow
}

func (r *recordPublisher) PublishFollowed(ctx context.Context, rel domain.Follow) error {
	r.events = append(r.events, rel)
	return nil
}

func TestFollowUser_NewFollowPublishesOnce(t *testing.T) {
	ctx := context.Background()
	users := postgres.NewInMemoryUserRepository()
	follows := postgres.NewInMemoryFollowRepository()
	rec := &recordPublisher{}
	svc := service.NewUserService(users, follows, rec)

	if err := svc.FollowUser(ctx, "alice", "bob"); err != nil {
		t.Fatal(err)
	}
	if len(rec.events) != 1 {
		t.Fatalf("want 1 published event, got %d", len(rec.events))
	}
	if err := svc.FollowUser(ctx, "alice", "bob"); err != nil {
		t.Fatal(err)
	}
	if len(rec.events) != 1 {
		t.Fatalf("duplicate follow should not publish again, got %d events", len(rec.events))
	}
}

func TestUpdateMe_Validation(t *testing.T) {
	ctx := context.Background()
	users := postgres.NewInMemoryUserRepository()
	rec := &recordPublisher{}
	svc := service.NewUserService(users, postgres.NewInMemoryFollowRepository(), rec)

	long := strings.Repeat("a", maxNameRunes+1)
	_, err := svc.UpdateMe(ctx, "alice", &long, nil, nil)
	if err == nil || !errors.Is(err, service.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

// maxNameRunes duplicated from service/constants for assertion size only.
const maxNameRunes = 120

func TestFollowUser_SelfForbidden(t *testing.T) {
	ctx := context.Background()
	svc := service.NewUserService(
		postgres.NewInMemoryUserRepository(),
		postgres.NewInMemoryFollowRepository(),
		&recordPublisher{},
	)

	err := svc.FollowUser(ctx, "alice", "alice")
	if err != service.ErrCannotFollowSelf {
		t.Fatalf("expected ErrCannotFollowSelf, got %v", err)
	}
}
