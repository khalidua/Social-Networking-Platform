package kafka

import "social-networking-platform/users-service/internal/domain"

type FollowProducer interface {
    PublishFollowed(rel domain.Follow) error
}

type StubFollowProducer struct{}

func NewStubFollowProducer() *StubFollowProducer {
    return &StubFollowProducer{}
}

func (p *StubFollowProducer) PublishFollowed(rel domain.Follow) error { return nil }
