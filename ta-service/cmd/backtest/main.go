package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	"github.com/example/tg-crypto-trader/ta-service/internal/candles"
	"github.com/example/tg-crypto-trader/ta-service/internal/indicators"
)

func main() {
	file := flag.String("file", "", "path to CSV with timestamp,open,high,low,close,volume")
	flag.Parse()
	if *file == "" {
		log.Fatal("csv file required")
	}
	candlesData, err := loadCSV(*file)
	if err != nil {
		log.Fatalf("load csv: %v", err)
	}
	source := &sliceSource{candles: candlesData}
	indicatorSvc := indicators.NewService(source)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cash := 10000.0
	position := 0.0
	entry := 0.0

	for i := 50; i < len(candlesData); i++ {
		source.cursor = i
		rsi, err := indicatorSvc.RSI("BACKTEST", "1m")
		if err != nil {
			continue
		}
		price := candlesData[i].Close
		if position == 0 && rsi.Value < 30 {
			qty := math.Floor((cash/price)*1000) / 1000
			if qty <= 0 {
				continue
			}
			position = qty
			cash -= qty * price
			entry = price
			logger.Info().Float64("price", price).Msg("BUY")
		} else if position > 0 && (rsi.Value > 70 || price < entry*0.95) {
			cash += position * price
			logger.Info().Float64("price", price).Msg("SELL")
			position = 0
		}
	}
	if position > 0 {
		cash += position * candlesData[len(candlesData)-1].Close
	}
	logger.Info().Float64("final_cash", cash).Msg("backtest complete")
}

type sliceSource struct {
	candles []candles.Candle
	cursor  int
}

func (s *sliceSource) Candles(exchange, pair, interval string) []candles.Candle {
	if s.cursor < len(s.candles) {
		return append([]candles.Candle(nil), s.candles[:s.cursor+1]...)
	}
	return s.candles
}

func loadCSV(path string) ([]candles.Candle, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := csv.NewReader(bufio.NewReader(f))
	reader.FieldsPerRecord = 6
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	var out []candles.Candle
	for _, row := range rows[1:] { // skip header
		ts, _ := strconv.ParseInt(row[0], 10, 64)
		open, _ := strconv.ParseFloat(row[1], 64)
		high, _ := strconv.ParseFloat(row[2], 64)
		low, _ := strconv.ParseFloat(row[3], 64)
		closePrice, _ := strconv.ParseFloat(row[4], 64)
		vol, _ := strconv.ParseFloat(row[5], 64)
		out = append(out, candles.Candle{
			Exchange: "backtest",
			Pair:     "BACKTEST",
			Interval: "1m",
			Open:     open,
			High:     high,
			Low:      low,
			Close:    closePrice,
			Volume:   vol,
			Start:    time.Unix(ts, 0),
		})
	}
	return out, nil
}
