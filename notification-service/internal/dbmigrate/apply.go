package dbmigrate

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

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
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func migrationsFileURL(absDir string) (string, error) {
	if absDir == "" {
		return "", errors.New("empty migrations directory")
	}
	if !filepath.IsAbs(absDir) {
		return "", fmt.Errorf("migrations path must be absolute: %q", absDir)
	}
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(absDir)}).String(), nil
}
