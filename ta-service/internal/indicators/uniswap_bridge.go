package indicators

/*
#cgo LDFLAGS: -L${SRCDIR}/../../rustlib/target/release -lta_engine
#include <stdlib.h>

typedef struct {
    double open;
    double high;
    double low;
    double close;
    double volume;
    long long timestamp_ms;
    char* pair;
} candle;

void uniswap_free_string(char*);
void uniswap_stop(void* handle);
void* uniswap_start(const char* pairs_json);
int uniswap_poll(void* handle, candle* out);
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"time"
	"unsafe"

	"github.com/rs/zerolog"

	"github.com/example/tg-crypto-trader/ta-service/internal/candles"
)

// RustBridge wraps the FFI bridge for Uniswap candles.
type RustBridge struct {
	handle unsafe.Pointer
	stream chan candles.Candle
	log    zerolog.Logger
}

// NewRustBridge starts the Rust collector.
func NewRustBridge(pairs []string, logger zerolog.Logger) (*RustBridge, error) {
	if len(pairs) == 0 {
		return &RustBridge{stream: make(chan candles.Candle)}, nil
	}
	payload, err := json.Marshal(pairs)
	if err != nil {
		return nil, err
	}
	cJSON := C.CString(string(payload))
	defer C.free(unsafe.Pointer(cJSON))
	handle := C.uniswap_start(cJSON)
	if handle == nil {
		return nil, fmt.Errorf("failed to start uniswap collector")
	}
	b := &RustBridge{
		handle: handle,
		stream: make(chan candles.Candle, 1024),
		log:    logger,
	}
	go b.loop()
	return b, nil
}

func (b *RustBridge) loop() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if b.handle == nil {
				close(b.stream)
				return
			}
			var raw C.candle
			if C.uniswap_poll(b.handle, &raw) == 0 {
				continue
			}
			pair := C.GoString(raw.pair)
			C.uniswap_free_string(raw.pair)
			c := candles.Candle{
				Exchange: "uniswap",
				Pair:     pair,
				Interval: "event",
				Open:     float64(raw.open),
				High:     float64(raw.high),
				Low:      float64(raw.low),
				Close:    float64(raw.close),
				Volume:   float64(raw.volume),
				Start:    time.UnixMilli(int64(raw.timestamp_ms)),
			}
			select {
			case b.stream <- c:
			default:
				b.log.Warn().Msg("uniswap bridge backpressure")
			}
		}
	}
}

// Stream returns the candle channel.
func (b *RustBridge) Stream() <-chan candles.Candle {
	return b.stream
}

// Close stops the bridge.
func (b *RustBridge) Close() {
	if b.handle != nil {
		C.uniswap_stop(b.handle)
		b.handle = nil
	}
}
