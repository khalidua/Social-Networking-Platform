# API Gateway Resilience

The API gateway applies sync downstream resilience around reverse-proxied service calls.

## Scope

The gateway protects calls to:

- `auth-service`
- `users-service`
- `posts-service`
- `feed-service`
- `notification-service`

This covers the main synchronous inter-service call path without changing downstream service ownership.

## Retry Policy

Config:

- `UPSTREAM_RETRY_ATTEMPTS`
- `UPSTREAM_RETRY_BACKOFF`

Default behavior:

- retry up to 3 total attempts
- exponential backoff starting at 100ms
- retry transient transport errors
- retry `502`, `503`, and `504` responses when the request can be safely replayed

Unsafe request bodies are not retried unless Go can replay the body.

## Circuit Breaker Policy

Config:

- `CIRCUIT_BREAKER_FAILURES`
- `CIRCUIT_BREAKER_OPEN_FOR`

Default behavior:

- open after 5 consecutive upstream failures
- remain open for 30 seconds
- return gateway-standard `UPSTREAM_UNAVAILABLE` with HTTP `503` while open
- set `Retry-After` while open
- log circuit breaker state transitions

## Structured Logs

Retry attempts emit:

```json
{
  "event": "upstream_retry",
  "service": "api-gateway",
  "upstream_service": "users-service",
  "attempt": 1
}
```

Circuit transitions emit:

```json
{
  "event": "circuit_breaker_transition",
  "service": "api-gateway",
  "upstream_service": "users-service",
  "state": "open"
}
```

These logs are shipped through the centralized Loki/Promtail pipeline.

## Manual Verification

See `tests/integration/gateway-resilience-manual-test.md`.
