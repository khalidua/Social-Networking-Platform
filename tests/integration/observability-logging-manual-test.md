# Manual Test Plan

## Feature Being Tested

Issue 9.1 centralized structured logging pipeline:

- Docker container logs are collected by Promtail.
- Logs are shipped to Loki.
- Grafana can query logs across gateway and services.
- Existing structured JSON log fields remain searchable.

## Preconditions

- Docker Desktop is running.
- Ports `3000`, `3100`, and `9080` are available.
- The local compose stack can build service images.

## Steps

1. Start the stack:

   ```powershell
   docker compose -f deploy\compose\compose.yml up --build
   ```

2. Generate logs:

   ```powershell
   Invoke-RestMethod http://localhost:8080/health
   Invoke-RestMethod http://localhost:8081/health
   Invoke-RestMethod http://localhost:8082/health
   Invoke-RestMethod http://localhost:8083/health
   Invoke-RestMethod http://localhost:8084/health
   Invoke-RestMethod http://localhost:8085/health
   ```

3. Verify Loki readiness:

   ```powershell
   Invoke-RestMethod http://localhost:3100/ready
   ```

4. Verify Promtail readiness:

   ```powershell
   Invoke-RestMethod http://localhost:9080/ready
   ```

5. Open Grafana:

   ```powershell
   Start-Process http://localhost:3000
   ```

6. Sign in with `admin` / `admin`.

7. Open the `Social Networking Platform Logs` dashboard.

8. Query logs in Grafana Explore using Loki:

   ```logql
   {service="api-gateway"}
   {service=~"auth-service|users-service|posts-service|feed-service|notification-service"}
   {route_group="health"}
   ```

9. Search for a request id by log content:

   ```logql
   {service=~".+"} |= "req-"
   ```

## Expected Results

- Loki returns ready.
- Promtail returns ready.
- Grafana has a Loki datasource.
- The logs dashboard shows service logs and log-rate panels.
- Logs can be queried by service, level, event, route group, and status group.
- Request ids remain searchable in log content.

## Edge Cases

- Non-JSON container logs should still be shipped to Loki.
- High-cardinality fields such as request id and raw path should not become Loki labels.
- Gateway and service logs should both be searchable through the same Loki datasource.

## Failure Cases

- Stop a service container and confirm no new logs appear for that service.
- Stop Promtail and confirm Loki no longer receives new Docker logs.
- Stop Loki and confirm Promtail reports push failures in its own logs.

## Regression Checks

- Existing structured JSON access logs are still written to stdout.
- Existing metrics stack and Grafana Prometheus datasource still work.
- Existing service routes and `/health` endpoints still work.
