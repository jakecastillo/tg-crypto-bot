package indicators

import (
	"errors"
	"fmt"

	talib "github.com/markcheno/go-talib"
)

// Series represents a float time series.
type Series []float64

// IndicatorResult contains the latest computed values.
type IndicatorResult struct {
	Value      float64            `json:"value"`
	Components map[string]float64 `json:"components,omitempty"`
}

// MACDResult carries MACD-specific values.
type MACDResult struct {
	MACD         float64 `json:"macd"`
	Signal       float64 `json:"signal"`
	Histogram    float64 `json:"histogram"`
	FastPeriod   int     `json:"fast_period"`
	SlowPeriod   int     `json:"slow_period"`
	SignalPeriod int     `json:"signal_period"`
}

// ComputeRSI calculates the Relative Strength Index.
func ComputeRSI(values Series, period int) (IndicatorResult, error) {
	if len(values) < period+1 {
		return IndicatorResult{}, errors.New("not enough data for RSI")
	}
	arr := talib.Rsi(values, period)
	if len(arr) == 0 {
		return IndicatorResult{}, errors.New("failed to compute RSI")
	}
	return IndicatorResult{Value: arr[len(arr)-1]}, nil
}

// ComputeEMA calculates Exponential Moving Average.
func ComputeEMA(values Series, period int) (IndicatorResult, error) {
	if len(values) < period {
		return IndicatorResult{}, errors.New("not enough data for EMA")
	}
	arr := talib.Ema(values, period)
	if len(arr) == 0 {
		return IndicatorResult{}, errors.New("failed to compute EMA")
	}
	return IndicatorResult{Value: arr[len(arr)-1]}, nil
}

// ComputeSMA calculates Simple Moving Average.
func ComputeSMA(values Series, period int) (IndicatorResult, error) {
	if len(values) < period {
		return IndicatorResult{}, errors.New("not enough data for SMA")
	}
	arr := talib.Sma(values, period)
	if len(arr) == 0 {
		return IndicatorResult{}, errors.New("failed to compute SMA")
	}
	return IndicatorResult{Value: arr[len(arr)-1]}, nil
}

// ComputeBollinger calculates Bollinger Bands.
func ComputeBollinger(values Series, period int, stddev float64) (IndicatorResult, error) {
	if len(values) < period {
		return IndicatorResult{}, errors.New("not enough data for Bollinger")
	}
	upper, middle, lower := talib.BBands(values, period, stddev, stddev, talib.Sma)
	if len(upper) == 0 {
		return IndicatorResult{}, errors.New("failed to compute bands")
	}
	idx := len(upper) - 1
	return IndicatorResult{
		Value: middle[idx],
		Components: map[string]float64{
			"upper":  upper[idx],
			"lower":  lower[idx],
			"middle": middle[idx],
		},
	}, nil
}

// ComputeATR calculates Average True Range.
func ComputeATR(high, low, close Series, period int) (IndicatorResult, error) {
	if len(high) < period || len(low) < period || len(close) < period {
		return IndicatorResult{}, errors.New("not enough data for ATR")
	}
	arr := talib.Atr(high, low, close, period)
	if len(arr) == 0 {
		return IndicatorResult{}, errors.New("failed to compute ATR")
	}
	return IndicatorResult{Value: arr[len(arr)-1]}, nil
}

// ComputeMACD delegates to the ta-rs bridge.
func ComputeMACD(values Series, fast, slow, signal int) (MACDResult, error) {
	if len(values) < slow+signal {
		return MACDResult{}, fmt.Errorf("not enough data for MACD")
	}
	return invokeMACD(values, fast, slow, signal)
}
