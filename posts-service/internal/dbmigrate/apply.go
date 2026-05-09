package dbmigrate

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// PostgresUp applies all SQL migrations in migrationsDirAbs (absolute path).
func PostgresUp(pgURL string, migrationsDirAbs string) error {
	fileURL, err := migrationsFileURL(migrationsDirAbs)
	if err != nil {
		return err
	}

	m, err := migrate.New(fileURL, pgURL)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

func migrationsFileURL(absDir string) (string, error) {
	p := filepath.ToSlash(absDir)
	p = strings.TrimSpace(p)
	if p == "" {
		return "", errors.New("empty migrations directory")
	}
	if len(p) > 0 && p[0] != '/' {
		p = "/" + p
	}
	return "file://" + p, nil
}
