package store

import (
    "context"

    "github.com/jackc/pgx/v5/pgxpool"
)

// Store encapsulates database access for portfolio and trade tracking.
type Store struct {
    pool *pgxpool.Pool
}

// New returns a new Store instance.
func New(pool *pgxpool.Pool) *Store {
    return &Store{pool: pool}
}

// InsertTrade stores an executed trade.
type TradeRecord struct {
    IntentID   string
    Token      string
    Side       string
    Size       float64
    PriceUSD   float64
    TxHash     string
    ExecutedAt int64
}

// SaveTrade persists a trade record.
func (s *Store) SaveTrade(ctx context.Context, trade TradeRecord) error {
    _, err := s.pool.Exec(ctx, `
        INSERT INTO trades(intent_id, token, side, size, price_usd, tx_hash, executed_at)
        VALUES($1,$2,$3,$4,$5,$6, to_timestamp($7))
        ON CONFLICT(intent_id) DO NOTHING`,
        trade.IntentID, trade.Token, trade.Side, trade.Size, trade.PriceUSD, trade.TxHash, trade.ExecutedAt,
    )
    return err
}
