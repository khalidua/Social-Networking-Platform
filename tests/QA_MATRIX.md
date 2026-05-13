# QA Matrix

| Area | Command / Evidence | Covers |
| --- | --- | --- |
| Unit tests | `go test ./...` in each service | handlers, middleware, services, repositories |
| Core coverage | `powershell -ExecutionPolicy Bypass -File tests\coverage\run-core-coverage.ps1 -SkipReportWrite` | auth, users, posts, feed, notification service logic above 80% |
| Contract tests | `powershell -ExecutionPolicy Bypass -File tests\contract\contract-validation.ps1` | OpenAPI paths/schemas and Kafka event schemas |
| E2E flow | `powershell -ExecutionPolicy Bypass -File tests\integration\e2e\e2e-user-flow.ps1 -StartStack` | auth session validation, profile, follow, post, feed, notifications |
| Load/stress | `powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -K6Runner docker` | gateway reads, post interactions, notification reads |
| Observability demo | `tests\integration\observability-demo-manual-test.md` | request rate, error rate, p95/p99, business metrics, DB metrics |
| Deployment | `docker compose -f deploy\compose\compose.yml config --quiet` and `deploy\scripts\health.ps1` | Compose validity and runtime health |

## Release Gate

All rows must pass before final submission. If live Docker tests are skipped because of local environment limits, record the reason and keep script validation output with the final evidence.
