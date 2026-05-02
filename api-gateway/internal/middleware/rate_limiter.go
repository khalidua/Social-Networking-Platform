package middleware

import (
	"sync"
	"time"
)

type RateLimitResult struct {
	Allowed    bool
	Remaining  int
	Limit      int
	RetryAfter time.Duration
	ResetAt    time.Time
}

type userWindow struct {
	Count     int
	ResetAt   time.Time
	LastSeen  time.Time
	CreatedAt time.Time
}

type UserRateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	buckets  map[string]*userWindow
	nowFunc  func() time.Time
	stopChan chan struct{}
}

func NewUserRateLimiter(limit int, window time.Duration) *UserRateLimiter {
	if limit <= 0 {
		limit = 100
	}
	if window <= 0 {
		window = time.Minute
	}

	limiter := &UserRateLimiter{
		limit:    limit,
		window:   window,
		buckets:  make(map[string]*userWindow),
		nowFunc:  time.Now,
		stopChan: make(chan struct{}),
	}

	go limiter.cleanupLoop()

	return limiter
}

func (l *UserRateLimiter) Allow(userID string) RateLimitResult {
	now := l.nowFunc().UTC()

	l.mu.Lock()
	defer l.mu.Unlock()

	window, exists := l.buckets[userID]
	if !exists || !now.Before(window.ResetAt) {
		resetAt := now.Add(l.window)
		l.buckets[userID] = &userWindow{
			Count:     1,
			ResetAt:   resetAt,
			LastSeen:  now,
			CreatedAt: now,
		}

		return RateLimitResult{
			Allowed:    true,
			Remaining:  l.limit - 1,
			Limit:      l.limit,
			RetryAfter: 0,
			ResetAt:    resetAt,
		}
	}

	window.LastSeen = now

	if window.Count >= l.limit {
		retryAfter := time.Until(window.ResetAt)
		if retryAfter < 0 {
			retryAfter = 0
		}

		return RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			Limit:      l.limit,
			RetryAfter: retryAfter,
			ResetAt:    window.ResetAt,
		}
	}

	window.Count++
	remaining := l.limit - window.Count
	if remaining < 0 {
		remaining = 0
	}

	return RateLimitResult{
		Allowed:    true,
		Remaining:  remaining,
		Limit:      l.limit,
		RetryAfter: 0,
		ResetAt:    window.ResetAt,
	}
}

func (l *UserRateLimiter) Stop() {
	close(l.stopChan)
}

func (l *UserRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.window)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.cleanupExpired()
		case <-l.stopChan:
			return
		}
	}
}

func (l *UserRateLimiter) cleanupExpired() {
	now := l.nowFunc().UTC()

	l.mu.Lock()
	defer l.mu.Unlock()

	for userID, window := range l.buckets {
		if now.After(window.ResetAt.Add(l.window)) {
			delete(l.buckets, userID)
		}
	}
}