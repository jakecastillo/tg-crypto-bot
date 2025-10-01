package engine

import (
    "testing"
    "time"

    "github.com/shopspring/decimal"

    "github.com/example/tg-crypto-trader/risk/internal/models"
)

func TestEngineEvaluate(t *testing.T) {
    limits := map[string]models.RiskLimits{
        "WETH": {
            MaxNotionalUSD: decimal.NewFromInt(5000),
            MaxSlippageBps: 75,
            Cooldown:       1,
        },
    }

    eng := New(limits, decimal.NewFromInt(10000))

    intent := models.TradeIntent{
        Token:          "WETH",
        Size:           decimal.NewFromFloat(0.5),
        Price:          decimal.NewFromInt(2000),
        Side:           "buy",
        MaxSlippageBps: 50,
    }

    if err := eng.Evaluate(intent); err != nil {
        t.Fatalf("expected intent to pass: %v", err)
    }

    // Should fail due to cooldown
    if err := eng.Evaluate(intent); err == nil {
        t.Fatalf("expected cooldown rejection")
    }

    time.Sleep(1100 * time.Millisecond)
    eng.Release(intent)

    intent.MaxSlippageBps = 120
    if err := eng.Evaluate(intent); err == nil {
        t.Fatalf("expected slippage rejection")
    }
}
