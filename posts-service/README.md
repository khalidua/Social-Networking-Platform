# posts-service

The posts service owns post CRUD operations and emits `post.created` after a successful create.

## Behavior

On startup the service:

- applies SQL migrations from `migrations/`
- connects to Postgres using `DB_*` environment variables
- exposes HTTP endpoints under `/api/v1/posts`
- publishes `post.created` to Kafka when `KAFKA_BROKERS` is configured

Authenticated user identity is forwarded by the API gateway through `X-User-ID` and is copied into request context inside the service.

## Endpoints

### `POST /api/v1/posts`

Creates a post for the authenticated user.

Headers:

- `X-User-ID: user:alice`
- `Content-Type: application/json`

Request:

```json
{
  "content": "Hello from my first post"
}
```

Success response: `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "3f5a1d8d4e4f0b24b0acb5d8c9b6f4a1",
    "authorId": "user:alice",
    "content": "Hello from my first post",
    "createdAt": "2026-05-09T17:00:00Z",
    "updatedAt": "2026-05-09T17:00:00Z"
  },
  "request_id": "req-123"
}
```

### `GET /api/v1/posts/{id}`

Returns a post by id.

Success response: `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "3f5a1d8d4e4f0b24b0acb5d8c9b6f4a1",
    "authorId": "user:alice",
    "content": "Hello from my first post",
    "createdAt": "2026-05-09T17:00:00Z",
    "updatedAt": "2026-05-09T17:00:00Z"
  },
  "request_id": "req-123"
}
```

### `GET /api/v1/posts?authorId=user:alice`

Lists posts for a single author ordered by newest first.

Success response: `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "3f5a1d8d4e4f0b24b0acb5d8c9b6f4a1",
      "authorId": "user:alice",
      "content": "Hello from my first post",
      "createdAt": "2026-05-09T17:00:00Z",
      "updatedAt": "2026-05-09T17:00:00Z"
    }
  ],
  "request_id": "req-123"
}
```

### `PUT /api/v1/posts/{id}`

Updates a post. Only the author can update.

Headers:

- `X-User-ID: user:alice`
- `Content-Type: application/json`

Request:

```json
{
  "content": "Edited post content"
}
```

Success response: `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "3f5a1d8d4e4f0b24b0acb5d8c9b6f4a1",
    "authorId": "user:alice",
    "content": "Edited post content",
    "createdAt": "2026-05-09T17:00:00Z",
    "updatedAt": "2026-05-09T17:05:00Z"
  },
  "request_id": "req-123"
}
```

### `DELETE /api/v1/posts/{id}`

Deletes a post. Only the author can delete.

Headers:

- `X-User-ID: user:alice`

Success response: `204 No Content`

## Validation

- `content` is required after trimming whitespace
- `content` must be at most `2000` characters
- `authorId` query parameter is required for the author listing endpoint

## Error Format

All post errors use this shape:

```json
{
  "error": "VALIDATION_ERROR",
  "message": "content is required: validation error",
  "status": 400
}
```

Common error codes:

- `BAD_REQUEST`
- `VALIDATION_ERROR`
- `UNAUTHENTICATED`
- `FORBIDDEN`
- `NOT_FOUND`
- `INTERNAL_ERROR`

Examples:

Validation error:

```json
{
  "error": "VALIDATION_ERROR",
  "message": "content must be at most 2000 characters: validation error",
  "status": 400
}
```

Unauthorized update/delete:

```json
{
  "error": "FORBIDDEN",
  "message": "forbidden",
  "status": 403
}
```

Post not found:

```json
{
  "error": "NOT_FOUND",
  "message": "post not found",
  "status": 404
}
```

## Kafka Event

Topic:

- `post.created`

The event is published only after the post insert succeeds.

Payload:

```json
{
  "postId": "3f5a1d8d4e4f0b24b0acb5d8c9b6f4a1",
  "authorId": "user:alice",
  "content": "Hello from my first post",
  "createdAt": "2026-05-09T17:00:00Z"
}
```

## Local Commands

Run the service:

```bash
go run ./cmd/server
```

Run tests:

```bash
go test ./...
```
