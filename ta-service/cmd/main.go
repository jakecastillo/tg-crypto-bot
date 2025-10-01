package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/example/tg-crypto-trader/ta-service/internal/candles"
	"github.com/example/tg-crypto-trader/ta-service/internal/config"
	"github.com/example/tg-crypto-trader/ta-service/internal/indicators"
	"github.com/example/tg-crypto-trader/ta-service/internal/server"
	"github.com/example/tg-crypto-trader/ta-service/internal/storage"
	"github.com/example/tg-crypto-trader/ta-service/internal/ws"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	store, err := storage.New(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect postgres")
	}
	defer store.Close()

	bridge, err := indicators.NewRustBridge(cfg.UniswapPairs, log.With().Str("component", "uniswap-bridge").Logger())
	if err != nil {
		log.Warn().Err(err).Msg("uniswap bridge disabled")
	}

	candleSvc := candles.NewService(cfg, store, bridge, log.With().Str("component", "candles").Logger())
	candleSvc.Start(ctx)

	indicatorSvc := indicators.NewService(candleSvc)

	httpSrv := server.NewHTTP(indicatorSvc, candleSvc, log.With().Str("component", "http").Logger())
	wsHub := ws.NewHub(indicatorSvc, candleSvc, log.With().Str("component", "ws").Logger())

	mux := http.NewServeMux()
	mux.Handle("/", httpSrv.Router())
	mux.HandleFunc("/ws", wsHub.Handle)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		wsHub.Broadcast(map[string]string{"status": "shutdown"})
	}()

	log.Info().Str("addr", cfg.HTTPAddr).Msg("ta-service listening")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("http server error")
	}
}
