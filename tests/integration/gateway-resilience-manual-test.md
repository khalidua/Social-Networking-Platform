# Manual Test Plan

## Feature Being Tested

Issue 9 Group 3 gateway sync resilience:

- Retry transient downstream failures with exponential backoff.
- Open a circuit breaker after repeated downstream failures.
- Return standardized gateway errors while the circuit is open.
- Log retry attempts and circuit breaker transitions.

## Preconditions

- Docker Desktop is running.
- The gateway can be run with local Compose.
- Monitoring/logging stack from issue 9 Group 2 is available if log verification is required.

## Steps

1. Run the automated gateway test suite:

   ```powershell
   cd "D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\api-gateway"
   $env:GOCACHE="D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\.gocache"; go test ./...
   ```

2. Start the local stack:

   ```powershell
   docker compose -f deploy\compose\compose.yml up --build
   ```

3. Confirm normal gateway health:

   ```powershell
   Invoke-RestMethod http://localhost:8080/health
   ```

4. Stop one downstream service, for example users-service:

   ```powershell
   docker compose -f deploy\compose\compose.yml stop users-service
   ```

5. Send authenticated requests through the gateway to a users route until the configured failure threshold is reached.

6. Confirm the gateway returns a standardized `UPSTREAM_UNAVAILABLE` error with HTTP `503` after the circuit opens.

7. Confirm `Retry-After` is present while the circuit is open.

8. In centralized logs, query retry and circuit events:

   ```logql
   {service="api-gateway", event="upstream_retry"}
   {service="api-gateway", event="circuit_breaker_transition"}
   ```

9. Restart the downstream service:

   ```powershell
   docker compose -f deploy\compose\compose.yml start users-service
   ```

10. Wait for `CIRCUIT_BREAKER_OPEN_FOR`, then retry the request.

## Expected Results

- Transient upstream failures are retried up to the configured attempt count.
- Repeated failures open the circuit breaker.
- Open circuits short-circuit gateway calls without hitting the downstream service.
- Gateway-generated circuit-open errors use the standard response envelope and `UPSTREAM_UNAVAILABLE` code.
- Retry and circuit transition logs are searchable.

## Edge Cases

- Non-replayable unsafe request bodies should not be retried.
- Request and correlation headers must still be forwarded.
- Auth and rate limiting must still run before protected upstream proxying.

## Failure Cases

- Downstream service unavailable: gateway retries and eventually records a failure.
- Downstream remains unavailable: circuit opens and returns `503`.
- Request context cancelled during backoff: retry stops.

## Regression Checks

- Existing gateway auth route proxying still works.
- Existing JWT/session/rate-limit behavior remains unchanged.
- Existing upstream timeout handling still returns `UPSTREAM_UNAVAILABLE`.
