# Kafka Broker and Topics

This document defines Kafka broker setup and initial topic naming for local development.

## Broker Setup (Local)

Kafka is provisioned in Docker Compose:

- Service: `kafka`
- Image: `confluentinc/cp-kafka:7.5.0`
- Internal broker endpoint for services: `kafka:29092`
- Host endpoint for local tools: `localhost:9092`
- Dependency: `zookeeper`

Reference: `deploy/compose/compose.yml`

## Initial Topics

- `user.followed`
  - Producer: `users-service`
  - Consumers: `feed-service`, `notification-service`
  - Purpose: notify downstream services when a new follow relation is created.

- `post.created`
  - Producer: `posts-service`
  - Consumer: `feed-service`
  - Purpose: fan-out post IDs into follower home feeds.

- `post.interacted`
  - Producer: reserved/not yet implemented in current phase
  - Consumer: `notification-service` (topic config already present)
  - Purpose: future notification triggers for reactions/interactions.

## Topic Configuration via Environment Variables

Configured in service env:

- `KAFKA_BROKERS`
- `KAFKA_TOPIC_USER_FOLLOWED`
- `KAFKA_TOPIC_POST_CREATED`
- `KAFKA_TOPIC_POST_INTERACTED`

Default local values (compose):

- `KAFKA_BROKERS=kafka:29092`
- `KAFKA_TOPIC_USER_FOLLOWED=user.followed`
- `KAFKA_TOPIC_POST_CREATED=post.created`
- `KAFKA_TOPIC_POST_INTERACTED=post.interacted`

## Local Verification

Use the smoke test section in `tests/integration/smoke/manual-smoke-test.md`:

- Verify broker readiness.
- Produce and consume sample events on configured topics.
- Verify service-level producer/consumer exchange.

## Related Contract Docs

- `docs/messaging/event-contracts.md`
- `docs/schemas/user-followed-v1.json`
- `docs/schemas/post-created-v1.json`
- `docs/schemas/post-interacted-v1.json`
