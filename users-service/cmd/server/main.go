package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "social-networking-platform/users-service/internal/bootstrap"
    "social-networking-platform/users-service/internal/config"
)

func main() {
    cfg := config.Load()

    app, err := bootstrap.NewApp(cfg)
    if err != nil {
        log.Fatalf("bootstrap failed: %v", err)
    }

    server := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      app.Router,
        ReadTimeout:  cfg.HTTP.ReadTimeout,
        WriteTimeout: cfg.HTTP.WriteTimeout,
        IdleTimeout:  cfg.HTTP.IdleTimeout,
    }

    go func() {
        log.Printf("starting %s on port %s", cfg.ServiceName, cfg.Port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server failed: %v", err)
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    <-stop

    log.Printf("shutting down %s", cfg.ServiceName)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("graceful shutdown failed: %v", err)
    }

    log.Printf("%s stopped", cfg.ServiceName)
}
