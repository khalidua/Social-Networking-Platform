# Detailed implementation requirements

## A. `main.go` responsibilities for every service

Each service’s `cmd/server/main.go` must:

1. load configuration from env using `internal/config`,
2. initialize the app via `internal/bootstrap`,
3. create an HTTP server,
4. start listening on configured port,
5. log startup,
6. listen for OS termination signals,
7. gracefully shut down with timeout.

### Pseudocode shape

```go
func main() {
    cfg := config.Load()

    app, err := bootstrap.NewApp(cfg)
    if err != nil {
        log.Fatal(err)
    }

    srv := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      app.Router,
        ReadTimeout:  cfg.HTTP.ReadTimeout,
        WriteTimeout: cfg.HTTP.WriteTimeout,
        IdleTimeout:  cfg.HTTP.IdleTimeout,
    }

    go func() {
        log.Printf("starting %s on port %s", cfg.ServiceName, cfg.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    <-stop

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal(err)
    }
}
```

This should be consistent across all services.

---

## B. `config.go` responsibilities

Every service must have `internal/config/config.go` with:

* a typed `Config` struct,
* `Load()` function,
* basic validation for required env vars,
* defaults for non-sensitive values.

At scaffold time, config can be minimal, but must include:

* service name,
* port,
* environment,
* log level,
* timeouts.

Service-specific placeholders:

* gateway/auth: Redis + JWT + service URLs
* auth: Google OAuth env placeholders
* users/posts/notifications: DB placeholders
* feed: Redis + Kafka placeholders
* users/posts/feed/notifications: Kafka placeholders

This aligns with the Phase 1 architecture decisions around Redis, JWT, Kafka, and DB per service. 

---

## C. Router requirements

Each service must expose at least:

* `GET /health`

Optional but recommended:

* `GET /ready`

The router layer must:

* register middleware,
* register health route,
* register placeholder API groups,
* not contain business logic.

Use a simple HTTP router consistently across services.

### Example route grouping

For Users Service:

```go
r.Route("/api/v1/users", func(r chi.Router) {
    r.Get("/me", userHandler.GetMe)
    r.Patch("/me", userHandler.UpdateMe)
    r.Get("/{id}", userHandler.GetByID)
    r.Post("/{id}/follow", userHandler.FollowUser)
    r.Delete("/{id}/follow", userHandler.UnfollowUser)
})
```

Handlers can return `501 Not Implemented` initially except `/health`.

---

## D. Middleware requirements

Each service must include these three middleware files in scaffold form:

### `request_id.go`

* reads `X-Request-ID` from incoming headers,
* generates one if missing,
* stores it in context,
* sets it in response header.

### `logging.go`

* logs method, path, status, duration, request ID,
* structured logging preferred.

### `recovery.go`

* catches panics,
* logs them,
* returns `500`.

This prepares the codebase for the later observability/logging requirements in Phase 2 and supports the Phase 1 target that debugging should be fast and centralized.  

---

## E. Health endpoint requirements

Every service must expose:

* `GET /health`

Response should be simple and consistent, for example:

```json
{
  "status": "ok",
  "service": "users-service"
}
```

Optional `GET /ready` may later check DB/Redis/Kafka readiness, but for now it can return a placeholder healthy response.

This is important for:

* container readiness later,
* deployment validation,
* monitoring,
* CI smoke tests.

---

## F. Dependency wiring via `bootstrap/app.go`

Each service should have a `bootstrap.NewApp(cfg)` function that:

* initializes the router,
* constructs placeholder repositories/services/handlers,
* returns an app object with the router.

This creates one clean place for dependency injection and avoids manually wiring everything in `main.go`.

### Example shape

```go
type App struct {
    Router http.Handler
}

func NewApp(cfg config.Config) (*App, error) {
    router := httptransport.NewRouter(cfg)
    return &App{Router: router}, nil
}
```

Later, this function will be expanded to create:

* DB pools,
* Redis clients,
* Kafka producers/consumers,
* services,
* handlers.

---

## G. Placeholder repository/service/handler contracts

Do not leave `internal/service` and `internal/repository` empty. Add starter interfaces or stubs so the architecture is visible.

### Example for Posts Service

`internal/service/post_service.go`

```go
package service

type PostService interface {
    CreatePost() error
    GetPost() error
    UpdatePost() error
    DeletePost() error
}
```

`internal/repository/postgres/post_repository.go`

```go
package postgres

type PostRepository interface {
    // placeholder for CRUD methods
}
```

`internal/handler/http/post_handler.go`

```go
package http

import "net/http"

type PostHandler struct{}

func NewPostHandler() *PostHandler { return &PostHandler{} }

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "not implemented", http.StatusNotImplemented)
}
```

These are intentionally minimal, but they establish the service layering clearly.

---

## H. Required placeholder route matrix

Create the following route placeholders now so the API surface already matches the project plan.

### API Gateway

* `/health`
* `/api/v1/auth/*`
* `/api/v1/users/*`
* `/api/v1/posts/*`
* `/api/v1/feed`
* `/api/v1/notifications`

### Auth Service

* `GET /health`
* `GET /api/v1/auth/login`
* `GET /api/v1/auth/callback`
* `POST /api/v1/auth/logout`

### Users Service

* `GET /health`
* `GET /api/v1/users/me`
* `PATCH /api/v1/users/me`
* `GET /api/v1/users/{id}`
* `POST /api/v1/users/{id}/follow`
* `DELETE /api/v1/users/{id}/follow`

### Posts Service

* `GET /health`
* `POST /api/v1/posts`
* `GET /api/v1/posts/{id}`
* `PATCH /api/v1/posts/{id}`
* `DELETE /api/v1/posts/{id}`

### Feed Service

* `GET /health`
* `GET /api/v1/feed`

### Notification Service

* `GET /health`
* `GET /api/v1/notifications`

These placeholder routes map directly to the documented functional requirements and user stories. 

---

## I. Service-specific scaffold notes

## API Gateway

Do not add DB code here.
Add placeholder config for:

* service upstream URLs,
* Redis,
* JWT secret/session validation mode.

## Auth Service

Do not add PostgreSQL migrations unless your actual implementation later requires them.
At scaffold level, focus on:

* auth domain model,
* session repository placeholder,
* OAuth config placeholders.

This matches the Phase 1 ADR focused on Google OAuth2 + Redis-backed sessions. 

## Users / Posts / Notification Services

These should include `migrations/` and repository placeholders because they are DB-backed service types under the database-per-service decision. 

## Feed Service

Do not create persistent SQL repository placeholders as primary storage.
Use Redis and Kafka placeholder repos only, because the Phase 1 ADR explicitly says feed has no persistent database. 
