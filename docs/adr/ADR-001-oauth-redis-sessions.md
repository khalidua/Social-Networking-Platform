# ADR-001: Google OAuth2 With Redis-Backed Sessions

## Context

The platform needs external sign-in, JWT issuance, logout/session invalidation, and gateway-side token verification.

## Decision

Use Google OAuth2 for user authentication, issue JWTs from auth-service, and persist active sessions in Redis. The API Gateway verifies both JWT claims and matching Redis session state before forwarding protected requests.

## Consequences

- Logout can invalidate active sessions even when JWTs are still cryptographically valid.
- Integration and load tests can seed deterministic JWT/Redis sessions without depending on Google.
- Redis availability is part of the auth path and must be monitored.
