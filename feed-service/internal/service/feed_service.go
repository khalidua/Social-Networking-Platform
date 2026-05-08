package service

type FeedService interface {
	GetFeed() error
}

type StubFeedService struct{}

func NewStubFeedService() *StubFeedService {
	return &StubFeedService{}
}

func (s *StubFeedService) GetFeed() error { return nil }
