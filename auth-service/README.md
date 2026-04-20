# auth-service

## Initial behavior

* starts service,
* exposes `/health`,
* creates placeholder route group:

  * `/api/v1/auth/login`
  * `/api/v1/auth/callback`
  * `/api/v1/auth/logout`
* contains stub auth service and stub Redis session repository interfaces,
* no real OAuth yet.