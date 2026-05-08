package bootstrap

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	
	_ "github.com/lib/pq"

	"social-networking-platform/users-service/internal/config"
	"social-networking-platform/users-service/internal/dbmigrate"
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

func runMigrations(pgURL, migrationsDirAbs string) error {
	if err := dbmigrate.PostgresUp(pgURL, migrationsDirAbs); err != nil {
		return err
	}
	return nil
}
