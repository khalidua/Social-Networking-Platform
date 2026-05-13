# Final Architecture And Evaluation Report

## Architecture Summary

The Social Networking Platform is implemented as Go microservices behind an API Gateway. The gateway is the public entry point and handles request IDs, tracing headers, JWT/session validation, rate limiting, upstream retries, circuit breaking, structured logs, and Prometheus HTTP metrics.

Services are independently owned:

- `auth-service`: Google OAuth2, JWT issuance, Redis-backed sessions, logout/session validation.
- `users-service`: profiles and follow relationships backed by Postgres, plus `user.followed` events.
- `posts-service`: post CRUD and post interaction events backed by Postgres.
- `feed-service`: Redis-backed feed reads and Kafka consumers for feed fan-out.
- `notification-service`: notification reads and Kafka-driven notification creation backed by Postgres.

Infrastructure is local Compose for evaluation: Redis, Kafka/Zookeeper, three Postgres databases, Prometheus, Grafana, Loki/Promtail, and exporters.

## Trade-Offs

- Redis-backed sessions add a runtime dependency but allow logout/session invalidation beyond pure JWT verification.
- Kafka introduces eventual consistency but keeps write paths decoupled from feed and notification fan-out.
- Docker Compose favors reproducible evaluation over production high availability.
- Per-service metric helpers duplicate a small amount of code to preserve independent Go modules and avoid introducing a new shared library during stabilization.

## Challenges And Solutions

- **Google OAuth local setup:** Real browser login requires Google Console client ID, secret, and redirect URI. Deterministic tests seed JWT/Redis sessions to avoid external dependency.
- **Docker container conflicts:** Helper scripts and Compose troubleshooting document start/stop/reset flows.
- **Gateway session validation:** Protected flows require both valid JWT and active Redis session; load/E2E runners seed both.
- **Kafka async timing:** E2E scripts poll for feed and notification outcomes instead of assuming immediate consistency.
- **k6 availability:** The load runner supports Docker k6 when local k6 is missing.
- **Observability gaps:** Service-level metrics, DB dependency metrics, p99 latency, and demo simulation are now part of the stabilization package.

## Performance Evaluation

Performance evidence is maintained in `docs/report/performance-evaluation.md` and selected k6 summaries under `tests/load/reports/`.

The expected evaluation dimensions are:

- request throughput through the gateway
- p95 and p99 HTTP latency
- HTTP 5xx error rate
- business operation rate and outcomes
- DB query rate and dependency errors
- Grafana visibility during load and demo simulations

## Known Boundaries

- Local Compose is not a production HA deployment.
- Physical database sharding and clustered Redis/Kafka are documented future scaling work.
- Demo simulation must remain disabled outside observability demonstrations.
