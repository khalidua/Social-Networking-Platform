# End-to-End Integration Tests

## Purpose

Issue #62 requires integration coverage for key user flows in a containerized environment.

The runnable harness is:

```powershell
tests\integration\e2e\e2e-user-flow.ps1
```

It drives the API Gateway and downstream services through Docker Compose and verifies:

- gateway health
- auth session validation with a gateway-valid JWT and Redis session
- protected-route authentication failure
- profile update
- follow relationship creation
- follow notification generation
- post creation
- post retrieval
- author post listing
- feed retrieval after Kafka fan-out
- post-like notification generation

## Why Session Seeding Is Used

The production login path depends on external Google OAuth. The integration harness avoids external network dependence by creating a valid HS256 JWT and matching Redis session using the same contracts that auth-service and api-gateway use.

This keeps the test deterministic while still exercising gateway JWT validation, Redis session verification, identity header forwarding, service persistence, Kafka event processing, Redis feed state, and notification reads.

## How to Run

From the repository root, start or rebuild the compose stack and run the E2E flow:

```powershell
powershell -ExecutionPolicy Bypass -File tests\integration\e2e\e2e-user-flow.ps1 -StartStack
```

If the stack is already running:

```powershell
powershell -ExecutionPolicy Bypass -File tests\integration\e2e\e2e-user-flow.ps1
```

Validate script wiring without Docker:

```powershell
powershell -ExecutionPolicy Bypass -File tests\integration\e2e\e2e-user-flow.ps1 -ValidateOnly
```

## Expected Result

The script exits with status 0 and prints:

```text
e2e user flow passed
```

## Manual Test Plan

## Feature Being Tested

End-to-end user flow integration across gateway, auth, users, posts, feed, notifications, Redis, Kafka, and PostgreSQL.

## Preconditions

- Docker Desktop is running.
- Ports from `deploy\compose\compose.yml` are available.
- Run from the repository root.
- The default local JWT secret is `change-me`, matching compose defaults.

## Steps

1. Run `docker compose -f deploy\compose\compose.yml config --quiet`.
2. Run `powershell -ExecutionPolicy Bypass -File tests\integration\e2e\e2e-user-flow.ps1 -StartStack`.
3. Wait for the script to complete.
4. Inspect Docker logs only if the script fails.

## Expected Results

- Health checks pass for gateway and all five downstream services.
- Unauthenticated protected gateway request returns 401.
- Auth session endpoint validates the seeded JWT/session.
- Profile update returns the expected user.
- Follow operation returns 204.
- Bob receives a follow notification.
- Bob creates a post and Alice can retrieve/list it.
- Alice's feed receives Bob's post through Kafka and Redis.
- Bob receives a post-like notification after Alice likes the post.

## Edge Cases

- The script uses unique user and session IDs per run to avoid collisions.
- Async Kafka-driven feed and notification checks poll until timeout.
- The script validates both direct service health and gateway-routed behavior.

## Failure Cases

- Missing Docker/Compose stack fails before flow assertions.
- Bad JWT/session contract fails at auth or gateway session checks.
- Broken Kafka consumers fail the feed or notification polling assertions.
- Broken persistence fails profile, post, or notification assertions.

## Regression Checks

- Run service unit tests with `go test ./...` in each service.
- Run contract tests with `powershell -ExecutionPolicy Bypass -File tests\contract\contract-validation.ps1`.
- Run core unit coverage with `powershell -ExecutionPolicy Bypass -File tests\coverage\run-core-coverage.ps1`.
