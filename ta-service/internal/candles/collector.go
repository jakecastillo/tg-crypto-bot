package candles

import (
	"context"
	"strconv"
	"sync"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/rs/zerolog"

	"github.com/example/tg-crypto-trader/ta-service/internal/config"
	"github.com/example/tg-crypto-trader/ta-service/internal/model"
	"github.com/example/tg-crypto-trader/ta-service/internal/storage"
)

// Candle represents OHLCV data.
type Candle = model.Candle

// Buffer maintains the in-memory candle cache per pair.
type Buffer struct {
	mu      sync.RWMutex
	storage map[string][]Candle
	limit   int
}

// NewBuffer returns a Buffer with the given size.
func NewBuffer(limit int) *Buffer {
	return &Buffer{storage: make(map[string][]Candle), limit: limit}
}

// Add stores a candle and evicts old values.
func (b *Buffer) Add(c Candle) {
	key := c.Exchange + ":" + c.Pair + ":" + c.Interval
	b.mu.Lock()
	defer b.mu.Unlock()
	arr := append(b.storage[key], c)
	if len(arr) > b.limit {
		arr = arr[len(arr)-b.limit:]
	}
	b.storage[key] = arr
}

// Get returns candles for the key.
func (b *Buffer) Get(exchange, pair, interval string) []Candle {
	key := exchange + ":" + pair + ":" + interval
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]Candle, len(b.storage[key]))
	copy(out, b.storage[key])
	return out
}

// Service coordinates candle ingest.
type Service struct {
	cfg    config.Config
	logger zerolog.Logger
	store  *storage.Store
	buffer *Buffer
	rust   UniswapBridge
}

// UniswapBridge exposes the Uniswap candle stream.
type UniswapBridge interface {
	Stream() <-chan Candle
	Close()
}

// NewService returns a Service.
func NewService(cfg config.Config, store *storage.Store, bridge UniswapBridge, logger zerolog.Logger) *Service {
	return &Service{
		cfg:    cfg,
		store:  store,
		buffer: NewBuffer(cfg.CandleLimit),
		rust:   bridge,
		logger: logger,
	}
}

// Start begins Binance and Uniswap streaming.
func (s *Service) Start(ctx context.Context) {
	go s.runBinance(ctx)
	go s.runUniswap(ctx)
}

func (s *Service) runBinance(ctx context.Context) {
	if len(s.cfg.BinanceSymbols) == 0 {
		return
	}
	wsHandler := func(event *binance.WsKlineEvent) {
		if event == nil {
			return
		}
		candle := Candle{
			Exchange: "binance",
			Pair:     event.Symbol,
			Interval: event.Kline.Interval,
			Open:     parseFloat(event.Kline.Open),
			High:     parseFloat(event.Kline.High),
			Low:      parseFloat(event.Kline.Low),
			Close:    parseFloat(event.Kline.Close),
			Volume:   parseFloat(event.Kline.Volume),
			Start:    time.UnixMilli(event.Kline.StartTime),
		}
		s.buffer.Add(candle)
		if err := s.store.UpsertCandle(ctx, candle); err != nil {
			s.logger.Error().Err(err).Msg("failed to persist binance candle")
		}
	}
	errHandler := func(err error) {
		s.logger.Error().Err(err).Msg("binance ws error")
	}
	for _, symbol := range s.cfg.BinanceSymbols {
		go func(sym string) {
			done, _, err := binance.WsKlineServe(sym, s.cfg.Interval, wsHandler, errHandler)
			if err != nil {
				s.logger.Error().Err(err).Str("symbol", sym).Msg("binance ws failed")
				return
			}
			<-ctx.Done()
			done <- struct{}{}
		}(symbol)
	}
}

func (s *Service) runUniswap(ctx context.Context) {
	if len(s.cfg.UniswapPairs) == 0 {
		return
	}
	if s.rust == nil {
		s.logger.Warn().Msg("uniswap bridge disabled")
		return
	}
	stream := s.rust.Stream()
	go func() {
		<-ctx.Done()
		s.rust.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case candle, ok := <-stream:
			if !ok {
				s.logger.Warn().Msg("uniswap stream closed")
				return
			}
			s.buffer.Add(candle)
			if err := s.store.UpsertCandle(context.Background(), candle); err != nil {
				s.logger.Error().Err(err).Msg("failed to persist uniswap candle")
			}
		}
	}
}

// Candles returns the latest candles for pair/interval.
func (s *Service) Candles(exchange, pair, interval string) []Candle {
	return s.buffer.Get(exchange, pair, interval)
}

func parseFloat(input string) float64 {
	f, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0
	}
	return f
}
