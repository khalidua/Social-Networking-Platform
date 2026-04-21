# Local Compose Baseline

This Compose file starts the local Social Networking Platform baseline:

- API Gateway
- Auth Service
- Users Service
- Posts Service
- Feed Service
- Notification Service
- Kafka + Zookeeper
- Redis
- PostgreSQL per DB-backed service

## Run

From this folder:

```bash
docker compose up --build
````

## Stop

```bash
docker compose down
```

## Remove containers + volumes

```bash
docker compose down -v
```

## Service Ports

* API Gateway: 8080
* Auth Service: 8081
* Users Service: 8082
* Posts Service: 8083
* Feed Service: 8084
* Notification Service: 8085
* Redis: 6379
* Kafka: 9092
* Zookeeper: 2181
* Users DB: 5433
* Posts DB: 5434
* Notifications DB: 5435

---

# How to run it

Go to:

```powershell
cd "\Social-Networking-Platform\deploy\compose"
````

Copy env file if needed:

```powershell
Copy-Item .env.example .env
```

Then run:

```powershell
docker compose up --build
```

---

# How to smoke-test after startup

In another terminal:

## Gateway

```powershell
curl http://localhost:8080/health
```

## Auth

```powershell
curl http://localhost:8081/health
```

## Users

```powershell
curl http://localhost:8082/health
```

## Posts

```powershell
curl http://localhost:8083/health
```

## Feed

```powershell
curl http://localhost:8084/health
```

## Notifications

```powershell
curl http://localhost:8085/health
```

Expected shape:

```json
{"status":"ok","service":"..."}
```

---

# Acceptance criteria mapping

## Acceptance criteria

**“full stack starts locally with one command”**

This is satisfied if:

```powershell
docker compose up --build
```

successfully starts:

* infra containers
* service containers
* all service `/health` routes respond

---

# Notes about current scaffold

Because your current code is scaffold-only:

* services should start
* `/health` should work
* placeholder routes will still return `501`

That is fine for this issue. This issue is about:

* Dockerfiles
* container startup
* local integration baseline

not full business functionality yet.

---

# Common Windows troubleshooting

## If Docker says a port is busy

Stop the process using that port or change the host-side port mapping.

Example:

```yaml
ports:
  - "18080:8080"
```

## If Docker Desktop is broken

Make sure Docker Desktop is running before:

```powershell
docker compose up --build
```

## If Kafka acts slow on first boot

Give it a minute. Kafka/Zookeeper often take longer than the Go services.

---