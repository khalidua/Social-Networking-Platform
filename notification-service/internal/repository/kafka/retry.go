package kafka

import (
	"context"
	"time"
)

type deadLetterMessage struct {
	SourceTopic string `json:"source_topic"`
	Partition   int    `json:"partition"`
	Offset      int64  `json:"offset"`
	Error       string `json:"error"`
	Payload     string `json:"payload"`
	FailedAt    int64  `json:"failed_at"`
}

func retry(ctx context.Context, attempts int, delay time.Duration, fn func() error) error {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt == attempts {
				return lastErr
			}
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			continue
		}
		return nil
	}
	return lastErr
}
