package bootstrap

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strings"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"social-networking-platform/users-service/internal/config"
)

func postgresURL(cfg config.Config) string {
	u := url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		Path:   "/" + cfg.DBName,
	}
	u.User = url.UserPassword(cfg.DBUser, cfg.DBPassword)
	q := url.Values{}
	q.Set("sslmode", cfg.DBSSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

func openDatabase(cfg config.Config) (*sql.DB, error) {
	pgURL := postgresURL(cfg)
	db, err := sql.Open("postgres", pgURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(4)
	return db, nil
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

func runMigrations(pgURL, migrationsDirAbs string) error {
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
