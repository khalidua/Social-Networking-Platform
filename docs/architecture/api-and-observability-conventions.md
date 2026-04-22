# API and Observability Conventions

This document defines the standard API response format, error codes, correlation ID handling, and structured logging pattern for the Social Networking Platform.

These conventions exist to support:
- consistent REST API behavior
- proper error handling and logging
- centralized logging
- request traceability across microservices
- easier debugging through the API Gateway and downstream services

## 1. Success Response Envelope

All successful JSON responses should use this structure:

```json
{
  "success": true,
  "data": {},
  "message": "optional human-readable message",
  "request_id": "req-123"
}
````

### Notes

* `success` is always `true`
* `data` contains the response payload
* `message` is optional
* `request_id` must be included for traceability

## 2. Error Response Envelope

All JSON error responses should use this structure:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "invalid request body",
    "details": {}
  },
  "request_id": "req-123"
}
```

### Notes

* `success` is always `false`
* `error.code` is a stable application error code
* `error.message` is human-readable
* `error.details` is optional and used for validation or extra context
* `request_id` must always be included

## 3. Standard Error Codes

Use the following application error codes:

* `BAD_REQUEST`
* `VALIDATION_ERROR`
* `UNAUTHENTICATED`
* `FORBIDDEN`
* `NOT_FOUND`
* `CONFLICT`
* `RATE_LIMITED`
* `UPSTREAM_UNAVAILABLE`
* `INTERNAL_ERROR`
* `NOT_IMPLEMENTED`

## 4. HTTP Status Mapping

Recommended mapping:

* `400` -> `BAD_REQUEST` or `VALIDATION_ERROR`
* `401` -> `UNAUTHENTICATED`
* `403` -> `FORBIDDEN`
* `404` -> `NOT_FOUND`
* `409` -> `CONFLICT`
* `429` -> `RATE_LIMITED`
* `500` -> `INTERNAL_ERROR`
* `501` -> `NOT_IMPLEMENTED`
* `503` -> `UPSTREAM_UNAVAILABLE`

## 5. Correlation / Request IDs

### Request header

Use:

`X-Request-ID`

### Rules

* If a request already contains `X-Request-ID`, preserve it
* If missing, generate one at the first entry point
* The gateway should always propagate it downstream
* Every error response should include the request ID
* Every structured log should include the request ID

## 6. Structured Logging Fields

Every request log should include at least:

* `timestamp`
* `level`
* `service`
* `request_id`
* `method`
* `path`
* `status`
* `duration_ms`
* `remote_addr`

Optional later:

* `user_id`
* `route`
* `trace_id`
* `span_id`

## 7. Logging Format

Use one-line structured JSON logs.

Example:

```json
{
  "timestamp": "2026-04-21T20:00:00Z",
  "level": "INFO",
  "service": "users-service",
  "request_id": "req-123",
  "method": "GET",
  "path": "/api/v1/users/me",
  "status": 200,
  "duration_ms": 12,
  "remote_addr": "127.0.0.1:50000"
}
```

## 8. Gateway Responsibilities

The API Gateway must:

* generate or preserve `X-Request-ID`
* include the request ID in responses
* use structured logs
* later apply auth, rate limiting, and upstream error mapping consistently



````