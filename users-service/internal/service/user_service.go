package service

type UserService interface {
    GetMe() error
    UpdateMe() error
    GetByID() error
    FollowUser() error
    UnfollowUser() error
}

type StubUserService struct{}

func NewStubUserService() *StubUserService {
    return &StubUserService{}
}

func (s *StubUserService) GetMe() error       { return nil }
func (s *StubUserService) UpdateMe() error    { return nil }
func (s *StubUserService) GetByID() error     { return nil }
func (s *StubUserService) FollowUser() error  { return nil }
func (s *StubUserService) UnfollowUser() error { return nil }
