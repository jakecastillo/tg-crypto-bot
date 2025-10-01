package indicators

import (
	"errors"
	"fmt"
	"sort"

	"github.com/example/tg-crypto-trader/ta-service/internal/candles"
)

// Service computes indicator values for candle data.
type Service struct {
	candleSource CandleSource
}

// CandleSource fetches candles for a pair/interval.
type CandleSource interface {
	Candles(exchange, pair, interval string) []candles.Candle
}

// NewService constructs Service.
func NewService(source CandleSource) *Service {
	return &Service{candleSource: source}
}

// RSI calculates RSI using close prices.
func (s *Service) RSI(pair, interval string) (IndicatorResult, error) {
	series, err := s.series(pair, interval)
	if err != nil {
		return IndicatorResult{}, err
	}
	return ComputeRSI(series.Close, 14)
}

// MACD calculates MACD using close prices.
func (s *Service) MACD(pair, interval string) (MACDResult, error) {
	series, err := s.series(pair, interval)
	if err != nil {
		return MACDResult{}, err
	}
	return ComputeMACD(series.Close, 12, 26, 9)
}

// Signals returns a summary map.
func (s *Service) Signals(pair, interval string) (map[string]float64, error) {
	series, err := s.series(pair, interval)
	if err != nil {
		return nil, err
	}
	result := make(map[string]float64)
	rsi, err := ComputeRSI(series.Close, 14)
	if err == nil {
		result["rsi"] = rsi.Value
	}
	ema, err := ComputeEMA(series.Close, 21)
	if err == nil {
		result["ema21"] = ema.Value
	}
	sma, err := ComputeSMA(series.Close, 50)
	if err == nil {
		result["sma50"] = sma.Value
	}
	boll, err := ComputeBollinger(series.Close, 20, 2.0)
	if err == nil {
		result["boll_upper"] = boll.Components["upper"]
		result["boll_lower"] = boll.Components["lower"]
	}
	atr, err := ComputeATR(series.High, series.Low, series.Close, 14)
	if err == nil {
		result["atr"] = atr.Value
	}
	macd, err := ComputeMACD(series.Close, 12, 26, 9)
	if err == nil {
		result["macd"] = macd.MACD
		result["macd_signal"] = macd.Signal
		result["macd_histogram"] = macd.Histogram
	}
	if len(result) == 0 {
		return nil, errors.New("no indicators available")
	}
	return result, nil
}

type ohlcSeries struct {
	Close Series
	High  Series
	Low   Series
}

func (s *Service) series(pair, interval string) (ohlcSeries, error) {
	candles := s.candleSource.Candles("binance", pair, interval)
	if len(candles) == 0 {
		return ohlcSeries{}, fmt.Errorf("no candles for %s %s", pair, interval)
	}
	sort.Slice(candles, func(i, j int) bool { return candles[i].Start.Before(candles[j].Start) })
	series := ohlcSeries{}
	for _, c := range candles {
		series.Close = append(series.Close, c.Close)
		series.High = append(series.High, c.High)
		series.Low = append(series.Low, c.Low)
	}
	return series, nil
}
