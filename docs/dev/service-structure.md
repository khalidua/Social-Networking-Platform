# Documentation of code

## Purpose of each folder

**`cmd/server/`**
Application entrypoint. Contains `main.go` only. Responsible for starting the service.

**`internal/app/`**
Optional place for shared app-level types or application startup coordination.

**`internal/config/`**
Typed configuration loading from environment variables.

**`internal/domain/`**
Core business entities and domain models.

**`internal/handler/http/`**
HTTP handlers that translate requests/responses only.

**`internal/middleware/`**
Cross-cutting HTTP middleware such as request ID, logging, recovery, auth hooks, rate limiting hooks.

**`internal/repository/`**
Persistence or infrastructure adapters:

* PostgreSQL repos
* Redis repos
* Kafka producers/consumers

**`internal/service/`**
Business logic layer.

**`internal/transport/http/`**
Router setup, route registration, server-related HTTP wiring.

**`internal/bootstrap/`**
Application wiring and dependency assembly.

**`pkg/`**
Reusable code safe to share if needed, though overuse should be avoided.

This structure supports the project’s clean-architecture requirement and keeps later testing focused on core business logic rather than handler-heavy code.

---