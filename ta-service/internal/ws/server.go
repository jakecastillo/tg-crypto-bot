package ws

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"github.com/example/tg-crypto-trader/ta-service/internal/candles"
	"github.com/example/tg-crypto-trader/ta-service/internal/indicators"
)

// Hub broadcasts indicator snapshots to clients.
type Hub struct {
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]struct{}
	mu       sync.Mutex
	service  *indicators.Service
	provider *candles.Service
	logger   zerolog.Logger
}

// NewHub builds a WebSocket hub.
func NewHub(service *indicators.Service, provider *candles.Service, logger zerolog.Logger) *Hub {
	return &Hub{
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		clients:  make(map[*websocket.Conn]struct{}),
		service:  service,
		provider: provider,
		logger:   logger,
	}
}

// Handle upgrades HTTP connections.
func (h *Hub) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("ws upgrade failed")
		return
	}
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
	go h.writer(conn)
}

func (h *Hub) writer(conn *websocket.Conn) {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
	}()
	for range ticker.C {
		// send aggregated snapshot for default symbol set
		snapshot := make(map[string]interface{})
		for _, pair := range []string{"ETHUSDT", "BTCUSDT"} {
			if value, err := h.service.Signals(pair, "1m"); err == nil {
				snapshot[pair] = value
			}
		}
		payload, _ := json.Marshal(snapshot)
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			h.logger.Warn().Err(err).Msg("ws write failed")
			return
		}
	}
}

// Broadcast sends the payload to all clients.
func (h *Hub) Broadcast(payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			h.logger.Warn().Err(err).Msg("broadcast failed")
			conn.Close()
			delete(h.clients, conn)
		}
	}
}
