# Manual Test Plan

## Feature Being Tested

Issue 9.8 distributed tracing:

- Gateway and services accept W3C `Traceparent`.
- Services preserve trace id and create service-local span ids.
- Gateway forwards trace context downstream.
- Logs contain `trace_id` and `span_id` so a single request can be followed across services.

## Preconditions

- Docker Desktop is running.
- Centralized logging stack from issue 9.1 is running.
- Grafana/Loki are available.

## Steps

1. Run automated tests:

   ```powershell
   cd "D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\api-gateway"
   $env:GOCACHE="D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\.gocache"; go test ./...

   cd "D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\auth-service"
   $env:GOCACHE="D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\.gocache"; go test ./...

   cd "D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\users-service"
   $env:GOCACHE="D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\.gocache"; go test ./...

   cd "D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\posts-service"
   $env:GOCACHE="D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\.gocache"; go test ./...

   cd "D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\feed-service"
   $env:GOCACHE="D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\.gocache"; go test ./...

   cd "D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\notification-service"
   $env:GOCACHE="D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\.gocache"; go test ./...
   ```

2. Start the local stack:

   ```powershell
   docker compose -f deploy\compose\compose.yml up --build
   ```

3. Send a request with a known trace id:

   ```powershell
   $trace = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
   Invoke-WebRequest `
     -Uri http://localhost:8080/health `
     -Headers @{ Traceparent = $trace; "X-Request-ID" = "trace-test-1" }
   ```

4. Confirm the response includes a `Traceparent` header.

5. Send a gateway request that proxies downstream, using a valid JWT/session setup if the route is protected.

6. Open Grafana Explore with the Loki datasource.

7. Search by trace id:

   ```logql
   {service=~".+"} |= "4bf92f3577b34da6a3ce929d0e0e4736"
   ```

8. Confirm matching logs from the gateway and downstream service share the same trace id.

## Expected Results

- Every service returns a `Traceparent` response header.
- Incoming trace ids are preserved.
- Each service creates a new span id.
- Gateway forwards `Traceparent` to downstream services.
- Logs contain `trace_id` and `span_id`.
- A single request can be followed across services by searching the trace id in Loki.

## Edge Cases

- Missing `Traceparent` generates a new trace id.
- Invalid `Traceparent` generates a new trace id.
- `trace_id` and `span_id` are not Loki labels.

## Failure Cases

- If downstream proxying fails, gateway error logs still include `trace_id` and `span_id`.
- If circuit breaker opens, resilience logs still include trace fields.

## Regression Checks

- Existing `X-Request-ID` and `X-Correlation-ID` behavior remains unchanged.
- Existing `/metrics` endpoints still work.
- Existing auth, rate limit, retry, and circuit breaker behavior remains unchanged.
