# SLI Definitions

## Latency (p95)
- **Metric:** `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[1m]))`
- **Target:** < 500ms for 95% of requests (excluding long-tail)
- **Measurement:** HTTP request duration from middleware.

## Error Rate
- **Metric:** `rate(http_requests_total{status_group="5xx"}[1m]) / rate(http_requests_total[1m])`
- **Target:** < 1% of requests resulting in 5xx errors.
- **Measurement:** HTTP status codes grouped as 5xx.

## Throughput
- **Metric:** `rate(http_requests_total[1m])`
- **Target:** Depends on load; used for capacity planning.
- **Measurement:** Total requests per second across all services.