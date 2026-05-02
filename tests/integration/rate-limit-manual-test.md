# API Gateway Rate Limiting Manual Test

This test verifies issue **2.3 — Per-user rate limiting at gateway**.

---

## Goal

Confirm that the API Gateway enforces a per-user request limit and returns a consistent `429 RATE_LIMITED` response when the user exceeds the configured limit.

---

## Test Configuration

For this test, use a small limit:

```powershell
$env:RATE_LIMIT_PER_MINUTE="2"
$env:RATE_LIMIT_WINDOW="1m"
````

Expected behavior:

* Request 1 → forwarded
* Request 2 → forwarded
* Request 3 → rejected with `429 RATE_LIMITED`

---

## Prerequisites

You need:

1. API Gateway running
2. Redis running
3. A valid JWT for the same user
4. A valid Redis session matching that JWT
5. A downstream users-service (mock) running

---

## Step 1 — Start Redis

```powershell
docker run -d --name snp-gateway-test-redis -p 16379:6379 redis:7-alpine
```

Insert session:

```powershell
docker exec snp-gateway-test-redis redis-cli SET auth:sessions:session-1 '{\"id\":\"session-1\",\"user_id\":\"google:123\",\"email\":\"user@example.com\",\"expires_at\":\"2030-01-01T00:00:00Z\"}'
```

---

## Step 2 — Generate JWT

```powershell
go run .\scripts\dev\generate-token.go
```

Copy output into:

```powershell
$TOKEN="<PASTE_VALID_JWT_HERE>"
```

---

## Step 3 — Start mock users-service

```powershell
go run .\scripts\dev\mock-slow-users-service.go
```

Make sure it runs on:

```text
:19082
```

---

## Step 4 — Start API Gateway

```powershell
cd .\api-gateway

$env:SERVICE_NAME="api-gateway"
$env:APP_ENV="test"
$env:PORT="18080"
$env:USERS_SERVICE_URL="http://localhost:19082"
$env:AUTH_SERVICE_URL="http://localhost:18081"
$env:POSTS_SERVICE_URL="http://localhost:18083"
$env:FEED_SERVICE_URL="http://localhost:18084"
$env:NOTIFICATION_SERVICE_URL="http://localhost:18085"
$env:REDIS_HOST="localhost"
$env:REDIS_PORT="16379"
$env:JWT_SECRET="secret"
$env:JWT_ISSUER="auth-service"
$env:UPSTREAM_TIMEOUT="2s"
$env:RATE_LIMIT_PER_MINUTE="2"
$env:RATE_LIMIT_WINDOW="1m"

go run .\cmd\server
```

---

## ⚠️ Important Before Testing

Restart the gateway **before sending requests** to reset limiter state.

---

## Step 5 — Send Requests

### Request 1

```powershell
curl.exe -i -H "Authorization: Bearer $TOKEN" -H "X-Request-ID: req-rate-limit-1" http://localhost:18080/api/v1/users/test
```

Expected:

```text
HTTP/1.1 200 OK
```

---

### Request 2

```powershell
curl.exe -i -H "Authorization: Bearer $TOKEN" -H "X-Request-ID: req-rate-limit-2" http://localhost:18080/api/v1/users/test
```

Expected:

```text
HTTP/1.1 200 OK
```

---

### Request 3

```powershell
curl.exe -i -H "Authorization: Bearer $TOKEN" -H "X-Request-ID: req-rate-limit-3" http://localhost:18080/api/v1/users/test
```

Expected:

```text
HTTP/1.1 429 Too Many Requests
```

---

## Expected Response Body (3rd request)

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMITED",
    "message": "rate limit exceeded",
    "details": {
      "limit": 2,
      "window_seconds": 60,
      "retry_after_seconds": 59
    }
  },
  "request_id": "req-rate-limit-3"
}
```

---

## Expected Headers

```text
X-RateLimit-Limit: 2
X-RateLimit-Remaining: 0
X-RateLimit-Reset: <unix_timestamp>
Retry-After: <seconds>
```

---

## Check Gateway Logs

You should see:

```json
{
  "timestamp": "...",
  "level": "WARN",
  "event": "rate_limit_exceeded",
  "service": "api-gateway",
  "request_id": "req-rate-limit-3",
  "user_id": "google:123",
  "method": "GET",
  "path": "/api/v1/users/test",
  "limit": 2,
  "reset_at": "..."
}
```

---

## Pass Criteria

This test passes if:

* first request → forwarded (200)
* second request → forwarded (200)
* third request → rejected (429)
* error code = `RATE_LIMITED`
* response uses standard envelope
* rate-limit headers exist
* structured log is printed

---

## Notes

* Limiter is **in-memory** → restart resets counters
* Limit is applied **per authenticated user**
* User identity comes from:

  ```text
  JWT → claims.Subject
  ```
* Failed upstream calls (e.g., 502) **still count toward rate limit**