package indicators

import (
	"errors"
	"fmt"
	"math"
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
	arr := rsiSeries(values, period)
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
	arr := emaSeries(values, period)
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
	arr := smaSeries(values, period)
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
	upper, middle, lower := bollingerBands(values, period, stddev)
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
	arr := atrSeries(high, low, close, period)
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
	macdLine, signalLine, histogram := macdSeries(values, fast, slow, signal)
	if len(macdLine) == 0 || len(signalLine) == 0 || len(histogram) == 0 {
		return MACDResult{}, errors.New("failed to compute MACD")
	}
	idx := len(values) - 1
	return MACDResult{
		MACD:         macdLine[idx],
		Signal:       signalLine[idx],
		Histogram:    histogram[idx],
		FastPeriod:   fast,
		SlowPeriod:   slow,
		SignalPeriod: signal,
	}, nil
}

func smaSeries(values Series, period int) []float64 {
	if period <= 0 || len(values) < period {
		return nil
	}
	result := make([]float64, len(values))
	windowSum := 0.0
	for i := 0; i < len(values); i++ {
		windowSum += values[i]
		if i >= period {
			windowSum -= values[i-period]
		}
		if i >= period-1 {
			result[i] = windowSum / float64(period)
		}
	}
	return result
}

func emaSeries(values Series, period int) []float64 {
	if period <= 0 || len(values) < period {
		return nil
	}
	k := 2.0 / (float64(period) + 1.0)
	result := make([]float64, len(values))
	// Seed with SMA of the first period.
	seed := 0.0
	for i := 0; i < period; i++ {
		seed += values[i]
	}
	ema := seed / float64(period)
	result[period-1] = ema
	for i := period; i < len(values); i++ {
		ema = values[i]*k + ema*(1-k)
		result[i] = ema
	}
	return result
}

func rsiSeries(values Series, period int) []float64 {
	if period <= 0 || len(values) < period+1 {
		return nil
	}
	gains := 0.0
	losses := 0.0
	for i := 1; i <= period; i++ {
		delta := values[i] - values[i-1]
		if delta > 0 {
			gains += delta
		} else {
			losses -= delta
		}
	}
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	result := make([]float64, len(values))
	rs := math.Inf(1)
	if avgLoss > 0 {
		rs = avgGain / avgLoss
	}
	result[period] = 100 - (100 / (1 + rs))

	for i := period + 1; i < len(values); i++ {
		delta := values[i] - values[i-1]
		gain := 0.0
		loss := 0.0
		if delta > 0 {
			gain = delta
		} else {
			loss = -delta
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)

		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs = avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}
	}
	return result
}

func bollingerBands(values Series, period int, stddev float64) (upper, middle, lower []float64) {
	if period <= 0 || len(values) < period {
		return nil, nil, nil
	}
	middle = smaSeries(values, period)
	if len(middle) == 0 {
		return nil, nil, nil
	}
	upper = make([]float64, len(values))
	lower = make([]float64, len(values))
	for i := period - 1; i < len(values); i++ {
		mean := middle[i]
		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := values[j] - mean
			variance += diff * diff
		}
		std := math.Sqrt(variance / float64(period))
		upper[i] = mean + stddev*std
		lower[i] = mean - stddev*std
	}
	return upper, middle, lower
}

func atrSeries(high, low, close Series, period int) []float64 {
	n := len(close)
	if period <= 0 || len(high) != n || len(low) != n || n < period {
		return nil
	}
	tr := make([]float64, n)
	for i := 0; i < n; i++ {
		highLow := high[i] - low[i]
		if i == 0 {
			tr[i] = math.Abs(highLow)
			continue
		}
		highClose := math.Abs(high[i] - close[i-1])
		lowClose := math.Abs(low[i] - close[i-1])
		tr[i] = math.Max(highLow, math.Max(highClose, lowClose))
	}

	result := make([]float64, n)
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += tr[i]
	}
	atr := sum / float64(period)
	result[period-1] = atr
	for i := period; i < n; i++ {
		atr = (atr*float64(period-1) + tr[i]) / float64(period)
		result[i] = atr
	}
	return result
}

func macdSeries(values Series, fast, slow, signal int) (Series, Series, Series) {
	if fast <= 0 || slow <= 0 || signal <= 0 || len(values) < slow+signal {
		return nil, nil, nil
	}
	fastEMA := emaSeries(values, fast)
	slowEMA := emaSeries(values, slow)
	if len(fastEMA) == 0 || len(slowEMA) == 0 {
		return nil, nil, nil
	}
	macdLine := make(Series, len(values))
	start := slow - 1
	for i := start; i < len(values); i++ {
		macdLine[i] = fastEMA[i] - slowEMA[i]
	}
	trimmed := macdLine[start:]
	signalSeries := emaSeries(trimmed, signal)
	if len(signalSeries) == 0 {
		return nil, nil, nil
	}
	signalLine := make(Series, len(values))
	for i := range signalSeries {
		signalLine[start+i] = signalSeries[i]
	}
	histogram := make(Series, len(values))
	for i := start + signal - 1; i < len(values); i++ {
		histogram[i] = macdLine[i] - signalLine[i]
	}
	return macdLine, signalLine, histogram
}
