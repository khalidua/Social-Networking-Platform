package postgres

import (
	"context"
	"database/sql"

	"social-networking-platform/users-service/internal/domain"
)

// SQLUserRepository persists profile fields in Postgres against the users table migration.
type SQLUserRepository struct {
	db *sql.DB
}

func NewSQLUserRepository(db *sql.DB) *SQLUserRepository {
	return &SQLUserRepository{db: db}
}

func (r *SQLUserRepository) Save(ctx context.Context, user domain.User) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO users (id, name, bio, profile_picture_url, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (id) DO UPDATE SET
	name = EXCLUDED.name,
	bio = EXCLUDED.bio,
	profile_picture_url = EXCLUDED.profile_picture_url,
	updated_at = NOW()
`, user.ID, user.Name, nullString(user.Bio), nullString(user.ProfilePicture))
	return err
}

func (r *SQLUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, name, COALESCE(bio, ''), COALESCE(profile_picture_url, '')
FROM users WHERE id = $1`, id)
	var u domain.User
	err := row.Scan(&u.ID, &u.Name, &u.Bio, &u.ProfilePicture)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
