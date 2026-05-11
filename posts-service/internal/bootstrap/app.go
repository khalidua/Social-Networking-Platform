package bootstrap

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"social-networking-platform/posts-service/internal/config"
	handlers "social-networking-platform/posts-service/internal/handler/http"
	postkafka "social-networking-platform/posts-service/internal/repository/kafka"
	"social-networking-platform/posts-service/internal/repository/postgres"
	"social-networking-platform/posts-service/internal/service"
	httptransport "social-networking-platform/posts-service/internal/transport/http"
)

type App struct {
	Router http.Handler
	db     *sql.DB
	pub    postkafka.PostProducer
}

func (a *App) Close() error {
	var err error
	if a.pub != nil {
		if e := a.pub.Close(); e != nil {
			err = e
		}
	}
	if a.db != nil {
		if e := a.db.Close(); e != nil {
			err = e
		}
	}
	return err
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

	postRepo := postgres.NewSQLPostRepository(db)

	var publisher postkafka.PostProducer
	kp, kerr := postkafka.NewKafkaPostProducer(cfg.KafkaBrokers, cfg.KafkaTopicPostCreated, cfg.KafkaTopicPostInteracted)
	if kerr != nil {
		log.Printf("posts-service: Kafka post producer disabled (%v); set KAFKA_BROKERS to enable events", kerr)
		publisher = postkafka.NewStubPostProducer()
	} else {
		publisher = kp
	}

	postSvc := service.NewPostService(postRepo, publisher)
	postHandler := handlers.NewPostHandler(postSvc)

	router := httptransport.NewRouter(cfg.ServiceName, postHandler)
	if router == nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize router")
	}

	return &App{Router: router, db: db, pub: publisher}, nil
}
