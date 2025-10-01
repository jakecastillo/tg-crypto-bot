CREATE TABLE IF NOT EXISTS ta_candles (
    exchange TEXT NOT NULL,
    pair TEXT NOT NULL,
    interval TEXT NOT NULL,
    open NUMERIC NOT NULL,
    high NUMERIC NOT NULL,
    low NUMERIC NOT NULL,
    close NUMERIC NOT NULL,
    volume NUMERIC NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (exchange, pair, interval, started_at)
);

SELECT create_hypertable('ta_candles', 'started_at', if_not_exists => TRUE);
CREATE INDEX IF NOT EXISTS ta_candles_pair_interval_idx ON ta_candles(exchange, pair, interval, started_at DESC);
