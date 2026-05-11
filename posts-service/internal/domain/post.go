package domain

import "time"

type Post struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"authorId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PostInteraction struct {
	PostID          string `json:"postId"`
	PostAuthorID    string `json:"postAuthorId"`
	ActorID         string `json:"actorId"`
	InteractionType string `json:"interactionType"`
	CreatedAt       int64  `json:"createdAt"`
}
