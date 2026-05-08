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

## API documentation

Gateway-facing routes for profiles and follows are documented in `docs/openapi/swagger.yaml` (paths under **Users**) and match `SuccessEnvelope` / error envelopes used by handlers.

## Tests

* **Unit (default):** `go test ./...` exercises service logic and HTTP handlers (`internal/service`, `internal/handler/http`).
* **Integration (Postgres required):**
  ```bash
  set INTEGRATION_PG_DSN=postgres://postgres:postgres@localhost:5433/users_db?sslmode=disable
  go test -tags=integration -count=1 ./internal/integration/
  ```
  Run from this module root with Docker Compose Postgres on port **5433** (or adjust the DSN).