package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/tg-crypto-trader/ta-service/internal/candles"
)

// Store wraps Postgres persistence.
type Store struct {
	pool *pgxpool.Pool
}

// New returns a new Store.
func New(ctx context.Context, url string) (*Store, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

// Close releases resources.
func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// UpsertCandle writes the candle to TimescaleDB.
func (s *Store) UpsertCandle(ctx context.Context, c candles.Candle) error {
	const q = `INSERT INTO ta_candles (exchange, pair, interval, open, high, low, close, volume, started_at)
               VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
               ON CONFLICT (exchange, pair, interval, started_at)
               DO UPDATE SET open = EXCLUDED.open, high = EXCLUDED.high, low = EXCLUDED.low, close = EXCLUDED.close, volume = EXCLUDED.volume`
	_, err := s.pool.Exec(ctx, q, c.Exchange, c.Pair, c.Interval, c.Open, c.High, c.Low, c.Close, c.Volume, c.Start)
	return err
}

// LoadCandles returns the latest candles up to limit.
func (s *Store) LoadCandles(ctx context.Context, exchange, pair, interval string, limit int) ([]candles.Candle, error) {
	const q = `SELECT exchange, pair, interval, open, high, low, close, volume, started_at
               FROM ta_candles
               WHERE exchange = $1 AND pair = $2 AND interval = $3
               ORDER BY started_at DESC
               LIMIT $4`
	rows, err := s.pool.Query(ctx, q, exchange, pair, interval, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []candles.Candle
	for rows.Next() {
		var c candles.Candle
		if err := rows.Scan(&c.Exchange, &c.Pair, &c.Interval, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume, &c.Start); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}
