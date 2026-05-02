## 1. Run the slow mock users service

Open **Terminal 1** from project root:

```powershell
go run .\scripts\dev\mock-slow-users-service.go
```

Expected:

```text
slow users mock running on :18082
```

---

## 2. Run API Gateway with short upstream timeout

Open **Terminal 2**:

```powershell
cd .\api-gateway

$env:SERVICE_NAME="api-gateway"
$env:APP_ENV="test"
$env:PORT="18080"
$env:AUTH_SERVICE_URL="http://localhost:18081"
$env:USERS_SERVICE_URL="http://localhost:18082"
$env:POSTS_SERVICE_URL="http://localhost:18083"
$env:FEED_SERVICE_URL="http://localhost:18084"
$env:NOTIFICATION_SERVICE_URL="http://localhost:18085"
$env:REDIS_HOST="localhost"
$env:REDIS_PORT="16379"
$env:JWT_SECRET="secret"
$env:JWT_ISSUER="auth-service"
$env:UPSTREAM_TIMEOUT="2s"

go run .\cmd\server
```

---

## 3. Test timeout through gateway

Because `/api/v1/users/test` is protected, this test requires a valid token/session. For a quick timeout-only test, temporarily make `ProxyUsers` public just for testing:

```go
func (h *ProxyHandler) ProxyUsers(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, h.usersServiceURL, nil)
}
```

Then run:

```powershell
curl.exe -i http://localhost:18080/api/v1/users/test
```

Expected after around **2 seconds**:

```text
HTTP/1.1 502 Bad Gateway
```

Body should include:

```json
{
  "success": false,
  "error": {
    "code": "UPSTREAM_UNAVAILABLE",
    "message": "upstream service is unavailable"
  }
}
```

---

## 4. Important: revert test change

After confirming timeout works, restore `ProxyUsers` back to protected:

```go
func (h *ProxyHandler) ProxyUsers(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.requireAuthenticated(w, r)
	if !ok {
		return
	}
	h.proxyRequest(w, r, h.usersServiceURL, claims)
}
```

Do **not** commit the public `ProxyUsers` version.