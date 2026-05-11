# Manual Test Plan

## Feature Being Tested

Issue 7 notification event flow:

- `users-service` produces `user.followed` events.
- `posts-service` produces `post.interacted` events for supported post likes.
- `notification-service` consumes both topics, persists notifications, and exposes them through `GET /api/v1/notifications`.
- Consumer failures are retried and sent to `<topic>.dlq` without blocking upstream user or post actions.

## Preconditions

- Docker Desktop is running.
- The local compose stack is available from `deploy/compose`.
- The notification database migration has created `notifications(id,user_id,type,message,is_read,created_at)`.
- Valid JWT/session setup exists for two users, or service-level requests can be made with gateway-forwarded `X-User-ID` headers in a local test environment.

## Steps

1. Start the local stack:

   ```powershell
   docker compose -f deploy\compose\compose.yml up --build
   ```

2. Authenticate or prepare two local users, `user:alice` and `user:bob`.

3. Create or ensure a follow action from `user:bob` to `user:alice` through the existing follow API.

4. Read Alice notifications:

   ```powershell
   Invoke-RestMethod -Method Get `
     -Uri http://localhost:8080/api/v1/notifications `
     -Headers @{ Authorization = "Bearer <alice-jwt>" }
   ```

5. Create a post as Alice through the posts API.

6. Like Alice's post as Bob:

   ```powershell
   Invoke-RestMethod -Method Post `
     -Uri http://localhost:8080/api/v1/posts/<post-id>/interactions `
     -Headers @{ Authorization = "Bearer <bob-jwt>"; "Content-Type" = "application/json" } `
     -Body '{"interaction_type":"like"}'
   ```

7. Read Alice notifications again.

8. Optional failure-handling check: stop `notification-service`, produce a malformed `post.interacted` message or restart with a temporary database failure, then inspect notification-service logs and the `<topic>.dlq` Kafka topic after retries.

## Expected Results

- The follow action returns success from users-service and is not blocked by notification persistence.
- Alice receives a `follow` notification with `read=false` and a populated `createdAt`.
- Bob's like returns `202 Accepted`.
- Alice receives a `post_like` notification for Bob's interaction.
- Re-reading notifications returns newest notifications first.
- Failed notification event processing is logged and moved to the configured DLQ topic after retry.

## Edge Cases

- Self-follow events are ignored by notification-service.
- Self-like attempts are rejected by posts-service and ignored by notification-service if somehow received.
- Unsupported interaction types are rejected and sent to DLQ after retry.
- Missing `X-User-ID` on notification retrieval returns an unauthorized error.

## Failure Cases

- Kafka unavailable: upstream services keep business actions non-blocking and log publish/consumer errors.
- Notification database unavailable: consumers retry and publish failing events to DLQ.
- Malformed events: consumers log validation errors, publish to DLQ, and continue consuming later events.

## Regression Checks

- `POST /api/v1/posts` still publishes `post.created` for feed fan-out.
- Existing post create/read/update/delete behavior remains unchanged.
- `GET /api/v1/notifications` still uses the standard success/error envelope.
- Existing `user.followed` feed consumer behavior remains unchanged.
