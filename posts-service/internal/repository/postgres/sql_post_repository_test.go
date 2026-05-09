package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"social-networking-platform/posts-service/internal/domain"
)

func TestSQLPostRepository_CreatePost(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLPostRepository(db)
	createdAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(30 * time.Second)
	post := &domain.Post{
		ID:       "post-1",
		AuthorID: "author-1",
		Content:  "hello world",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
INSERT INTO posts (id, author_id, content)
VALUES ($1, $2, $3)
RETURNING created_at, updated_at
`)).
		WithArgs(post.ID, post.AuthorID, post.Content).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(createdAt, updatedAt))

	if err := repo.CreatePost(context.Background(), post); err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if !post.CreatedAt.Equal(createdAt) {
		t.Fatalf("createdAt = %v, want %v", post.CreatedAt, createdAt)
	}
	if !post.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("updatedAt = %v, want %v", post.UpdatedAt, updatedAt)
	}
	assertExpectations(t, mock)
}

func TestSQLPostRepository_GetPostByID(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLPostRepository(db)
	createdAt := time.Date(2026, 5, 9, 11, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Minute)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id, author_id, content, created_at, updated_at
FROM posts
WHERE id = $1
`)).
		WithArgs("post-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "author_id", "content", "created_at", "updated_at"}).
			AddRow("post-1", "author-1", "hello", createdAt, updatedAt))

	post, err := repo.GetPostByID(context.Background(), "post-1")
	if err != nil {
		t.Fatalf("GetPostByID: %v", err)
	}
	if post == nil {
		t.Fatal("expected post, got nil")
	}
	if post.AuthorID != "author-1" || post.Content != "hello" {
		t.Fatalf("unexpected post: %+v", post)
	}
	assertExpectations(t, mock)
}

func TestSQLPostRepository_GetPostByIDNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLPostRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id, author_id, content, created_at, updated_at
FROM posts
WHERE id = $1
`)).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	post, err := repo.GetPostByID(context.Background(), "missing")
	if err != nil {
		t.Fatalf("GetPostByID: %v", err)
	}
	if post != nil {
		t.Fatalf("expected nil post, got %+v", post)
	}
	assertExpectations(t, mock)
}

func TestSQLPostRepository_GetPostsByAuthor(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLPostRepository(db)
	firstCreatedAt := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	secondCreatedAt := firstCreatedAt.Add(-1 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id, author_id, content, created_at, updated_at
FROM posts
WHERE author_id = $1
ORDER BY created_at DESC, id DESC
`)).
		WithArgs("author-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "author_id", "content", "created_at", "updated_at"}).
			AddRow("post-2", "author-1", "newest", firstCreatedAt, firstCreatedAt).
			AddRow("post-1", "author-1", "older", secondCreatedAt, secondCreatedAt))

	posts, err := repo.GetPostsByAuthor(context.Background(), "author-1")
	if err != nil {
		t.Fatalf("GetPostsByAuthor: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("len(posts) = %d, want 2", len(posts))
	}
	if posts[0].ID != "post-2" || posts[1].ID != "post-1" {
		t.Fatalf("unexpected order: %+v", posts)
	}
	assertExpectations(t, mock)
}

func TestSQLPostRepository_UpdatePost(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLPostRepository(db)
	createdAt := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	post := &domain.Post{
		ID:      "post-1",
		Content: "edited",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
UPDATE posts
SET content = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING created_at, updated_at
`)).
		WithArgs(post.ID, post.Content).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(createdAt, updatedAt))

	if err := repo.UpdatePost(context.Background(), post); err != nil {
		t.Fatalf("UpdatePost: %v", err)
	}
	if !post.CreatedAt.Equal(createdAt) || !post.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("unexpected timestamps: %+v", post)
	}
	assertExpectations(t, mock)
}

func TestSQLPostRepository_UpdatePostNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLPostRepository(db)
	post := &domain.Post{
		ID:      "missing",
		Content: "edited",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
UPDATE posts
SET content = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING created_at, updated_at
`)).
		WithArgs(post.ID, post.Content).
		WillReturnError(sql.ErrNoRows)

	err = repo.UpdatePost(context.Background(), post)
	if err != sql.ErrNoRows {
		t.Fatalf("UpdatePost error = %v, want %v", err, sql.ErrNoRows)
	}
	assertExpectations(t, mock)
}

func TestSQLPostRepository_DeletePost(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLPostRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
DELETE FROM posts
WHERE id = $1
`)).
		WithArgs("post-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.DeletePost(context.Background(), "post-1"); err != nil {
		t.Fatalf("DeletePost: %v", err)
	}
	assertExpectations(t, mock)
}

func TestSQLPostRepository_DeletePostNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLPostRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
DELETE FROM posts
WHERE id = $1
`)).
		WithArgs("missing").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.DeletePost(context.Background(), "missing")
	if err != sql.ErrNoRows {
		t.Fatalf("DeletePost error = %v, want %v", err, sql.ErrNoRows)
	}
	assertExpectations(t, mock)
}

func assertExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet: %v", err)
	}
}
