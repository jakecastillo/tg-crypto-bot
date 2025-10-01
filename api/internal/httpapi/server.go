package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"

	"github.com/example/tg-crypto-trader/api/internal/auth"
	"github.com/example/tg-crypto-trader/api/internal/jobs"
	"github.com/example/tg-crypto-trader/api/internal/middleware"
	"github.com/example/tg-crypto-trader/api/internal/ta"
)

// TradeRequest describes the trade payload expected from the bot.
type TradeRequest struct {
	Mode         string  `json:"mode"`
	Token        string  `json:"token"`
	Size         float64 `json:"size"`
	SlippageBps  int     `json:"slippage_bps"`
	Side         string  `json:"side"`
	Trigger      string  `json:"trigger"`
	PaperTrading bool    `json:"paper_trading"`
	Interval     string  `json:"interval,omitempty"`
	Force        bool    `json:"force,omitempty"`
}

// ActionRequest is a generic action payload.
type ActionRequest struct {
	Mode string `json:"mode,omitempty"`
}

// Server wraps HTTP handlers.
type Server struct {
	router     chi.Router
	dispatcher *jobs.Dispatcher
	logger     zerolog.Logger
	taClient   *ta.Client
}

// NewServer builds the HTTP router.
func NewServer(authz *auth.Authenticator, limiter *middleware.RateLimiter, dispatcher *jobs.Dispatcher, taClient *ta.Client, logger zerolog.Logger) *Server {
	r := chi.NewRouter()
	r.Use(cors.AllowAll().Handler)
	r.Use(limiter.Middleware)
	r.Use(authz.Middleware)

	srv := &Server{router: r, dispatcher: dispatcher, taClient: taClient, logger: logger}
	r.Get("/healthz", srv.health)
	r.Get("/readyz", srv.ready)
	r.Post("/v1/trades", srv.createTrade)
	r.Post("/v1/actions", srv.action)
	r.Get("/v1/ta/rsi/{pair}/{interval}", srv.rsi)
	r.Get("/v1/ta/macd/{pair}/{interval}", srv.macd)
	r.Get("/v1/ta/signals/{pair}/{interval}", srv.signals)

	return srv
}

// Router exposes the chi router.
func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) ready(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

func (s *Server) createTrade(w http.ResponseWriter, r *http.Request) {
	var req TradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.Size <= 0 || req.SlippageBps <= 0 {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	principal := r.Context().Value(auth.CtxKeyPrincipal).(string)
	intentID, err := s.dispatcher.Publish(r.Context(), principal, req)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to dispatch trade")
		http.Error(w, "failed to enqueue trade", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "queued",
		"intent_id": intentID,
		"queued_at": time.Now().UTC(),
	})
}

func (s *Server) action(w http.ResponseWriter, r *http.Request) {
	action := r.Header.Get("X-Action")
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err.Error() != "EOF" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	principal := r.Context().Value(auth.CtxKeyPrincipal).(string)
	if _, err := s.dispatcher.Publish(r.Context(), principal, map[string]interface{}{
		"action":  action,
		"payload": payload,
	}); err != nil {
		s.logger.Error().Err(err).Msg("failed to dispatch action")
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) rsi(w http.ResponseWriter, r *http.Request) {
	pair := chi.URLParam(r, "pair")
	interval := chi.URLParam(r, "interval")
	result, err := s.taClient.FetchRSI(pair, interval)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	s.writeJSON(w, result)
}

func (s *Server) macd(w http.ResponseWriter, r *http.Request) {
	pair := chi.URLParam(r, "pair")
	interval := chi.URLParam(r, "interval")
	result, err := s.taClient.FetchMACD(pair, interval)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	s.writeJSON(w, result)
}

func (s *Server) signals(w http.ResponseWriter, r *http.Request) {
	pair := chi.URLParam(r, "pair")
	interval := chi.URLParam(r, "interval")
	result, err := s.taClient.FetchSignals(pair, interval)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	s.writeJSON(w, result)
}

func (s *Server) writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
