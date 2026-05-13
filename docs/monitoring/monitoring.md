# Monitoring & metrics

## Initial behavior

* each service exposes Prometheus metrics at `/metrics`
* endpoint uses `promhttp.Handler()` (standard Prometheus library)
* metrics include Go runtime (`go_*`, `process_*`) and custom application metrics

## Common metrics (present in every service)

* `http_requests_total` – counter with labels:
  * `service`
  * `method`
  * `route`
  * `status`
  * `status_group` (`2xx`, `4xx`, `5xx`, `other`)
* `http_request_duration_seconds` – histogram with buckets:
  * `0.1`, `0.3`, `0.5`, `1`, `2`, `5` seconds
  * same labels as `http_requests_total`
* `http_requests_active` – gauge (label: `service`)
* `service_operations_total` – counter (same labels as requests)

## Service endpoints (inside Docker network)

* `api-gateway` – port `8080`
* `auth-service` – port `8081`
* `users-service` – port `8082`
* `posts-service` – port `8083`
* `feed-service` – port `8084`
* `notification-service` – port `8085`

## Prometheus configuration example

Add to `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'api-gateway'
    static_configs:
      - targets: ['api-gateway:8080']
  - job_name: 'auth-service'
    static_configs:
      - targets: ['auth-service:8081']
  - job_name: 'users-service'
    static_configs:
      - targets: ['users-service:8082']
  - job_name: 'posts-service'
    static_configs:
      - targets: ['posts-service:8083']
  - job_name: 'feed-service'
    static_configs:
      - targets: ['feed-service:8084']
  - job_name: 'notification-service'
    static_configs:
      - targets: ['notification-service:8085']
```

## Verifying metrics

* from host (if ports published):
  ```bash
  curl http://localhost:8080/metrics | grep http_requests_total
  ```
* from inside a container:
  ```bash
  docker exec <container_name> curl http://<service>:<port>/metrics
  ```

## Grafana queries example

* request rate: `rate(http_requests_total[5m])`
* error rate (5xx): `rate(http_requests_total{status_group="5xx"}[5m])`
* p99 latency: `histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))`
* active requests: `http_requests_active`


## Example of: Check metrics directly from a service

```bash
# From host (if ports are mapped)
curl http://localhost:8080/metrics | grep http_requests_total

# Inside Docker network
docker exec -it api-gateway curl http://api-gateway:8080/metrics
```