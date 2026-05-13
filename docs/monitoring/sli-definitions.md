# SLI Definitions

## Latency

- **p95:** `histogram_quantile(0.95, sum by (service, le) (rate(http_request_duration_seconds_bucket[5m])))`
- **p99:** `histogram_quantile(0.99, sum by (service, le) (rate(http_request_duration_seconds_bucket[5m])))`
- **Target:** p95 under 500 ms during normal local load.

## Error Rate

- **Metric:** `sum by (service) (rate(http_requests_total{status_group="5xx"}[5m])) / clamp_min(sum by (service) (rate(http_requests_total[5m])), 1)`
- **Target:** under 1% during normal local load.

## Throughput

- **Metric:** `sum by (service) (rate(http_requests_total[5m]))`
- **Use:** capacity and demo traffic visibility.

## Business Operation Rate

- **Metric:** `sum by (service, operation, status) (rate(business_operation_total[1m]))`
- **Use:** verify real service use cases are executing, not only HTTP middleware.

## Dependency Health

- **DB query rate:** `sum by (service, operation) (rate(db_query_duration_seconds_count[1m]))`
- **DB errors:** `sum by (service, operation) (rate(db_errors_total[1m]))`
- **Use:** detect PostgreSQL dependency latency and failures from application code.
