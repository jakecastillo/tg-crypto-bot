package cex

import "context"

// Ticker represents a market ticker update.
type Ticker struct {
    Symbol    string
    Price     float64
    Timestamp int64
}

// OrderRequest describes a request to place an order.
type OrderRequest struct {
    Symbol string
    Side   string
    Size   float64
    Price  float64
}

// Connector defines the interface exchanges must implement.
type Connector interface {
    SubscribeTickers(ctx context.Context, symbols []string) (<-chan Ticker, error)
    PlaceOrder(ctx context.Context, req OrderRequest) (string, error)
    CancelOrder(ctx context.Context, orderID string) error
}
