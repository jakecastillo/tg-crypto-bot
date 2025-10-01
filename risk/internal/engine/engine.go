package engine

import (
    "errors"
    "sync"
    "time"

    "github.com/shopspring/decimal"

    "github.com/example/tg-crypto-trader/risk/internal/models"
)

// ErrRiskRejected indicates a trade was denied by policy checks.
var ErrRiskRejected = errors.New("trade rejected by risk policy")

// Engine enforces per-token and global risk limits.
type Engine struct {
    mu            sync.Mutex
    limits        map[string]models.RiskLimits
    cooldownState map[string]time.Time
    maxPortfolio  decimal.Decimal
    exposure      decimal.Decimal
}

// New creates an Engine instance.
func New(limits map[string]models.RiskLimits, maxPortfolio decimal.Decimal) *Engine {
    return &Engine{
        limits:        limits,
        maxPortfolio:  maxPortfolio,
        cooldownState: make(map[string]time.Time),
        exposure:      decimal.Zero,
    }
}

// Evaluate validates a trade intent against configured limits.
func (e *Engine) Evaluate(intent models.TradeIntent) error {
    e.mu.Lock()
    defer e.mu.Unlock()

    notional := intent.Size.Mul(intent.Price)
    if e.exposure.Add(notional).GreaterThan(e.maxPortfolio) {
        return ErrRiskRejected
    }

    limit, ok := e.limits[intent.Token]
    if !ok {
        return ErrRiskRejected
    }

    if notional.GreaterThan(limit.MaxNotionalUSD) {
        return ErrRiskRejected
    }

    if intent.MaxSlippageBps > limit.MaxSlippageBps {
        return ErrRiskRejected
    }

    if until, exists := e.cooldownState[intent.Token]; exists {
        if time.Now().Before(until) {
            return ErrRiskRejected
        }
    }

    if limit.Cooldown > 0 {
        e.cooldownState[intent.Token] = time.Now().Add(time.Duration(limit.Cooldown) * time.Second)
    }

    e.exposure = e.exposure.Add(notional)
    return nil
}

// Release reduces tracked exposure after a trade settles.
func (e *Engine) Release(intent models.TradeIntent) {
    e.mu.Lock()
    defer e.mu.Unlock()
    notional := intent.Size.Mul(intent.Price)
    e.exposure = e.exposure.Sub(notional)
    if e.exposure.IsNegative() {
        e.exposure = decimal.Zero
    }
}
