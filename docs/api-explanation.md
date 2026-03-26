# API Communication Strategy

## Overview
The system uses a combination of synchronous (REST) and asynchronous (event-driven) communication.

---

## 1. Synchronous Communication (REST APIs)

### Used For:
- Authentication
- User profiles
- CRUD operations (Posts)

### Flow:
Client → API Gateway → Service → Response

### Reason:
- Immediate response required
- Simpler interaction

---

## 2. Asynchronous Communication (Event-Driven)

### Used For:
- Notifications
- Feed updates
- Follow events

### Flow:
Service → Message Broker → Consumer Service

### Example:
- User creates post → event sent → Feed updated
- User follows someone → notification event

---

## 3. Why Hybrid Approach?

| Type | Advantage |
|------|----------|
| Sync | Fast, simple |
| Async | Scalable, fault-tolerant |

---

## 4. Message Broker
- Kafka or RabbitMQ
- Ensures reliable delivery of events

---

## 5. Benefits
- Loose coupling
- Improved scalability
- Better fault tolerance