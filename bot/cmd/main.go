package main

import (
    "context"
    "os/signal"
    "syscall"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"

    "github.com/example/tg-crypto-trader/bot/internal/config"
    "github.com/example/tg-crypto-trader/bot/internal/handlers"
    "github.com/example/tg-crypto-trader/bot/internal/telemetry"
)

func main() {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
    cfg, err := config.Load()
    if err != nil {
        log.Fatal().Err(err).Msg("failed to load config")
    }

    bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to create bot api")
    }
    bot.Debug = false

    logger := log.With().Str("component", "bot").Logger()

    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    telemetry.StartHealthServer(ctx, cfg.HealthAddr, logger)

    router := handlers.NewRouter(handlers.NewHTTPAPIClient(cfg.APIBaseURL, cfg.APIToken, logger), logger)

    updateCfg := tgbotapi.NewUpdate(0)
    updateCfg.Timeout = 60

    updates := bot.GetUpdatesChan(updateCfg)

    logger.Info().Msg("tg-crypto-trader bot started")

    for {
        select {
        case <-ctx.Done():
            logger.Info().Msg("shutdown signal received")
            time.Sleep(500 * time.Millisecond)
            return
        case update := <-updates:
            go router.HandleUpdate(ctx, bot, update)
        }
    }
}
