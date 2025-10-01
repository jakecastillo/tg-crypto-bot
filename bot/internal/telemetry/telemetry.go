package telemetry

import (
    "context"
    "net/http"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

// StartHealthServer starts a minimal health/readiness HTTP server exposing liveness and readiness endpoints.
func StartHealthServer(ctx context.Context, addr string, logger zerolog.Logger) {
    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })
    mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ready"))
    })

    srv := &http.Server{
        Addr:    addr,
        Handler: mux,
    }

    go func() {
        <-ctx.Done()
        if err := srv.Shutdown(context.Background()); err != nil {
            logger.Error().Err(err).Msg("failed to shutdown health server")
        }
    }()

    go func() {
        logger.Info().Str("addr", addr).Msg("starting bot health server")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal().Err(err).Msg("health server died")
        }
    }()
}
