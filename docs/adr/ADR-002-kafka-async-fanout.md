# ADR-002: Kafka For Asynchronous Fan-Out

## Context

Follow, post, feed, and notification workflows cross service boundaries.

## Decision

Use Kafka topics for asynchronous propagation:

- `user.followed`
- `post.created`
- `post.interacted`

Users and posts services publish domain events. Feed and notification services consume events and update their own read models.

## Consequences

- Write paths avoid synchronous calls to every downstream service.
- Tests and demos must account for eventual consistency with polling.
- Kafka outages should be visible through logs, exporters, and dashboards.
