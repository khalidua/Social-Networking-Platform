package postgres

import (
	"context"
	"database/sql"

	"social-networking-platform/posts-service/internal/domain"
)

// SQLPostRepository persists posts in Postgres.
type SQLPostRepository struct {
	db *sql.DB
}

func NewSQLPostRepository(db *sql.DB) *SQLPostRepository {
	return &SQLPostRepository{db: db}
}

func (r *SQLPostRepository) CreatePost(ctx context.Context, post *domain.Post) error {
	return r.db.QueryRowContext(ctx, `
INSERT INTO posts (id, author_id, content)
VALUES ($1, $2, $3)
RETURNING created_at, updated_at
`, post.ID, post.AuthorID, post.Content).Scan(&post.CreatedAt, &post.UpdatedAt)
}

func (r *SQLPostRepository) GetPostByID(ctx context.Context, id string) (*domain.Post, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, author_id, content, created_at, updated_at
FROM posts
WHERE id = $1
`, id)

	var post domain.Post
	if err := row.Scan(&post.ID, &post.AuthorID, &post.Content, &post.CreatedAt, &post.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &post, nil
}

func (r *SQLPostRepository) GetPostsByAuthor(ctx context.Context, authorID string) ([]domain.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, author_id, content, created_at, updated_at
FROM posts
WHERE author_id = $1
ORDER BY created_at DESC, id DESC
`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]domain.Post, 0)
	for rows.Next() {
		var post domain.Post
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Content, &post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *SQLPostRepository) UpdatePost(ctx context.Context, post *domain.Post) error {
	return r.db.QueryRowContext(ctx, `
UPDATE posts
SET content = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING created_at, updated_at
`, post.ID, post.Content).Scan(&post.CreatedAt, &post.UpdatedAt)
}

func (r *SQLPostRepository) DeletePost(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `
DELETE FROM posts
WHERE id = $1
`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
