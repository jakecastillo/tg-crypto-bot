package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"

	"github.com/example/tg-crypto-trader/ta-service/internal/candles"
	"github.com/example/tg-crypto-trader/ta-service/internal/indicators"
)

// IndicatorService defines the indicator computation contract.
type IndicatorService interface {
	RSI(pair, interval string) (indicators.IndicatorResult, error)
	MACD(pair, interval string) (indicators.MACDResult, error)
	Signals(pair, interval string) (map[string]float64, error)
}

// CandleProvider fetches candles for indicator calculations.
type CandleProvider interface {
	Candles(exchange, pair, interval string) []candles.Candle
}

// HTTPServer exposes indicator endpoints.
type HTTPServer struct {
	router   chi.Router
	logger   zerolog.Logger
	service  IndicatorService
	provider CandleProvider
}

// NewHTTP builds a new server.
func NewHTTP(service IndicatorService, provider CandleProvider, logger zerolog.Logger) *HTTPServer {
	r := chi.NewRouter()
	r.Use(cors.AllowAll().Handler)
	srv := &HTTPServer{router: r, service: service, provider: provider, logger: logger}
	r.Get("/healthz", srv.health)
	r.Get("/readyz", srv.health)
	r.Get("/v1/indicators/rsi/{pair}/{interval}", srv.getRSI)
	r.Get("/v1/indicators/macd/{pair}/{interval}", srv.getMACD)
	r.Get("/v1/indicators/signals/{pair}/{interval}", srv.getSignals)
	return srv
}

// Router exposes the chi router.
func (h *HTTPServer) Router() http.Handler {
	return h.router
}

func (h *HTTPServer) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *HTTPServer) getRSI(w http.ResponseWriter, r *http.Request) {
	pair := strings.ToUpper(chi.URLParam(r, "pair"))
	interval := chi.URLParam(r, "interval")
	value, err := h.service.RSI(pair, interval)
	if err != nil {
		h.respondErr(w, http.StatusBadRequest, err)
		return
	}
	h.respondJSON(w, value)
}

func (h *HTTPServer) getMACD(w http.ResponseWriter, r *http.Request) {
	pair := strings.ToUpper(chi.URLParam(r, "pair"))
	interval := chi.URLParam(r, "interval")
	value, err := h.service.MACD(pair, interval)
	if err != nil {
		h.respondErr(w, http.StatusBadRequest, err)
		return
	}
	h.respondJSON(w, value)
}

func (h *HTTPServer) getSignals(w http.ResponseWriter, r *http.Request) {
	pair := strings.ToUpper(chi.URLParam(r, "pair"))
	interval := chi.URLParam(r, "interval")
	value, err := h.service.Signals(pair, interval)
	if err != nil {
		h.respondErr(w, http.StatusBadRequest, err)
		return
	}
	candles := h.provider.Candles("binance", pair, interval)
	payload := map[string]interface{}{
		"signals": value,
		"candles": candles,
	}
	h.respondJSON(w, payload)
}

func (h *HTTPServer) respondJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *HTTPServer) respondErr(w http.ResponseWriter, status int, err error) {
	h.logger.Warn().Err(err).Int("status", status).Msg("indicator error")
	http.Error(w, err.Error(), status)
}
