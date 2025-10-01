package binance

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "time"

    "github.com/gorilla/websocket"
    "github.com/rs/zerolog"

    "github.com/example/tg-crypto-trader/connectors/cex"
)

// Client implements the Binance Spot TESTNET API integration.
type Client struct {
    apiKey     string
    secret     string
    restURL    string
    wsURL      string
    httpClient *http.Client
    logger     zerolog.Logger
}

// NewClient returns a configured Binance testnet client.
func NewClient(apiKey, secret string, logger zerolog.Logger) *Client {
    return &Client{
        apiKey:  apiKey,
        secret:  secret,
        restURL: "https://testnet.binance.vision",
        wsURL:   "wss://testnet.binance.vision/ws",
        httpClient: &http.Client{
            Timeout: 5 * time.Second,
        },
        logger: logger,
    }
}

// SubscribeTickers subscribes to live ticker updates.
func (c *Client) SubscribeTickers(ctx context.Context, symbols []string) (<-chan cex.Ticker, error) {
    stream := strings.ToLower(strings.Join(symbols, "@ticker/")) + "@ticker"
    u := fmt.Sprintf("%s/%s", c.wsURL, stream)
    conn, _, err := websocket.DefaultDialer.DialContext(ctx, u, nil)
    if err != nil {
        return nil, err
    }

    out := make(chan cex.Ticker)
    go func() {
        defer close(out)
        defer conn.Close()
        for {
            select {
            case <-ctx.Done():
                return
            default:
                _, message, err := conn.ReadMessage()
                if err != nil {
                    c.logger.Error().Err(err).Msg("binance ws read failed")
                    return
                }
                var payload struct {
                    Symbol string  `json:"s"`
                    Price  string  `json:"c"`
                    Time   int64   `json:"E"`
                }
                if err := json.Unmarshal(message, &payload); err != nil {
                    c.logger.Error().Err(err).Msg("binance ws decode failed")
                    continue
                }
                price, _ := strconv.ParseFloat(payload.Price, 64)
                out <- cex.Ticker{Symbol: payload.Symbol, Price: price, Timestamp: payload.Time}
            }
        }
    }()
    return out, nil
}

// PlaceOrder submits a market order on the testnet.
func (c *Client) PlaceOrder(ctx context.Context, req cex.OrderRequest) (string, error) {
    endpoint := "/api/v3/order/test"
    params := url.Values{}
    params.Set("symbol", req.Symbol)
    params.Set("side", strings.ToUpper(req.Side))
    params.Set("type", "MARKET")
    params.Set("quantity", fmt.Sprintf("%f", req.Size))
    reqURL := fmt.Sprintf("%s%s?%s", c.restURL, endpoint, params.Encode())

    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
    if err != nil {
        return "", err
    }
    httpReq.Header.Set("X-MBX-APIKEY", c.apiKey)

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return "", fmt.Errorf("binance rejected order: %d", resp.StatusCode)
    }
    return "test-order", nil
}

// CancelOrder is a stub for the testnet.
func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
    c.logger.Info().Str("order_id", orderID).Msg("cancel order noop on testnet")
    return nil
}
