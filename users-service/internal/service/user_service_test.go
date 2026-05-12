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

func TestGetMe_CreatesMissingUser(t *testing.T) {
	ctx := context.Background()
	users := postgres.NewInMemoryUserRepository()
	svc := service.NewUserService(users, postgres.NewInMemoryFollowRepository(), &recordPublisher{})

	user, err := svc.GetMe(ctx, "alice")
	if err != nil {
		t.Fatalf("GetMe returned error: %v", err)
	}
	if user.ID != "alice" {
		t.Fatalf("expected created user id alice, got %q", user.ID)
	}

	stored, err := users.GetByID(ctx, "alice")
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if stored == nil || stored.ID != "alice" {
		t.Fatalf("expected user to be persisted, got %+v", stored)
	}
}

func TestUpdateMe_TrimsProfileFields(t *testing.T) {
	ctx := context.Background()
	users := postgres.NewInMemoryUserRepository()
	svc := service.NewUserService(users, postgres.NewInMemoryFollowRepository(), &recordPublisher{})
	name := " Alice "
	bio := " hello "
	picture := " https://example.com/a.png "

	user, err := svc.UpdateMe(ctx, "alice", &name, &bio, &picture)
	if err != nil {
		t.Fatalf("UpdateMe returned error: %v", err)
	}
	if user.Name != "Alice" || user.Bio != "hello" || user.ProfilePicture != "https://example.com/a.png" {
		t.Fatalf("expected trimmed profile fields, got %+v", user)
	}
}

func TestUpdateMe_RejectsLongBioAndProfilePicture(t *testing.T) {
	ctx := context.Background()
	svc := service.NewUserService(
		postgres.NewInMemoryUserRepository(),
		postgres.NewInMemoryFollowRepository(),
		&recordPublisher{},
	)

	longBio := strings.Repeat("b", 2001)
	if _, err := svc.UpdateMe(ctx, "alice", nil, &longBio, nil); !errors.Is(err, service.ErrValidation) {
		t.Fatalf("expected long bio validation error, got %v", err)
	}

	longPicture := strings.Repeat("p", 2049)
	if _, err := svc.UpdateMe(ctx, "alice", nil, nil, &longPicture); !errors.Is(err, service.ErrValidation) {
		t.Fatalf("expected long profile picture validation error, got %v", err)
	}
}

func TestListFollowerIDs_InMemory(t *testing.T) {
	ctx := context.Background()
	users := postgres.NewInMemoryUserRepository()
	follows := postgres.NewInMemoryFollowRepository()
	svc := service.NewUserService(users, follows, &recordPublisher{})

	_ = svc.FollowUser(ctx, "alice", "bob")
	_ = svc.FollowUser(ctx, "carol", "bob")

	ids, err := svc.ListFollowerIDs(ctx, "bob")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("want 2 followers, got %v", ids)
	}
}

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

func TestServiceValidationRejectsMissingUserIDs(t *testing.T) {
	ctx := context.Background()
	svc := service.NewUserService(
		postgres.NewInMemoryUserRepository(),
		postgres.NewInMemoryFollowRepository(),
		&recordPublisher{},
	)

	if _, err := svc.GetMe(ctx, " "); err == nil {
		t.Fatal("expected GetMe to reject missing user id")
	}
	if _, err := svc.GetByID(ctx, " "); err == nil {
		t.Fatal("expected GetByID to reject missing user id")
	}
	if err := svc.FollowUser(ctx, "alice", " "); err == nil {
		t.Fatal("expected FollowUser to reject missing followee id")
	}
	if err := svc.UnfollowUser(ctx, " ", "bob"); err == nil {
		t.Fatal("expected UnfollowUser to reject missing follower id")
	}
	if _, err := svc.ListFollowerIDs(ctx, " "); err == nil {
		t.Fatal("expected ListFollowerIDs to reject missing followee id")
	}
}

func TestUnfollowUser_RemovesRelationship(t *testing.T) {
	ctx := context.Background()
	follows := postgres.NewInMemoryFollowRepository()
	svc := service.NewUserService(
		postgres.NewInMemoryUserRepository(),
		follows,
		&recordPublisher{},
	)

	if err := svc.FollowUser(ctx, "alice", "bob"); err != nil {
		t.Fatalf("FollowUser returned error: %v", err)
	}
	if err := svc.UnfollowUser(ctx, "alice", "bob"); err != nil {
		t.Fatalf("UnfollowUser returned error: %v", err)
	}
	ids, err := svc.ListFollowerIDs(ctx, "bob")
	if err != nil {
		t.Fatalf("ListFollowerIDs returned error: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected no followers after unfollow, got %v", ids)
	}
}

func TestStubUserServiceMethods(t *testing.T) {
	ctx := context.Background()
	stub := service.NewStubUserService()

	if user, err := stub.GetMe(ctx, "user-1"); err != nil || user != nil {
		t.Fatalf("GetMe = %+v, %v; want nil, nil", user, err)
	}
	if user, err := stub.UpdateMe(ctx, "user-1", nil, nil, nil); err != nil || user != nil {
		t.Fatalf("UpdateMe = %+v, %v; want nil, nil", user, err)
	}
	if user, err := stub.GetByID(ctx, "user-1"); err != nil || user != nil {
		t.Fatalf("GetByID = %+v, %v; want nil, nil", user, err)
	}
	if err := stub.FollowUser(ctx, "alice", "bob"); err != nil {
		t.Fatalf("FollowUser returned error: %v", err)
	}
	if err := stub.UnfollowUser(ctx, "alice", "bob"); err != nil {
		t.Fatalf("UnfollowUser returned error: %v", err)
	}
	if ids, err := stub.ListFollowerIDs(ctx, "bob"); err != nil || ids != nil {
		t.Fatalf("ListFollowerIDs = %v, %v; want nil, nil", ids, err)
	}
}
