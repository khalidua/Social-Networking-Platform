package kafka

type InteractionConsumer interface {
	Start() error
}

type StubInteractionConsumer struct{}

func NewStubInteractionConsumer() *StubInteractionConsumer {
	return &StubInteractionConsumer{}
}

func (c *StubInteractionConsumer) Start() error { return nil }
