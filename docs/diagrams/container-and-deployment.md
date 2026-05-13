# Container And Deployment View

```mermaid
flowchart TB
    client["Client"]
    google["Google OAuth2"]
    gateway["api-gateway :8080"]
    auth["auth-service :8081"]
    users["users-service :8082"]
    posts["posts-service :8083"]
    feed["feed-service :8084"]
    notifications["notification-service :8085"]
    redis["Redis :6379"]
    kafka["Kafka :29092 / host :9092"]
    zookeeper["Zookeeper :2181"]
    usersdb["users-db Postgres"]
    postsdb["posts-db Postgres"]
    notificationsdb["notifications-db Postgres"]
    prometheus["Prometheus :9090"]
    grafana["Grafana :3000"]
    loki["Loki :3100"]
    promtail["Promtail"]
    exporters["Redis/Kafka/Postgres/node/cAdvisor exporters"]

    client --> gateway
    gateway --> auth
    gateway --> users
    gateway --> posts
    gateway --> feed
    gateway --> notifications
    auth --> google
    auth --> redis
    gateway --> redis
    users --> usersdb
    posts --> postsdb
    notifications --> notificationsdb
    users --> kafka
    posts --> kafka
    kafka --> feed
    kafka --> notifications
    kafka --> zookeeper
    feed --> redis
    prometheus --> gateway
    prometheus --> auth
    prometheus --> users
    prometheus --> posts
    prometheus --> feed
    prometheus --> notifications
    prometheus --> exporters
    grafana --> prometheus
    grafana --> loki
    promtail --> loki
```

## Local Deployment

- Docker Compose runs one instance of each service plus infrastructure.
- Services use Docker network names from `deploy/compose/compose.yml`.
- Prometheus and Grafana are provisioned for local observability and demo validation.
