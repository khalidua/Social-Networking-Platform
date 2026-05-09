package bootstrap

import (
	"fmt"
	"net/http"

	"social-networking-platform/posts-service/internal/config"
	handlers "social-networking-platform/posts-service/internal/handler/http"
	kafkarepo "social-networking-platform/posts-service/internal/repository/kafka"
	"social-networking-platform/posts-service/internal/repository/postgres"
	"social-networking-platform/posts-service/internal/service"
	httptransport "social-networking-platform/posts-service/internal/transport/http"
)

type App struct {
	Router http.Handler
	pub    kafkarepo.PostProducer
}

func NewApp(cfg config.Config) (*App, error) {
	repo := postgres.NewInMemoryPostRepository()
	pub := kafkarepo.NewPostProducer(cfg.KafkaBrokers, cfg.KafkaTopicPostCreated)
	svc := service.NewPostService(repo, pub)
	postHandler := handlers.NewPostHandler(svc)

	router := httptransport.NewRouter(cfg.ServiceName, postHandler)
	if router == nil {
		return nil, fmt.Errorf("failed to initialize router")
	}
	return &App{Router: router, pub: pub}, nil
}

func (a *App) Close() error {
	if a == nil || a.pub == nil {
		return nil
	}
	return a.pub.Close()
}
