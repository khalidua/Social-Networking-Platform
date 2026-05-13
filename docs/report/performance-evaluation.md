# Performance Evaluation

## Test Environment

Run from the repository root on a Docker Desktop environment with the local Compose stack.

## Commands

```powershell
powershell -ExecutionPolicy Bypass -File tests\integration\e2e\e2e-user-flow.ps1 -StartStack
powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -K6Runner docker -ReadVus 20 -ReadDuration "2m" -WriteVus 10 -WriteDuration "1m"
```

## Evidence To Capture

- k6 JSON summaries in `tests/load/reports/`
- generated Markdown report from `tests/load/run-load-tests.ps1`
- Grafana panels during baseline, load, latency simulation, and failure simulation
- Prometheus query results for:
  - `sum by (service) (rate(http_requests_total[5m]))`
  - `histogram_quantile(0.95, sum by (service, le) (rate(http_request_duration_seconds_bucket[5m])))`
  - `histogram_quantile(0.99, sum by (service, le) (rate(http_request_duration_seconds_bucket[5m])))`
  - `sum by (service, operation, status) (rate(business_operation_total[1m]))`
  - `sum by (service, operation) (rate(db_query_duration_seconds_count[1m]))`
  - `sum by (service, operation) (rate(db_errors_total[1m]))`

## Latest Local Results

- Gateway read load on 2026-05-13: 20 VUs, 2m duration, 7,968 HTTP requests, 100% checks, 0% HTTP failures, p95 HTTP request duration 13.7ms.
- Social write stress on 2026-05-13: 10 VUs, 1m duration, 766 HTTP requests, 100% checks, 0% HTTP failures, p95 HTTP request duration 19.02ms.
- Latest successful write report: `tests/load/reports/issue-64-load-report-20260513-232755.md`.
- Earlier full read/write run produced a passing read scenario and exposed Kafka producer batch latency in the write scenario; the producer batch timeout was reduced and the write scenario was rerun successfully.

## Interpretation

- Read load should increase gateway, users, feed, and notification request rate.
- Write stress should exercise posts and notification paths.
- Latency simulation should raise p95 and p99 for the configured path.
- Failure simulation should raise 5xx error rate and keep request envelopes bounded.
