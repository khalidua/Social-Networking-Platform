# Centralized Logging

This folder contains the local centralized logging pipeline for issue 9.1.

## Components

- Loki stores searchable logs.
- Promtail discovers Docker containers and ships container logs to Loki.
- Grafana provisions Loki as a datasource and includes a service logs dashboard.

## Log Source

The Go services already write one-line structured JSON logs to stdout. Promtail reads Docker container logs and parses these fields when present:

- `timestamp`
- `level`
- `event`
- `service`
- `request_id`
- `correlation_id`
- `method`
- `path`
- `route_group`
- `status`
- `status_group`
- `duration_ms`
- `upstream_service`

Only low-cardinality fields are promoted to Loki labels:

- Docker Compose service name
- container name
- stream
- `level`
- `event`
- `route_group`
- `status_group`
- `upstream_service`

Do not label `request_id`, `correlation_id`, raw `path`, user ids, emails, session ids, or tokens.

## Local URLs

When the compose stack is running:

- Loki: `http://localhost:3100`
- Promtail: `http://localhost:9080`
- Grafana: `http://localhost:3000`

## Run

From the repository root:

```powershell
docker compose -f deploy\compose\compose.yml up --build
```

## Grafana Queries

Open Grafana, select the Loki datasource, and run examples:

```logql
{service="api-gateway"}
{service=~"auth-service|users-service|posts-service|feed-service|notification-service"}
{level=~"WARN|ERROR"}
{route_group="posts"}
```

For request-level debugging, search by request id in the log line content rather than as a label:

```logql
{service=~".+"} |= "req-123"
```
