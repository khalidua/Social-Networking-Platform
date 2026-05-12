package config

import (
	"testing"
	"time"
)

func TestLoadReadsResilienceConfig(t *testing.T) {
	t.Setenv("UPSTREAM_RETRY_ATTEMPTS", "4")
	t.Setenv("UPSTREAM_RETRY_BACKOFF", "25ms")
	t.Setenv("CIRCUIT_BREAKER_FAILURES", "2")
	t.Setenv("CIRCUIT_BREAKER_OPEN_FOR", "5s")

	cfg := Load()

	if cfg.UpstreamRetryAttempts != 4 {
		t.Fatalf("UpstreamRetryAttempts = %d, want 4", cfg.UpstreamRetryAttempts)
	}
	if cfg.UpstreamRetryBackoff != 25*time.Millisecond {
		t.Fatalf("UpstreamRetryBackoff = %s, want 25ms", cfg.UpstreamRetryBackoff)
	}
	if cfg.CircuitBreakerFailures != 2 {
		t.Fatalf("CircuitBreakerFailures = %d, want 2", cfg.CircuitBreakerFailures)
	}
	if cfg.CircuitBreakerOpenFor != 5*time.Second {
		t.Fatalf("CircuitBreakerOpenFor = %s, want 5s", cfg.CircuitBreakerOpenFor)
	}
}
