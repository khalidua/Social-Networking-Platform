package kafka

type PostConsumer interface {
	Start() error
}

type StubPostConsumer struct{}

func NewStubPostConsumer() *StubPostConsumer {
	return &StubPostConsumer{}
}

func (c *StubPostConsumer) Start() error { return nil }
