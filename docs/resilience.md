# Resilience Strategies

## Overview
To ensure system reliability and prevent failures from spreading across services, the system implements multiple resilience mechanisms.

---

## 1. Retries
- Used for temporary failures (network issues, timeouts)
- Maximum 3 retries
- Exponential backoff applied

### Example:
Feed → Posts Service call fails → retry 3 times → fallback

---

## 2. Circuit Breaker
Prevents repeated calls to failing services.

### States:
- Closed → normal operation
- Open → block calls
- Half-open → test recovery

### Example:
If Notifications service fails:
- Circuit opens
- Feed service stops calling it

---

## 3. Fallback Mechanisms
Provide alternative responses when services fail.

### Examples:
- Feed → return cached data
- Notifications → store event in queue
- Auth → return error without crashing system

---

## 4. Asynchronous Messaging
Used to decouple services.

### Used in:
- Notifications
- Feed updates
- Follow events

### Benefit:
- Improves fault tolerance
- Prevents cascading failures

---

## 5. Rate Limiting
- 100 requests per minute per user
- Prevents system overload