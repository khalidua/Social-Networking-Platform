# notification-service

## Initial behavior

* starts service,
* exposes `/health`,
* creates placeholder route:

  * `GET /api/v1/notifications`
* on startup applies SQL migrations from `migrations/` to `notifications_db` using `DB_*` env vars.