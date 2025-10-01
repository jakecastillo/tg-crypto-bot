package models

import "github.com/shopspring/decimal"

// TradeIntent models a trade request for the risk engine.
type TradeIntent struct {
    Token          string
    Size           decimal.Decimal
    Price          decimal.Decimal
    Side           string
    MaxSlippageBps int
}

// RiskLimits aggregates per-token policy configuration.
type RiskLimits struct {
    MaxNotionalUSD decimal.Decimal
    MaxSlippageBps int
    Cooldown       int64 // seconds
}
