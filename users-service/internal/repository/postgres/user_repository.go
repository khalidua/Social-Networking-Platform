package postgres

import "social-networking-platform/users-service/internal/domain"

type UserRepository interface {
    Save(user domain.User) error
    GetByID(id string) (*domain.User, error)
}

type StubUserRepository struct{}

func NewStubUserRepository() *StubUserRepository {
    return &StubUserRepository{}
}

func (r *StubUserRepository) Save(user domain.User) error           { return nil }
func (r *StubUserRepository) GetByID(id string) (*domain.User, error) { return nil, nil }
