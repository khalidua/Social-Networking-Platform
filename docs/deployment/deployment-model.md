# Deployment Model

## Local Evaluation Model

The runnable evaluation environment is Docker Compose:

- API Gateway on `localhost:8080`
- service containers on `8081` through `8085`
- Redis for sessions and feed cache
- Kafka/Zookeeper for asynchronous fan-out
- PostgreSQL per DB-backed service
- Prometheus, Grafana, Loki, Promtail, and exporters for operations visibility

Start the stack:

```powershell
powershell -ExecutionPolicy Bypass -File deploy\scripts\up.ps1 -Build
```

Check service health:

```powershell
powershell -ExecutionPolicy Bypass -File deploy\scripts\health.ps1
```

## Production-Oriented Target

The current code is designed so stateless services can scale horizontally behind a load balancer or orchestrator:

- A reverse proxy or ingress/load balancer routes public traffic to API Gateway replicas.
- Gateway replicas validate JWTs and Redis-backed sessions using the shared Redis session store.
- Auth, users, posts, feed, and notification services remain independently deployable containers.
- PostgreSQL remains owned per service boundary.
- Kafka decouples feed and notification fan-out from write paths.

## Failover And Scaling Boundaries

- Local Compose uses single-node Redis, Kafka, and PostgreSQL for reproducibility.
- Production high availability would require managed/clustered Redis, replicated Kafka, and PostgreSQL backup/failover per service.
- No physical database sharding is implemented in this phase; sharding strategy is documented in `docs/architecture/sharding-and-scaling-considerations.md`.

## Compatibility Notes

- Keep Compose service names stable because gateway upstream URLs and Prometheus scrape targets depend on them.
- Google OAuth secrets must come from local ignored env files or environment variables, not committed Compose defaults.
- Demo simulation flags are disabled by default and are only for observability demonstrations.
