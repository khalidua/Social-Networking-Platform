package postgres

import "social-networking-platform/users-service/internal/domain"

type FollowRepository interface {
    Follow(rel domain.Follow) error
    Unfollow(rel domain.Follow) error
}

type StubFollowRepository struct{}

func NewStubFollowRepository() *StubFollowRepository {
    return &StubFollowRepository{}
}

func (r *StubFollowRepository) Follow(rel domain.Follow) error   { return nil }
func (r *StubFollowRepository) Unfollow(rel domain.Follow) error { return nil }
