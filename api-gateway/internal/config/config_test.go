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

func TestLoadReadsDemoSimulationConfig(t *testing.T) {
	t.Setenv("DEMO_SIMULATION_ENABLED", "true")
	t.Setenv("DEMO_SIMULATION_PATH", "/api/v1/feed")
	t.Setenv("DEMO_LATENCY", "150ms")
	t.Setenv("DEMO_FAILURE_RATE", "0.25")

	cfg := Load()

	if !cfg.DemoSimulationEnabled {
		t.Fatal("DemoSimulationEnabled = false, want true")
	}
	if cfg.DemoSimulationPath != "/api/v1/feed" {
		t.Fatalf("DemoSimulationPath = %q", cfg.DemoSimulationPath)
	}
	if cfg.DemoLatency != 150*time.Millisecond {
		t.Fatalf("DemoLatency = %s, want 150ms", cfg.DemoLatency)
	}
	if cfg.DemoFailureRate != 0.25 {
		t.Fatalf("DemoFailureRate = %v, want 0.25", cfg.DemoFailureRate)
	}
}
