package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	goredis "github.com/redis/go-redis/v9"
	"social-networking-platform/feed-service/internal/config"
	handlers "social-networking-platform/feed-service/internal/handler/http"
	usersclient "social-networking-platform/feed-service/internal/integration/users"
	kafkarepo "social-networking-platform/feed-service/internal/repository/kafka"
	redisrepo "social-networking-platform/feed-service/internal/repository/redis"
	feedservice "social-networking-platform/feed-service/internal/service"
	httptransport "social-networking-platform/feed-service/internal/transport/http"
)

type App struct {
	Router   http.Handler
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	follower kafkarepo.FollowConsumer
	post     kafkarepo.PostConsumer
	feedRepo redisrepo.FeedRepository
}

func NewApp(cfg config.Config) (*App, error) {

	redisClient := goredis.NewClient(&goredis.Options{
		Addr: "localhost:6379",
	})

	feedRepo := redisrepo.NewFeedRepository(redisClient)
	feedService := feedservice.NewFeedService(feedRepo)
	feedHandler := handlers.NewFeedHandler(feedService)
	router := httptransport.NewRouter(
		cfg.ServiceName,
		feedHandler,
	)

	if router == nil {
		return nil, fmt.Errorf("failed to initialize router")
	}

	cons, err := kafkarepo.NewFollowConsumer(cfg)
	if err != nil {
		return nil, err
	}

	var followerSrc kafkarepo.FollowerIDsProvider = kafkarepo.NopFollowerIDs{}
	if u := cfg.UsersServiceURL; u != "" {
		followerSrc = usersclient.NewClient(u)
	}
	postCons, err := kafkarepo.NewPostConsumer(cfg, feedRepo, followerSrc)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		Router:   router,
		cancel:   cancel,
		follower: cons,
		post:     postCons,
		feedRepo: feedRepo,
	}

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		if err := cons.Run(ctx); err != nil {
			log.Printf("feed-service: user.followed consumer exited: %v", err)
		}
	}()

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		if err := postCons.Run(ctx); err != nil {
			log.Printf("feed-service: post.created consumer exited: %v", err)
		}
	}()

	return app, nil
}

func (a *App) Close() error {
	if a.cancel != nil {
		a.cancel()
	}
	a.wg.Wait()
	var closeErrs []error
	if a.follower != nil {
		if err := a.follower.Close(); err != nil {
			closeErrs = append(closeErrs, err)
		}
	}
	if a.post != nil {
		if err := a.post.Close(); err != nil {
			closeErrs = append(closeErrs, err)
		}
	}
	return errors.Join(closeErrs...)
}
