# ADR-005: Gateway Resilience And Gated Demo Simulation

## Context

The gateway already centralizes auth, session verification, rate limiting, retries, and circuit breaking. The project also needs a controlled observability demo showing latency and error-rate changes.

## Decision

Keep normal gateway behavior unchanged and add an explicit demo simulation middleware disabled by default. Demo simulation can inject latency or failures for a configured route prefix only when `DEMO_SIMULATION_ENABLED=true`.

## Consequences

- Normal local and production-like runs are unaffected.
- Demo behavior is deterministic enough to test with 0% and 100% failure rates.
- Metrics capture simulated latency/failures because the middleware runs inside HTTP metrics collection.
