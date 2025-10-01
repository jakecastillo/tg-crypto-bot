package main

import (
    "context"
    "net/http"
    "os/signal"
    "syscall"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"

    "github.com/example/tg-crypto-trader/api/internal/auth"
    "github.com/example/tg-crypto-trader/api/internal/config"
    "github.com/example/tg-crypto-trader/api/internal/httpapi"
    "github.com/example/tg-crypto-trader/api/internal/jobs"
    "github.com/example/tg-crypto-trader/api/internal/middleware"
    "github.com/example/tg-crypto-trader/api/internal/telemetry"
)

func main() {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
    cfg, err := config.Load()
    if err != nil {
        log.Fatal().Err(err).Msg("load config")
    }

    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    telemetry.StartMetricsServer(ctx, cfg.MetricsAddr, log.With().Str("component", "api-metrics").Logger())

    redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
    if err := redisClient.Ping(context.Background()).Err(); err != nil {
        log.Fatal().Err(err).Msg("connect redis")
    }

    dispatcher := jobs.NewDispatcher(redisClient, "trade-intents", cfg.RequestTTL, log.With().Str("component", "dispatcher").Logger())
    authz := auth.NewAuthenticator(cfg.APIToken)
    limiter := middleware.NewRateLimiter(cfg.RateLimitRPS)

    server := httpapi.NewServer(authz, limiter, dispatcher, log.With().Str("component", "api").Logger())

    srv := &http.Server{
        Addr:    cfg.HTTPAddr,
        Handler: server.Router(),
    }

    go func() {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = srv.Shutdown(shutdownCtx)
    }()

    log.Info().Str("addr", cfg.HTTPAddr).Msg("starting api server")
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal().Err(err).Msg("api server crashed")
    }
}
