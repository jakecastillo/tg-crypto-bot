package candles

import (
	"sync"

	"github.com/rs/zerolog"
)

// RustBridge provides a stubbed adapter when the Rust collector is unavailable.
type RustBridge struct {
	stream chan Candle
	once   sync.Once
	log    zerolog.Logger
}

// NewRustBridge returns a bridge that emits no events when the Rust library is absent.
func NewRustBridge(pairs []string, logger zerolog.Logger) (*RustBridge, error) {
	b := &RustBridge{
		stream: make(chan Candle),
		log:    logger,
	}
	if len(pairs) > 0 {
		logger.Warn().Msg("uniswap bridge unavailable: built without ta_engine; stream will remain idle")
	}
	return b, nil
}

// Stream exposes the candle channel.
func (b *RustBridge) Stream() <-chan Candle {
	return b.stream
}

// Close releases resources.
func (b *RustBridge) Close() {
	b.once.Do(func() {
		close(b.stream)
	})
}
