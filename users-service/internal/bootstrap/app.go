package bootstrap

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"social-networking-platform/users-service/internal/config"
	handlers "social-networking-platform/users-service/internal/handler/http"
	userkafka "social-networking-platform/users-service/internal/repository/kafka"
	"social-networking-platform/users-service/internal/repository/postgres"
	"social-networking-platform/users-service/internal/service"
	httptransport "social-networking-platform/users-service/internal/transport/http"
)

type App struct {
	Router http.Handler
	db     *sql.DB
	pub    userkafka.FollowProducer
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

	userRepo := postgres.NewSQLUserRepository(db)
	followRepo := postgres.NewSQLFollowRepository(db)

	var publisher userkafka.FollowProducer
	kp, kerr := userkafka.NewKafkaFollowProducer(cfg.KafkaBrokers, cfg.KafkaTopicFollowed)
	if kerr != nil {
		log.Printf("users-service: Kafka follow producer disabled (%v); set KAFKA_BROKERS to enable events", kerr)
		publisher = userkafka.NewStubFollowProducer()
	} else {
		publisher = kp
	}

	userSvc := service.NewUserService(userRepo, followRepo, publisher)
	userHandler := handlers.NewUserHandler(userSvc)

	router := httptransport.NewRouter(cfg.ServiceName, userHandler)
	return &App{Router: router, db: db, pub: publisher}, nil
}
