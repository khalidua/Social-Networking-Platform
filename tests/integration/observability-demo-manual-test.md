# Observability Demo Manual Test

## Feature Being Tested

Demo validation for request rate, p95/p99 latency, error rate, business metrics, and DB dependency metrics.

## Preconditions

- Docker Desktop is running.
- Stack is started with `deploy\scripts\up.ps1 -Build`.
- Grafana is available at `http://localhost:3000`.
- Prometheus is available at `http://localhost:9090`.

## Steps

1. Run the E2E flow to seed and exercise core paths.
2. Run the load suite with Docker k6.
3. Open the Social Networking Platform Grafana dashboard.
4. Confirm request rate, business operation rate, and DB query rate increase.
5. Set:
   ```powershell
   $env:DEMO_SIMULATION_ENABLED="true"
   $env:DEMO_SIMULATION_PATH="/api/v1/feed"
   $env:DEMO_LATENCY="2s"
   $env:DEMO_FAILURE_RATE="0"
   docker compose -f deploy\compose\compose.yml up -d --build api-gateway
   ```
6. Rerun the read load scenario and confirm p95/p99 latency rises.
7. Set:
   ```powershell
   $env:DEMO_FAILURE_RATE="0.3"
   docker compose -f deploy\compose\compose.yml up -d --build api-gateway
   ```
8. Rerun the read load scenario and confirm 5xx error rate rises.
9. Disable demo simulation and restart the gateway.

## Expected Results

- Normal load increases request and business operation rates.
- DB-backed service calls emit `db_query_duration_seconds_count`.
- Latency simulation increases p95/p99 for gateway feed requests.
- Failure simulation increases HTTP 5xx error rate.

## Edge Cases

- Use Docker k6 if local k6 is missing.
- Use seeded JWT/Redis sessions if Google OAuth is not configured.

## Failure Cases

- Missing DB metrics indicate repository instrumentation is not wired.
- Empty Grafana dependency panels indicate PromQL labels do not match emitted metrics.
- Demo simulation active during normal tests indicates env cleanup is incomplete.

## Regression Checks

- With `DEMO_SIMULATION_ENABLED=false`, feed requests should behave normally.
- No metric label should include user ID, email, request ID, session ID, or token.
