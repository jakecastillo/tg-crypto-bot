package telemetry

import (
    "context"
    "net/http"

    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/rs/zerolog"
)

// StartMetricsServer exposes Prometheus metrics at /metrics.
func StartMetricsServer(ctx context.Context, addr string, logger zerolog.Logger) {
    srv := &http.Server{
        Addr:    addr,
        Handler: promhttp.Handler(),
    }

    go func() {
        <-ctx.Done()
        _ = srv.Shutdown(context.Background())
    }()

    go func() {
        logger.Info().Str("addr", addr).Msg("starting metrics server")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Error().Err(err).Msg("metrics server failed")
        }
    }()
}
