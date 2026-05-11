package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"social-networking-platform/notification-service/internal/config"
	kafkarepo "social-networking-platform/notification-service/internal/repository/kafka"
	httptransport "social-networking-platform/notification-service/internal/transport/http"
)

type App struct {
	Router   http.Handler
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	follower kafkarepo.FollowConsumer
	db       *sql.DB
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
	cons, err := kafkarepo.NewFollowConsumer(cfg)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		Router:   router,
		cancel:   cancel,
		follower: cons,
		db:       db,
	}
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		if err := cons.Run(ctx); err != nil {
			log.Printf("notification-service: user.followed consumer exited: %v", err)
		}
	}()
	return app, nil
}

func (a *App) Close() error {
	if a.cancel != nil {
		a.cancel()
	}
	a.wg.Wait()
	if a.follower != nil {
		if err := a.follower.Close(); err != nil {
			return err
		}
	}
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			return err
		}
	}
	return nil
}
