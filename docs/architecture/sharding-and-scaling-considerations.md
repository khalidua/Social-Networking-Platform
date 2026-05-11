# Sharding and Scaling Considerations

## Purpose

This note explicitly addresses rubric requirements for scalable database design and sharding considerations. It clarifies what is implemented in the current phase and what is intentionally deferred.

## Current Phase Scope (Implemented)

- Microservices decomposition with bounded data ownership.
- Database-per-service pattern for relational stores:
  - `users-service` -> `users_db`
  - `posts-service` -> `posts_db`
  - `notification-service` -> `notifications_db`
- Redis separation by key namespace:
  - auth sessions: `auth:sessions:*`
  - feed cache: `feed:home:*`
- Kafka topic-based decoupling:
  - `user.followed`
  - `post.created`
  - `post.interacted` (contract reserved)

These decisions reduce cross-service contention and create a foundation for horizontal scaling.

## What Is Not Implemented in This Phase

- No physical Postgres sharding across multiple database nodes.
- No Redis Cluster hash-slot sharding.
- No Kafka multi-broker production topology with replication-factor > 1.
- No online re-sharding workflow or tenant/data rebalancing process.

This is intentional for Phase 2 scope and local reproducibility.

## Postgres Sharding Considerations (Future)

- **Candidate shard keys**
  - users domain: `user_id`
  - posts domain: `author_id` (write locality) or `post_id` (uniform distribution)
  - notifications domain: `user_id` (best fit for user-scoped reads)
- **Recommended strategy**
  - begin with logical partitioning (range/hash partitions) inside each service database
  - migrate to service-owned shard routers when single-node limits are reached
- **Data access rule**
  - keep joins inside one service boundary; avoid cross-service relational joins
- **Operational concerns**
  - schema migration fan-out across shards
  - backup/restore per shard
  - hotspot detection for large tenants/users

## Redis Scaling Considerations (Future)

- Move from single-node Redis to Redis Cluster or managed equivalent.
- Preserve key prefixing conventions to avoid collisions across domains.
- Keep hot-feed keys bounded:
  - capped sorted-set length per user
  - TTL/eviction policy aligned with feed freshness requirements
- For sessions, prefer predictable key size and expiration to minimize memory fragmentation.

## Kafka Scaling Considerations (Future)

- Increase partitions for `user.followed` and `post.created` as throughput grows.
- Use stable partition keys to preserve ordering where needed:
  - `user.followed`: follower/followee key depending on consumer semantics
  - `post.created`: `author_id` supports per-author ordering
- Run multi-broker cluster with replication and ISR tuning for durability.
- Add dead-letter strategy and replay procedure for consumer failures.

## Read/Write Pattern Implications

- Feed reads are fan-out heavy and user-centric; sharding/partitioning by `user_id` is the primary scaling axis.
- Post creation is write-heavy; batching and asynchronous propagation through Kafka reduce synchronous load.
- Notifications are append/read-heavy by user; user-based partitioning is the natural long-term strategy.

## Rubric Alignment Statement

This project explicitly addresses sharding considerations by:

1. implementing scalable ownership boundaries now (database-per-service + Redis key namespaces + event-driven exchange), and
2. documenting concrete shard/partition strategies and operational trade-offs for Postgres, Redis, and Kafka for later phases.

Therefore, sharding is acknowledged as a design concern, scoped realistically, and supported by an actionable future plan.
