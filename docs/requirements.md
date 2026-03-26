# Requirements Specification

## 1. Overview
This project is a Social Networking Platform that allows users to connect, share posts, and interact with each other. The system follows a microservices architecture to ensure scalability, resilience, and maintainability.

---

## 2. Functional Requirements

### Authentication
- User can sign in using Google OAuth2
- User can log out
- System issues and validates JWT tokens

### User Management
- User has a profile (name, bio, profile picture)
- User can follow/unfollow other users
- User can view profiles

### Posts (CRUD)
- Create post
- View posts
- Update post
- Delete post

### Feed (Dashboard)
- Personalized feed based on followed users
- Near real-time updates

### Notifications
- Notifications for follows
- Notifications for post interactions

---

## 3. Non-Functional Requirements

| Attribute | Description | Target |
|----------|------------|--------|
| Scalability | Handle many users | 1000 concurrent users |
| Performance | Fast response time | < 500 ms |
| Security | Protect user data | JWT + OAuth2 |
| Resilience | System remains functional on failure | No cascading failures |
| Observability | Easy debugging | Centralized logging |
| Fault Tolerance | Retry failures | 3 retries |

---

## 4. Architectural Drivers

- Scalability
- Resilience
- Performance
- Security
- Observability