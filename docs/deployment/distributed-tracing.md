# Distributed Tracing

The platform uses W3C Trace Context propagation to follow a request across the API gateway and downstream services.

## Scope

Implemented for:

- API Gateway
- Auth Service
- Users Service
- Posts Service
- Feed Service
- Notification Service

## Trace Context

Incoming requests may include:

- `Traceparent`

If the header is missing or invalid, the receiving service generates a new trace id and span id.

Each service:

- preserves the incoming trace id
- creates a new local span id
- stores `trace_id` and `span_id` in request context
- returns a `Traceparent` response header
- writes `trace_id` and `span_id` in request logs

The API gateway forwards the current `Traceparent` header to downstream services.

## Log-Based Trace Lookup

The current tracing implementation is log-correlated through Loki/Grafana. Search by trace id in log content:

```logql
{service=~".+"} |= "trace_id=4bf92f3577b34da6a3ce929d0e0e4736"
{service=~".+"} |= "\"trace_id\":\"4bf92f3577b34da6a3ce929d0e0e4736\""
```

Do not promote `trace_id` or `span_id` to Loki labels. They are high-cardinality values.

## Manual Verification

See `tests/integration/distributed-tracing-manual-test.md`.
