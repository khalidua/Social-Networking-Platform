# api-gateway

## Initial behavior

* starts an HTTP server,
* loads config,
* registers `/health`,
* registers placeholder route groups:

  * `/api/v1/auth`
  * `/api/v1/users`
  * `/api/v1/posts`
  * `/api/v1/feed`
  * `/api/v1/notifications`
* applies request ID, logging, and recovery middleware,
* shuts down gracefully.

Got it — I’ve read through the entire **api-gateway** code carefully. No guessing, no shortcuts. Here’s a precise understanding of what you currently have (this will guide all upcoming implementations so we don’t break anything).

---
---

# Current Gateway Capabilities (Ground Truth)

## 1. Entry Point

* `main.go`:

  * Loads config
  * Bootstraps app
  * Starts HTTP server with timeouts
  * Graceful shutdown implemented correctly

---

## 2. Architecture Style

You’re following a **clean + layered microservice gateway design**:

```
cmd → bootstrap → transport/router → handlers → middleware → domain/repo/security
```

---

### API Response Standardization

* `apiresponse`:

  * `SuccessEnvelope`
  * `ErrorEnvelope`
  * Includes:

    * `success`
    * `data` / `error`
    * `request_id`

---

### Error Codes

* Centralized in `apperrors`
* Includes:

  * `UNAUTHENTICATED`
  * `NOT_IMPLEMENTED`
  * `UPSTREAM_UNAVAILABLE`
  * etc.

---

###  Middleware Stack

* `RequestID` → injects `X-Request-ID`
* `Logging` → structured JSON logs
* `Recovery` → panic-safe

---

### Auth & Session Validation

Inside `ProxyHandler`:

Flow:

1. Extract Bearer token
2. Parse JWT
3. Load session from Redis
4. Validate:

   * exists
   * matches user
   * not expired

---

### Reverse Proxy (Core Gateway Logic)

Using:

```go
httputil.NewSingleHostReverseProxy
```

Features implemented:

* Path forwarding
* Query forwarding
* Header injection:

  * `X-Request-ID`
  * `X-User-ID`
  * `X-User-Email`
  * `X-Session-ID`
* Error handling → returns standardized error


> “Route auth, user, post, feed, notification requests”

---

### Redis (Custom Implementation)

* Raw TCP Redis client (no external lib)
* Session storage:

  ```
  auth:sessions:<sessionID>
  ```
---

### Config System

* Fully env-driven
* Includes:

  * service URLs
  * Redis
  * JWT
  * timeouts

---

### Tests

* `proxy_handler_test.go`

  * Tests:

    * header forwarding
    * auth validation
    * session validation

---