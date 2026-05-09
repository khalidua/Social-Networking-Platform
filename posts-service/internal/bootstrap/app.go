package bootstrap

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"social-networking-platform/posts-service/internal/config"
	httptransport "social-networking-platform/posts-service/internal/transport/http"
)

type App struct {
	Router http.Handler
	db     *sql.DB
}

func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func NewApp(cfg config.Config) (*App, error) {
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		return nil, fmt.Errorf("working directory: %w", wdErr)
	}

	migrationsDir := cfg.MigrationsDir
	if !filepath.IsAbs(migrationsDir) {
		migrationsDir = filepath.Join(wd, migrationsDir)
	}
	migrationsAbs, err := filepath.Abs(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("migrations path: %w", err)
	}

	pgURL := postgresURL(cfg)
	if err := runMigrations(pgURL, migrationsAbs); err != nil {
		return nil, err
	}

	db, err := openDatabase(cfg)
	if err != nil {
		return nil, err
	}

	router := httptransport.NewRouter(cfg.ServiceName)
	if router == nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize router")
	}

	return &App{Router: router, db: db}, nil
}
