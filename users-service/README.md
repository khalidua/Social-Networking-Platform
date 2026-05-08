# user-service

## Initial behavior

* starts service,
* exposes `/health`,
* creates placeholder route group:

  * `GET /api/v1/users/me`
  * `PATCH /api/v1/users/me`
  * `GET /api/v1/users/{id}`
  * `POST /api/v1/users/{id}/follow`
  * `DELETE /api/v1/users/{id}/follow`

On startup the service applies SQL migrations (see `migrations/`), connects to Postgres using `DB_*` env vars, and publishes `user.followed` to Kafka when configured (`KAFKA_BROKERS`). User primary keys are TEXT so they can match auth subjects forwarded as `X-User-ID` via the API gateway. **If you previously created the DB with the old UUID-based migration, drop/recreate the users volume before running this version.**