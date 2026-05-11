package kafka

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetrySucceedsAfterTransientFailures(t *testing.T) {
	attempts := 0
	err := retry(context.Background(), 3, time.Nanosecond, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary failure")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("retry returned error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestRetryStopsAfterAttemptsExhausted(t *testing.T) {
	attempts := 0
	wantErr := errors.New("persistent failure")
	err := retry(context.Background(), 2, time.Nanosecond, func() error {
		attempts++
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("retry error = %v, want %v", err, wantErr)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestRetryReturnsContextCancellationWhileWaiting(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	err := retry(ctx, 3, time.Hour, func() error {
		attempts++
		cancel()
		return errors.New("temporary failure")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("retry error = %v, want context canceled", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}
