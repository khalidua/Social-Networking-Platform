# Event Contracts and Ownership

This document defines message contracts, producer ownership, consumer expectations, and versioning policy for the messaging layer.

## Versioning Rules

- Contract names use `<topic>-v<major>.json`, for example `post-created-v1.json`.
- Additive, backward-compatible changes (new optional fields) stay in the same major version.
- Breaking changes (rename/remove field, type change, semantics change) require:
  - new major contract file (for example `post-created-v2.json`)
  - a topic/version migration plan before switching producers.
- Producers must emit payloads that validate against the active contract version.
- Consumers must ignore unknown optional fields to allow additive evolution.

## `user.followed` (v1)

- Topic default: `user.followed`
- Contract: `docs/schemas/user-followed-v1.json`
- Producer owner: `users-service`
- Payload fields:
  - `follower_id` (string)
  - `followee_id` (string)
- Current consumers:
  - `feed-service` (`UserFollowedV1`)
  - `notification-service` (`UserFollowedV1`)
- Consumer expectation:
  - both ids are non-empty strings representing auth subjects/user ids.

## `post.created` (v1)

- Topic default: `post.created`
- Contract: `docs/schemas/post-created-v1.json`
- Producer owner: `posts-service`
- Payload fields:
  - `post_id` (string)
  - `author_id` (string)
  - `content` (string)
  - `created_at` (int64 unix milliseconds)
- Current consumer:
  - `feed-service` (`PostCreatedV1`)
- Consumer expectation:
  - ids are non-empty strings
  - `created_at` is epoch millis and used for feed ordering.

## `post.interacted` (v1, reserved)

- Topic default: `post.interacted`
- Contract: `docs/schemas/post-interacted-v1.json`
- Producer owner: reserved for interaction-producing service(s)
- Intended consumer:
  - `notification-service` (topic config present)
- Current status:
  - contract documented and versioned
  - producer and runtime consumer flow not yet implemented in this phase.

## Implementation Source of Truth

- Producer payload code:
  - `users-service/internal/repository/kafka/follow_producer.go`
  - `posts-service/internal/repository/kafka/post_producer.go`
- Consumer payload code:
  - `feed-service/internal/repository/kafka/follow_consumer.go`
  - `feed-service/internal/repository/kafka/post_consumer.go`
  - `notification-service/internal/repository/kafka/follow_consumer.go`
