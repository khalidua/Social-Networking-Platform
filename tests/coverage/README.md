# Unit Coverage Strategy

## Purpose

Issue #60 and #61 require a per-service unit testing strategy and at least 80% coverage for core business logic modules.

The issue scope names these domain services:

- `auth-service`
- `users-service`
- `posts-service`
- `feed-service`
- `notification-service`

The API gateway remains covered by its existing handler, middleware, proxy, and config tests, but it is outside the service list named in issue #60.

## Core Coverage Targets

| Service | Target package | Core logic covered |
| --- | --- | --- |
| `auth-service` | `./internal/service` | OAuth login start, callback validation, JWT-backed session validation, logout |
| `users-service` | `./internal/service` | profile lifecycle, follow/unfollow rules, follower lookup |
| `posts-service` | `./internal/service` | post creation/update/delete rules, ownership checks, interaction validation |
| `feed-service` | `./internal/service` | feed reads and stale fallback behavior when Redis is unavailable |
| `notification-service` | `./internal/service` | notification reads, follow notification creation, post interaction notification creation |

## How to Run

From the repository root:

```powershell
powershell -ExecutionPolicy Bypass -File tests\coverage\run-core-coverage.ps1
```

To use a stricter threshold:

```powershell
powershell -ExecutionPolicy Bypass -File tests\coverage\run-core-coverage.ps1 -MinimumCoverage 85
```

The script writes `tests\coverage\core-coverage-report.md` and fails if any core target is below the configured threshold.

## Full Regression Command

Run all Go tests from each service module:

```powershell
$root = (Resolve-Path '.').Path
$env:GOCACHE = Join-Path $root '.gocache'
$env:GOTELEMETRY = 'off'
$services = @('api-gateway','auth-service','users-service','posts-service','feed-service','notification-service')
foreach ($s in $services) {
    Push-Location $s
    go test ./...
    Pop-Location
}
```

## Manual Test Plan

## Feature Being Tested

Per-service unit testing strategy and 80% core business logic coverage.

## Preconditions

- Go is installed and available in `PATH`.
- Run from the repository root.
- No Docker services are required for these unit coverage checks.

## Steps

1. Run `powershell -ExecutionPolicy Bypass -File tests\coverage\run-core-coverage.ps1`.
2. Confirm each service row reports `Pass`.
3. Run the full regression command above.
4. Confirm all service modules complete with `ok` or `[no test files]`.

## Expected Results

- Core coverage is at least 80% for every target package.
- `tests\coverage\core-coverage-report.md` is updated with the latest measured coverage.
- Full service test suites pass.

## Edge Cases

- Feed service covers Redis failure fallback when a stale in-memory feed exists.
- Auth service covers invalid, missing, revoked, expired, and mismatched sessions.
- Users service covers missing IDs, self-follow rejection, validation limits, and idempotent follow behavior.
- Posts service covers ownership failures, validation failures, missing posts, repository errors, and publish failures.
- Notification service covers missing fields, self-interaction suppression, unsupported types, and repository errors.

## Failure Cases

- The coverage script exits non-zero when any service target is below the threshold.
- The script exits non-zero when `go test` or `go tool cover` fails for a target.

## Regression Checks

- Existing handler, middleware, repository, integration, and service tests still run through `go test ./...` in each service.
