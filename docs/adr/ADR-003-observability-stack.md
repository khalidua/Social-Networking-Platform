# ADR-003: Prometheus, Grafana, Loki, And Low-Cardinality Metrics

## Context

The project needs API metrics, business-operation metrics, dependency metrics, dashboards, logs, and demo visibility.

## Decision

Use Prometheus metrics exposed from each service, Grafana dashboards, Loki/Promtail logs, and infrastructure exporters. Application metrics must use low-cardinality labels only.

## Consequences

- HTTP middleware captures request rate, latency, active requests, and status groups.
- Service logic emits `business_operation_*` metrics.
- PostgreSQL repositories emit `db_query_duration_seconds` and `db_errors_total`.
- User IDs, emails, request IDs, session IDs, and raw dynamic paths are not metric labels.
