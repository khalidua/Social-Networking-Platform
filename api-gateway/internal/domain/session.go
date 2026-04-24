package domain

import "time"

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}
