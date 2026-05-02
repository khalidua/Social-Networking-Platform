# API Gateway HTTPS Deployment

This document explains how the API Gateway supports secure API deployment through either:

1. TLS termination at a reverse proxy/load balancer
2. Native TLS directly in the Go gateway

## Recommended Production Path: TLS Termination

Recommended deployment:

```text
Client
  ↓ HTTPS
Load Balancer / Nginx / Ingress Controller
  ↓ HTTP inside private network
API Gateway
  ↓ HTTP inside private network
Microservices
````

In this mode, the reverse proxy owns certificates and forwards requests to the gateway over the private network.

## Required Gateway Environment

```env
TLS_ENABLED=false
TRUST_PROXY_HEADERS=true
REQUIRE_HTTPS=true
```

## Required Proxy Headers

The TLS-terminating proxy must set:

```text
X-Forwarded-Proto: https
X-Forwarded-Host: <original-host>
X-Forwarded-For: <client-ip>
X-Real-IP: <client-ip>
```

The gateway trusts these headers only when:

```env
TRUST_PROXY_HEADERS=true
```

## Why TRUST_PROXY_HEADERS Is Disabled by Default

Forwarded headers can be spoofed by clients if the gateway is directly exposed.

Only enable:

```env
TRUST_PROXY_HEADERS=true
```

when the gateway is reachable only from a trusted reverse proxy, load balancer, or ingress controller.

## Native TLS Mode

The gateway can also serve HTTPS directly.

Required env:

```env
TLS_ENABLED=true
TLS_CERT_FILE=/certs/server.crt
TLS_KEY_FILE=/certs/server.key
REQUIRE_HTTPS=true
TRUST_PROXY_HEADERS=false
```

When this mode is enabled, the server uses:

```text
ListenAndServeTLS
```

## Local Self-Signed Certificate Example

From project root:

```powershell
mkdir certs

openssl req -x509 -newkey rsa:2048 -nodes `
  -keyout certs/server.key `
  -out certs/server.crt `
  -days 365 `
  -subj "/CN=localhost"
```

Run gateway:

```powershell
cd api-gateway

$env:TLS_ENABLED="true"
$env:TLS_CERT_FILE="../certs/server.crt"
$env:TLS_KEY_FILE="../certs/server.key"
$env:REQUIRE_HTTPS="true"
$env:PORT="18443"

go run .\cmd\server
```

Test:

```powershell
curl.exe -k https://localhost:18443/health
```

## HTTP Rejection Mode Behind TLS Proxy

If using TLS termination and HTTPS is required:

```env
TLS_ENABLED=false
TRUST_PROXY_HEADERS=true
REQUIRE_HTTPS=true
```

A request without:

```text
X-Forwarded-Proto: https
```

will be rejected with:

```text
403 FORBIDDEN
```

A request with:

```text
X-Forwarded-Proto: https
```

will be accepted.

## Security Notes

* Do not enable `TRUST_PROXY_HEADERS=true` when the gateway is directly public.
* In production, expose the gateway only behind the trusted proxy.
* Keep backend service traffic on a private Docker/Kubernetes network.
* Use real certificates from the platform, ingress, or certificate manager.

````

---

## 8. Docker Compose env addition

In `deploy/compose/docker-compose.yml`, inside `api-gateway.environment`, add:

```yaml
TLS_ENABLED: false
TLS_CERT_FILE: ""
TLS_KEY_FILE: ""
TRUST_PROXY_HEADERS: true
REQUIRE_HTTPS: false
````

Recommended local Compose keeps `REQUIRE_HTTPS=false` so local HTTP testing still works.

---