# Monitoring And Metrics

## Runtime Endpoints

Every gateway/service process exposes Prometheus metrics at `/metrics` through `promhttp.Handler()`.

## HTTP Metrics

- `http_requests_total{service,method,route,status,status_group}`
- `http_request_duration_seconds{service,method,route,status,status_group}`
- `http_requests_active{service}`
- `service_operations_total{service,method,route,status,status_group}` for backward-compatible route/status operation counting

Route labels are grouped and must remain low-cardinality.

## Business Metrics

Service-layer code emits:

- `business_operation_duration_seconds{service,operation}`
- `business_operation_total{service,operation,status}`

Representative operations include auth callback, profile update, follow/unfollow, post CRUD/interaction, feed reads, and notification creation/listing.

## Dependency Metrics

PostgreSQL repositories emit:

- `db_query_duration_seconds{service,operation}`
- `db_errors_total{service,operation}`

Expected no-row lookups are not counted as dependency errors.

## Dashboard Queries

- Request rate: `sum by (service) (rate(http_requests_total[5m]))`
- 5xx error rate: `sum by (service) (rate(http_requests_total{status=~"5.."}[5m])) / clamp_min(sum by (service) (rate(http_requests_total[5m])), 1)`
- p95/p99 latency: `histogram_quantile(0.95, sum by (service, le) (rate(http_request_duration_seconds_bucket[5m])))`
- Business operation rate: `sum by (service, operation, status) (rate(business_operation_total[1m]))`
- DB query rate: `sum by (service, operation) (rate(db_query_duration_seconds_count[1m]))`
- DB errors: `sum by (service, operation) (rate(db_errors_total[1m]))`

## Demo Simulation

The API Gateway can inject demo-only latency and failures when explicitly enabled:

- `DEMO_SIMULATION_ENABLED=true`
- `DEMO_SIMULATION_PATH=/api/v1/feed`
- `DEMO_LATENCY=2s`
- `DEMO_FAILURE_RATE=0.3`

These flags are disabled by default and should only be used for observability demonstrations.
