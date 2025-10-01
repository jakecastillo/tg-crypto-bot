package main

import (
    "context"
    stdlog "log"
    "os"
    "time"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"

    "github.com/example/tg-crypto-trader/connectors/cex/internal/binance"
)

func main() {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
    apiKey := os.Getenv("BINANCE_TESTNET_API_KEY")
    secret := os.Getenv("BINANCE_TESTNET_SECRET")
    client := binance.NewClient(apiKey, secret, log.Logger)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    ch, err := client.SubscribeTickers(ctx, []string{"btcusdt"})
    if err != nil {
        stdlog.Fatalf("subscribe: %v", err)
    }
    for tick := range ch {
        log.Info().Str("symbol", tick.Symbol).Float64("price", tick.Price).Msg("ticker")
    }
}
