# Monitoring Stack

This folder contains the local Prometheus and Grafana configuration for issue 9 observability work.

## Components

- Prometheus scrapes gateway and service `/metrics` endpoints.
- Prometheus scrapes Redis, Kafka, PostgreSQL, host, and container exporters.
- Grafana provisions the Prometheus datasource and the Social Networking Platform dashboard.
- Prometheus loads alert rules for service availability, HTTP errors, latency, Redis, PostgreSQL, Kafka, consumer lag, and host memory usage.

## Local URLs

When the compose stack is running:

- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`
- cAdvisor: `http://localhost:8086`
- node-exporter: `http://localhost:9100`
- redis-exporter: `http://localhost:9121`
- kafka-exporter: `http://localhost:9308`
- users-db exporter: `http://localhost:9187`
- posts-db exporter: `http://localhost:9188`
- notifications-db exporter: `http://localhost:9189`

Grafana local credentials:

- Username: `admin`
- Password: `admin`

## Application Metrics

Each Go HTTP service exposes:

- `http_requests_total`
- `http_request_duration_seconds`
- `http_requests_active`
- `service_operations_total`

Labels are intentionally low-cardinality:

- `service`
- `method`
- `route`
- `status`

Do not add labels containing user ids, request ids, emails, session ids, tokens, or raw paths with resource ids.

## Run

From the repository root:

```powershell
docker compose -f deploy\compose\compose.yml up --build
```

Then open Grafana and select the `Social Networking Platform Overview` dashboard.

## Alert Rule Validation

In Prometheus, open `Status > Rules` and confirm the rule group `social-networking-platform-critical` is loaded.

Useful Prometheus expressions:

```promql
up
sum by (service) (rate(http_requests_total[5m]))
histogram_quantile(0.95, sum by (service, le) (rate(http_request_duration_seconds_bucket[5m])))
kafka_consumergroup_lag
```
