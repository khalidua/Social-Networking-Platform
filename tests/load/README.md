# Load and Stress Tests

## Purpose

Issue #64 requires runnable load/stress scripts and recorded results for concurrency and scalability evaluation.

This folder contains k6 scenarios that exercise the API Gateway and downstream services through the integrated Docker Compose stack.

## Scenarios

| Scenario | Script | Purpose |
| --- | --- | --- |
| Gateway read load | `tests/load/k6/gateway-read-load.js` | Exercises authenticated profile, feed, and notification reads through the gateway. |
| Social write stress | `tests/load/k6/social-write-stress.js` | Exercises post interaction writes and notification reads through the gateway. |

The PowerShell runner seeds valid JWT/Redis sessions for multiple users and rotates read tokens across iterations to avoid tripping the gateway per-user rate limiter during normal load runs.

## Prerequisites

- Docker Desktop is running.
- `k6` is installed and available in `PATH`, or Docker can pull/run `grafana/k6:0.53.0`.
- Required compose ports are available.
- Run commands from the repository root.

## Validate Scripts Without Running Load

```powershell
powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -ValidateOnly
```

## Run Full Load and Stress Suite

```powershell
powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -StartStack
```

The runner uses local `k6` when available. If `k6` is not installed, it automatically runs the same scenarios through Docker:

```powershell
powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -K6Runner docker
```

## Run Against an Already Running Stack

```powershell
powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1
```

## Tune Load

```powershell
powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 `
  -ReadVus 20 `
  -ReadDuration "2m" `
  -WriteVus 10 `
  -WriteDuration "1m"
```

## Outputs

The runner writes:

- k6 JSON summaries under `tests/load/reports/`
- a Markdown run report under `tests/load/reports/`

Use these outputs in the final performance evaluation report.

## Manual Test Plan

## Feature Being Tested

Load and stress behavior for authenticated gateway reads and social write interactions.

## Preconditions

- Docker Desktop is running.
- `k6` is installed.
- The compose stack can build and start.
- No stale `snp-*` containers conflict with compose container names.

## Steps

1. Run `docker compose -f deploy\compose\compose.yml config --quiet`.
2. Run `powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -ValidateOnly`.
3. Run `powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -StartStack`.
4. Review the generated files under `tests\load\reports`.
5. Optional: watch Grafana/Prometheus during the test for latency, error rate, Redis, Kafka, and PostgreSQL signals.

## Expected Results

- Validation command exits with status 0.
- Full load command exits with status 0.
- k6 thresholds pass for both scenarios.
- JSON summaries and Markdown report are created under `tests\load\reports`.

## Edge Cases

- Multiple test users are seeded so normal load is distributed across session identities.
- Gateway read tokens rotate across a larger user pool than the VU count so the scenario measures service throughput instead of the per-user limiter.
- The write stress script reuses a seed post and multiple actor users.
- Async notification behavior is exercised through bounded author notification reads after writes.

## Failure Cases

- Missing Docker or unavailable Docker engine fails stack startup.
- Missing local `k6` falls back to Docker by default. If Docker cannot pull/run `grafana/k6:0.53.0`, the scenario fails before load execution.
- Gateway auth/session contract breakage causes setup or k6 checks to fail.
- Service latency/error regressions cause k6 thresholds to fail.
- If setup waits on `GET /api/v1/users/me`, inspect `docker logs snp-users-service --tail 100` and `docker logs snp-api-gateway --tail 100`.

## Observability Demo Runs

Normal load runs should leave gateway demo simulation disabled. To demonstrate latency or failure panels, enable the gateway flags explicitly and restart only the gateway:

```powershell
$env:DEMO_SIMULATION_ENABLED="true"
$env:DEMO_SIMULATION_PATH="/api/v1/feed"
$env:DEMO_LATENCY="2s"
$env:DEMO_FAILURE_RATE="0.3"
docker compose -f deploy\compose\compose.yml up -d --build api-gateway
```

Disable the flags again before recording normal performance evidence.

## Regression Checks

- Run `powershell -ExecutionPolicy Bypass -File tests\integration\e2e\e2e-user-flow.ps1 -StartStack` before larger load runs.
- Run `powershell -ExecutionPolicy Bypass -File tests\contract\contract-validation.ps1` after API/event contract changes.
- Run `powershell -ExecutionPolicy Bypass -File tests\coverage\run-core-coverage.ps1` after business logic changes.
