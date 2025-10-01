CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS trades (
    intent_id TEXT PRIMARY KEY,
    token TEXT NOT NULL,
    side TEXT NOT NULL,
    size NUMERIC NOT NULL,
    price_usd NUMERIC NOT NULL,
    tx_hash TEXT NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

SELECT create_hypertable('trades', 'executed_at', if_not_exists => TRUE);
