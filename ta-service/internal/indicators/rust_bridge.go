package indicators

/*
#cgo LDFLAGS: -L${SRCDIR}/../../rustlib/target/release -lta_engine
#include <stdlib.h>

typedef struct {
    double macd;
    double signal;
    double histogram;
    int error_code;
} macd_result;

macd_result ta_macd(double* values, int length, int fast, int slow, int signal);
*/
import "C"

import (
	"errors"
	"unsafe"

	"github.com/example/tg-crypto-trader/ta-service/internal/candles"
)

// UniswapBridge provides an adapter to the Rust Uniswap stream.
type UniswapBridge interface {
	Stream() <-chan candles.Candle
	Close()
}

// invokeMACD bridges to the Rust ta-rs implementation.
func invokeMACD(values Series, fast, slow, signal int) (MACDResult, error) {
	arr := make([]C.double, len(values))
	for i, v := range values {
		arr[i] = C.double(v)
	}
	res := C.ta_macd((*C.double)(unsafe.Pointer(&arr[0])), C.int(len(arr)), C.int(fast), C.int(slow), C.int(signal))
	if res.error_code != 0 {
		return MACDResult{}, errors.New("ta-rs macd error")
	}
	return MACDResult{
		MACD:         float64(res.macd),
		Signal:       float64(res.signal),
		Histogram:    float64(res.histogram),
		FastPeriod:   fast,
		SlowPeriod:   slow,
		SignalPeriod: signal,
	}, nil
}
