# Social Networking Platform

A microservices-based social networking platform built for the Software Architecture and Design Project.

## Services
- API Gateway
- Auth Service
- Users Service
- Posts Service
- Feed Service
- Notification Service

## Architecture Summary
The system follows a microservices architecture with:
- Google OAuth2 authentication
- JWT-based access
- API Gateway as the entry point
- PostgreSQL database per service where applicable
- Redis for session storage and feed caching
- Kafka for asynchronous inter-service communication

## Repository Structure
- `api-gateway/` public entry point
- `auth-service/` login and token/session logic
- `users-service/` profiles and follow relationships
- `posts-service/` post CRUD
- `feed-service/` personalized feed
- `notification-service/` notifications
- `docs/` ADRs, OpenAPI, diagrams, report material
- `deploy/` Docker, Compose, K8s, monitoring, logging assets
- `tests/` integration, contract, and load testing

## Key Design Notes
- `docs/architecture/sharding-and-scaling-considerations.md` sharding/partitioning rationale and phase scope
- `docs/messaging/event-contracts.md` versioned event ownership and payload expectations
- `docs/messaging/kafka-topics.md` broker/topic setup and local verification

## Run a service
From any service folder:

```bash
go run ./cmd/server
```

Or build it:

```bash
go build ./cmd/server
```
