package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"social-networking-platform/feed-service/internal/config"
	kafkarepo "social-networking-platform/feed-service/internal/repository/kafka"
	httptransport "social-networking-platform/feed-service/internal/transport/http"
	redisrepo "social-networking-platform/feed-service/internal/repository/redis"
	goredis "github.com/redis/go-redis/v9"
)

type App struct {
	Router   http.Handler
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	follower kafkarepo.FollowConsumer
	feedRepo redisrepo.FeedRepository
}

func NewApp(cfg config.Config) (*App, error) {
	router := httptransport.NewRouter(cfg.ServiceName)
	if router == nil {
		return nil, fmt.Errorf("failed to initialize router")
	}
	cons, err := kafkarepo.NewFollowConsumer(cfg)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	redisClient := goredis.NewClient(&goredis.Options{
	Addr: "localhost:6379",
	})

	feedRepo := redisrepo.NewFeedRepository(redisClient)

	app := &App{
	Router:   router,
	cancel:   cancel,
	follower: cons,
	feedRepo: feedRepo,
	}
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		if err := cons.Run(ctx); err != nil {
			log.Printf("feed-service: user.followed consumer exited: %v", err)
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
		return a.follower.Close()
	}
	return nil
}
