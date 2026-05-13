# ADR-004: Local Compose Baseline With Documented Production HA Boundary

## Context

The evaluation environment needs a reproducible one-command stack, while production deployment would require stronger high availability.

## Decision

Use Docker Compose as the runnable baseline and document the production-oriented target separately.

## Consequences

- The local stack is easy to run and reset.
- Compose intentionally uses single-node infrastructure.
- Production HA, replication, and shard routing are documented as future operational work rather than overclaimed as implemented.
