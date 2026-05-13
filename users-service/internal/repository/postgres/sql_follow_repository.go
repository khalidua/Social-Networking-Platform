package postgres

import (
	"context"
	"database/sql"
	"time"

	"social-networking-platform/users-service/internal/domain"
)

// SQLFollowRepository persists follow edges in Postgres.
type SQLFollowRepository struct {
	db *sql.DB
}

func NewSQLFollowRepository(db *sql.DB) *SQLFollowRepository {
	return &SQLFollowRepository{db: db}
}

func (r *SQLFollowRepository) Follow(ctx context.Context, rel domain.Follow) (bool, error) {
	started := time.Now()
	res, err := r.db.ExecContext(ctx, `
INSERT INTO follows (follower_id, following_id)
VALUES ($1, $2)
ON CONFLICT (follower_id, following_id) DO NOTHING
`, rel.FollowerID, rel.FolloweeID)
	observeDBOperation("insert_follow", started, err)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (r *SQLFollowRepository) Unfollow(ctx context.Context, rel domain.Follow) error {
	started := time.Now()
	_, err := r.db.ExecContext(ctx, `
DELETE FROM follows WHERE follower_id = $1 AND following_id = $2
`, rel.FollowerID, rel.FolloweeID)
	observeDBOperation("delete_follow", started, err)
	return err
}

func (r *SQLFollowRepository) ListFollowerIDs(ctx context.Context, followeeID string) ([]string, error) {
	started := time.Now()
	var opErr error
	defer func() {
		observeDBOperation("list_follower_ids", started, opErr)
	}()
	rows, err := r.db.QueryContext(ctx, `
SELECT follower_id FROM follows WHERE following_id = $1 ORDER BY follower_id
`, followeeID)
	if err != nil {
		opErr = err
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			opErr = err
			return nil, err
		}
		ids = append(ids, id)
	}
	opErr = rows.Err()
	return ids, opErr
}
