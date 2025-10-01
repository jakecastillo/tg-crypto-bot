package handlers

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "strings"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/rs/zerolog"
)

// APIClient abstracts the REST interface exposed by the api service.
type APIClient interface {
    CreateTrade(ctx context.Context, payload TradeIntent) error
    SendAction(ctx context.Context, action, payload string) error
}

// TradeIntent mirrors the API payload for trade execution requests.
type TradeIntent struct {
    Mode           string  `json:"mode"`
    Token          string  `json:"token"`
    Size           float64 `json:"size"`
    SlippageBps    int     `json:"slippage_bps"`
    Side           string  `json:"side"`
    Trigger        string  `json:"trigger"`
    PaperTrading   bool    `json:"paper_trading"`
    CopySourceID   string  `json:"copy_source_id,omitempty"`
    RiskPresetName string  `json:"risk_preset_name,omitempty"`
}

// HTTPAPIClient implements APIClient using net/http.
type HTTPAPIClient struct {
    baseURL string
    token   string
    client  *http.Client
    logger  zerolog.Logger
}

// NewHTTPAPIClient returns a new HTTP API client.
func NewHTTPAPIClient(baseURL, token string, logger zerolog.Logger) *HTTPAPIClient {
    return &HTTPAPIClient{
        baseURL: strings.TrimRight(baseURL, "/"),
        token:   token,
        client: &http.Client{
            Timeout: 10 * time.Second,
        },
        logger: logger,
    }
}

func (c *HTTPAPIClient) CreateTrade(ctx context.Context, payload TradeIntent) error {
    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("marshal intent: %w", err)
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/trades", bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("build request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    if c.token != "" {
        req.Header.Set("Authorization", "Bearer "+c.token)
    }

    resp, err := c.client.Do(req)
    if err != nil {
        return fmt.Errorf("exec request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return fmt.Errorf("api returned status %d", resp.StatusCode)
    }

    return nil
}

func (c *HTTPAPIClient) SendAction(ctx context.Context, action, payload string) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/actions", strings.NewReader(payload))
    if err != nil {
        return fmt.Errorf("build action request: %w", err)
    }
    req.Header.Set("X-Action", action)
    if c.token != "" {
        req.Header.Set("Authorization", "Bearer "+c.token)
    }
    resp, err := c.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 300 {
        return fmt.Errorf("action failed: %d", resp.StatusCode)
    }
    return nil
}

// Router routes Telegram commands to API actions.
type Router struct {
    api    APIClient
    logger zerolog.Logger
}

// NewRouter constructs a Router.
func NewRouter(api APIClient, logger zerolog.Logger) *Router {
    return &Router{api: api, logger: logger}
}

// HandleUpdate dispatches telegram updates to command handlers.
func (r *Router) HandleUpdate(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    if update.Message == nil {
        return
    }

    msg := update.Message
    if !msg.IsCommand() {
        r.reply(ctx, bot, msg.Chat.ID, "Send /help for usage.")
        return
    }

    switch msg.Command() {
    case "start":
        r.reply(ctx, bot, msg.Chat.ID, "Welcome to tg-crypto-trader. Use /buy or /sell to execute trades.")
    case "help":
        r.reply(ctx, bot, msg.Chat.ID, "Commands:\n/buy <token> <size> <slippage%>\n/sell <token> <size> <slippage%>\n/mode <paper|live>\n/portfolio")
    case "buy", "sell":
        r.handleTrade(ctx, bot, msg)
    case "mode":
        r.handleMode(ctx, bot, msg)
    case "portfolio":
        r.handlePortfolio(ctx, bot, msg)
    default:
        r.reply(ctx, bot, msg.Chat.ID, "Unknown command. Use /help.")
    }
}

func (r *Router) handleTrade(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    parts := strings.Fields(msg.CommandArguments())
    if len(parts) < 3 {
        r.reply(ctx, bot, msg.Chat.ID, "Usage: /buy <token> <size> <slippage%>")
        return
    }

    token := strings.ToUpper(parts[0])
    size, err := parseFloat(parts[1])
    if err != nil {
        r.reply(ctx, bot, msg.Chat.ID, "invalid size")
        return
    }

    slippagePct, err := parseFloat(strings.TrimSuffix(parts[2], "%"))
    if err != nil {
        r.reply(ctx, bot, msg.Chat.ID, "invalid slippage")
        return
    }

    intent := TradeIntent{
        Mode:        "market",
        Token:       token,
        Size:        size,
        SlippageBps: int(slippagePct * 100),
        Side:        msg.Command(),
        Trigger:     "manual",
    }

    if err := r.api.CreateTrade(ctx, intent); err != nil {
        r.logger.Error().Err(err).Msg("failed to create trade")
        r.reply(ctx, bot, msg.Chat.ID, "Trade rejected: "+err.Error())
        return
    }

    r.reply(ctx, bot, msg.Chat.ID, fmt.Sprintf("Submitted %s %s %.4f with max %.2f%% slippage", msg.Command(), token, size, slippagePct))
}

func (r *Router) handleMode(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    mode := strings.TrimSpace(msg.CommandArguments())
    if mode != "paper" && mode != "live" {
        r.reply(ctx, bot, msg.Chat.ID, "Mode must be 'paper' or 'live'")
        return
    }
    payload := fmt.Sprintf(`{"mode":"%s"}`, mode)
    if err := r.api.SendAction(ctx, "set-mode", payload); err != nil {
        r.logger.Error().Err(err).Msg("failed to toggle mode")
        r.reply(ctx, bot, msg.Chat.ID, "Failed to switch mode")
        return
    }
    r.reply(ctx, bot, msg.Chat.ID, fmt.Sprintf("Mode set to %s", mode))
}

func (r *Router) handlePortfolio(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    if err := r.api.SendAction(ctx, "portfolio", "{}"); err != nil {
        r.logger.Error().Err(err).Msg("failed portfolio request")
        r.reply(ctx, bot, msg.Chat.ID, "Portfolio unavailable")
        return
    }
    r.reply(ctx, bot, msg.Chat.ID, "Portfolio request submitted. Check the app for details.")
}

func (r *Router) reply(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, message string) {
    msg := tgbotapi.NewMessage(chatID, message)
    msg.ParseMode = "Markdown"
    if _, err := bot.Send(msg); err != nil {
        r.logger.Error().Err(err).Msg("failed to send reply")
    }
}

func parseFloat(v string) (float64, error) {
    return strconv.ParseFloat(strings.TrimSpace(v), 64)
}
