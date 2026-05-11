# Feed Load Test Guide

This document explains how to run the feed load test with the automated script.

## Test Files

- Load test script: `feed-service/tests/load/feed-load-test.js`
- Runner script: `scripts/dev/token/run-feed-load-test.ps1`
- Session seeding helper: `scripts/dev/token/seed-session.ps1`

## Prerequisites

- Docker stack is running (at least `api-gateway`, `feed-service`, `redis`).
- `k6` is installed and available in PATH.
- `go` is installed and available in PATH.

Start services if needed:

```powershell
docker compose -f deploy/compose/compose.yml up -d
```

## Recommended One-Command Run

Use this command for normal feed load testing:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/dev/token/run-feed-load-test.ps1
```

What this command does:

1. Generates a JWT token (`generate-token.go`).
2. Seeds a matching auth session in Redis.
3. Runs k6 against `http://localhost:8080` with default load (`50 VUs`, `30s`).

## Custom Run Options

Change target URL, VUs, and duration:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/dev/token/run-feed-load-test.ps1 `
  -BaseUrl "http://localhost:8080" `
  -Vus 50 `
  -Duration "30s"
```

Skip session seeding when a valid session already exists:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/dev/token/run-feed-load-test.ps1 -SkipSessionSeed
```

## Interpreting Results

- `status is 200` check validates successful feed responses.
- `latency under 500ms` check validates response-time target.
- If `status is 200` fails heavily after some initial successes, check for API gateway rate limits.

Current gateway config includes a per-user rate limit (`RATE_LIMIT_PER_MINUTE`), so high load with one token can produce `429` responses.

## Troubleshooting

- **`token format is invalid`**: token env was empty/malformed in manual runs. Use the runner script.
- **`token signature is invalid`**: token secret mismatch. Ensure services and generator use same `JWT_SECRET`.
- **`session is invalid or revoked`**: missing Redis session. Use runner script or `seed-session.ps1`.
- **Redis connection refused from feed-service**: ensure feed-service uses `REDIS_HOST`/`REDIS_PORT` and container was rebuilt/restarted.
