# Manual Test Plan

## Feature Being Tested

Issue 9 Group 1 observability:

- Application-level metrics collection.
- Host and container metrics collection.
- Grafana dashboard provisioning.
- Prometheus alert rule loading for critical failures.

## Preconditions

- Docker Desktop is running.
- Ports `3000`, `8086`, `9090`, `9100`, `9121`, `9187`, `9188`, `9189`, and `9308` are available.
- The local compose stack can build service images.

## Steps

1. Start the local stack:

   ```powershell
   docker compose -f deploy\compose\compose.yml up --build
   ```

2. Generate a few requests:

   ```powershell
   Invoke-RestMethod http://localhost:8080/health
   Invoke-RestMethod http://localhost:8081/health
   Invoke-RestMethod http://localhost:8082/health
   Invoke-RestMethod http://localhost:8083/health
   Invoke-RestMethod http://localhost:8084/health
   Invoke-RestMethod http://localhost:8085/health
   ```

3. Verify application metrics directly:

   ```powershell
   Invoke-RestMethod http://localhost:8080/metrics
   Invoke-RestMethod http://localhost:8081/metrics
   Invoke-RestMethod http://localhost:8082/metrics
   Invoke-RestMethod http://localhost:8083/metrics
   Invoke-RestMethod http://localhost:8084/metrics
   Invoke-RestMethod http://localhost:8085/metrics
   ```

4. Open Prometheus targets:

   ```powershell
   Start-Process http://localhost:9090/targets
   ```

5. Confirm these targets are `UP`:

   - `api-gateway`
   - `auth-service`
   - `users-service`
   - `posts-service`
   - `feed-service`
   - `notification-service`
   - `node-exporter`
   - `cadvisor`
   - `redis`
   - `kafka`
   - `postgres`

6. Open Grafana:

   ```powershell
   Start-Process http://localhost:3000
   ```

7. Sign in with `admin` / `admin`.

8. Open the `Social Networking Platform Overview` dashboard.

9. Open Prometheus rules:

   ```powershell
   Start-Process http://localhost:9090/rules
   ```

10. Confirm the `social-networking-platform-critical` alert group is loaded.

## Expected Results

- Every service exposes Prometheus-compatible metrics.
- Prometheus shows application, host, container, Redis, Kafka, and PostgreSQL targets.
- Grafana has the Prometheus datasource provisioned.
- The dashboard shows request rate, error rate, latency, active requests, container CPU/memory, Kafka lag, and target health.
- Prometheus alert rules are loaded without syntax errors.

## Edge Cases

- Metrics route labels must use route groups such as `posts`, not raw paths such as `/api/v1/posts/123`.
- `/metrics` must not require user authentication headers.
- Services with no traffic should still expose metric metadata and active request gauge.

## Failure Cases

- Stop a service container and confirm `ServiceDown` becomes pending/firing after its configured duration.
- Stop Redis and confirm the Redis target drops.
- Stop one PostgreSQL exporter and confirm `PostgresExporterDown` becomes pending/firing.

## Regression Checks

- Existing `/health` endpoints still return success.
- Existing API routes still use request ID, logging, recovery, and authentication middleware.
- Gateway proxy routing still works for auth, users, posts, feed, and notifications.
