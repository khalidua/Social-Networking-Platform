# Manual Test Plan

## Feature Being Tested

Notification retrieval for issue 7.2: authenticated users can fetch their notifications through the API Gateway.

## Preconditions

- Local Compose stack is running.
- API Gateway is reachable on `http://localhost:8080`.
- Notification Service is reachable through the gateway.
- A valid JWT/session exists for the test user, or a request can be made directly to the notification service with `X-User-ID` for service-level verification.
- The notifications database contains at least one row for the test user.

Example database setup from `deploy/compose`:

```powershell
docker exec -it snp-notifications-db psql -U postgres -d notifications_db
```

```sql
INSERT INTO notifications (id, user_id, type, message, is_read)
VALUES ('manual-notification-1', 'test-user-1', 'follow', 'test-user-2 followed you', false)
ON CONFLICT (id) DO NOTHING;
```

## Steps

1. Start the local stack:

   ```powershell
   cd "D:\Abdelrahman\Gam3a\Term 6\Software Architecture\Project\Social-Networking-Platform\deploy\compose"
   docker compose up -d
   ```

2. Verify service-level retrieval directly:

   ```powershell
   curl.exe -i http://localhost:8085/api/v1/notifications -H "X-User-ID: test-user-1"
   ```

3. Verify gateway retrieval with a valid bearer token:

   ```powershell
   curl.exe -i http://localhost:8080/api/v1/notifications -H "Authorization: Bearer <TOKEN>"
   ```

4. Verify unauthenticated direct service behavior:

   ```powershell
   curl.exe -i http://localhost:8085/api/v1/notifications
   ```

## Expected Results

- Direct service request with `X-User-ID` returns `200 OK`.
- Gateway request with a valid token returns `200 OK`.
- Response uses the standard envelope:

  ```json
  {
    "success": true,
    "data": [],
    "request_id": "req-..."
  }
  ```

- Notifications include `id`, `userId`, `type`, `message`, `read`, and `createdAt`.
- Missing `X-User-ID` at the service returns `401`.

## Edge Cases

- User with no notifications returns `200 OK` and an empty `data` array.
- Missing direct service `X-User-ID` returns `401`.
- Whitespace-only user ID is rejected.

## Failure Cases

- Database unavailable returns `500`.
- Invalid or missing gateway bearer token is rejected by the API Gateway before reaching the notification service.

## Regression Checks

- `GET /health` still returns healthy.
- Gateway still protects `/api/v1/notifications`.
- Existing notification service startup and migrations still run.
