# Local Compose Runtime

This folder contains the runnable local deployment for the Social Networking Platform.

## Services

- API Gateway: `8080`
- Auth Service: `8081`
- Users Service: `8082`
- Posts Service: `8083`
- Feed Service: `8084`
- Notification Service: `8085`
- Redis: `6379`
- Kafka host listener: `9092`
- Prometheus: `9090`
- Grafana: `3000`
- Loki: `3100`
- PostgreSQL host ports: `5433`, `5434`, `5435`

## Environment

Copy the example file and fill local-only secrets when using real Google OAuth:

```powershell
Copy-Item deploy\compose\.env.example deploy\compose\.env
```

Do not commit real Google client secrets. Dummy defaults let the stack start, but the real browser OAuth callback requires valid Google credentials and redirect URL configuration.

## Start

```powershell
powershell -ExecutionPolicy Bypass -File deploy\scripts\up.ps1 -Build
```

Equivalent raw command:

```powershell
docker compose -f deploy\compose\compose.yml up -d --build
```

## Health Check

```powershell
powershell -ExecutionPolicy Bypass -File deploy\scripts\health.ps1
```

## Logs

```powershell
powershell -ExecutionPolicy Bypass -File deploy\scripts\logs.ps1
powershell -ExecutionPolicy Bypass -File deploy\scripts\logs.ps1 -Service api-gateway
```

## Stop

```powershell
powershell -ExecutionPolicy Bypass -File deploy\scripts\down.ps1
```

## Reset Volumes

This deletes databases, Redis state, and observability storage:

```powershell
powershell -ExecutionPolicy Bypass -File deploy\scripts\reset.ps1 -ConfirmVolumeDelete
```

## Demo Simulation

The API Gateway includes disabled-by-default demo simulation flags for observability validation:

```powershell
$env:DEMO_SIMULATION_ENABLED="true"
$env:DEMO_SIMULATION_PATH="/api/v1/feed"
$env:DEMO_LATENCY="2s"
$env:DEMO_FAILURE_RATE="0.3"
docker compose -f deploy\compose\compose.yml up -d --build api-gateway
```

Set `DEMO_SIMULATION_ENABLED=false` or unset the variables for normal behavior.

## Test Harnesses

Deterministic integration and load tests seed JWT/Redis sessions instead of depending on external Google OAuth:

```powershell
powershell -ExecutionPolicy Bypass -File tests\integration\e2e\e2e-user-flow.ps1 -StartStack
powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -K6Runner docker
```

## Troubleshooting

- Container name conflicts: run `docker compose -f deploy\compose\compose.yml down` for this project before starting again.
- Port conflicts: stop the process using the host port or edit the host-side port mapping.
- Kafka first boot: wait for health checks and consumers to settle before judging feed/notification fan-out.
- OAuth `invalid_client`: verify `.env` values and Google Console redirect URI.
