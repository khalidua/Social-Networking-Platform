package domain

type Post struct {
	ID        string
	AuthorID  string
	Content   string
	CreatedAt int64 // Unix milliseconds; emitted on post.created for feed ordering
}
