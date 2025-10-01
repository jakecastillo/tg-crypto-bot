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
	FetchRSI(ctx context.Context, pair, interval string) (float64, error)
	FetchMACD(ctx context.Context, pair, interval string) (MACDResponse, error)
	FetchSignals(ctx context.Context, pair, interval string) (map[string]float64, error)
	SetAutoTradeFilter(ctx context.Context, expression, interval string, enabled bool) error
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
	Interval       string  `json:"interval,omitempty"`
	Force          bool    `json:"force,omitempty"`
}

type MACDResponse struct {
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
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

func (c *HTTPAPIClient) FetchRSI(ctx context.Context, pair, interval string) (float64, error) {
	var resp struct {
		Value float64 `json:"value"`
	}
	if err := c.get(ctx, "/v1/ta/rsi/"+pair+"/"+interval, &resp); err != nil {
		return 0, err
	}
	return resp.Value, nil
}

func (c *HTTPAPIClient) FetchMACD(ctx context.Context, pair, interval string) (MACDResponse, error) {
	var resp MACDResponse
	if err := c.get(ctx, "/v1/ta/macd/"+pair+"/"+interval, &resp); err != nil {
		return MACDResponse{}, err
	}
	return resp, nil
}

func (c *HTTPAPIClient) FetchSignals(ctx context.Context, pair, interval string) (map[string]float64, error) {
	var resp struct {
		Signals map[string]float64 `json:"signals"`
	}
	if err := c.get(ctx, "/v1/ta/signals/"+pair+"/"+interval, &resp); err != nil {
		return nil, err
	}
	return resp.Signals, nil
}

func (c *HTTPAPIClient) SetAutoTradeFilter(ctx context.Context, expression, interval string, enabled bool) error {
	payload := map[string]interface{}{
		"expression": expression,
		"interval":   interval,
		"enabled":    enabled,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.SendAction(ctx, "set-autotrade-filter", string(body))
}

func (c *HTTPAPIClient) get(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ta query failed: %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
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
		r.reply(ctx, bot, msg.Chat.ID, "Commands:\n/buy <pair> <size> <slippage%>\n/sell <pair> <size> <slippage%>\n/forcebuy <pair> <size> <slippage%>\n/rsi <pair> <interval>\n/macd <pair> <interval>\n/signals <pair> <interval>\n/autotrade <on|off> [expr] [interval]\n/mode <paper|live>\n/portfolio")
	case "buy", "sell":
		r.handleTrade(ctx, bot, msg)
	case "forcebuy":
		r.handleForceTrade(ctx, bot, msg)
	case "mode":
		r.handleMode(ctx, bot, msg)
	case "portfolio":
		r.handlePortfolio(ctx, bot, msg)
	case "rsi":
		r.handleRSI(ctx, bot, msg)
	case "macd":
		r.handleMACD(ctx, bot, msg)
	case "signals":
		r.handleSignals(ctx, bot, msg)
	case "autotrade":
		r.handleAutoTrade(ctx, bot, msg)
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
		Interval:    "1m",
	}

	if err := r.api.CreateTrade(ctx, intent); err != nil {
		r.logger.Error().Err(err).Msg("failed to create trade")
		r.reply(ctx, bot, msg.Chat.ID, "Trade rejected: "+err.Error())
		return
	}

	r.reply(ctx, bot, msg.Chat.ID, fmt.Sprintf("Submitted %s %s %.4f with max %.2f%% slippage", msg.Command(), token, size, slippagePct))
}

func (r *Router) handleForceTrade(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	parts := strings.Fields(msg.CommandArguments())
	if len(parts) < 3 {
		r.reply(ctx, bot, msg.Chat.ID, "Usage: /forcebuy <pair> <size> <slippage%>")
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
		Side:        "buy",
		Trigger:     "force",
		Interval:    "1m",
		Force:       true,
	}
	if err := r.api.CreateTrade(ctx, intent); err != nil {
		r.logger.Error().Err(err).Msg("failed force trade")
		r.reply(ctx, bot, msg.Chat.ID, "Force trade rejected: "+err.Error())
		return
	}
	r.reply(ctx, bot, msg.Chat.ID, fmt.Sprintf("Force buy submitted for %s %.4f", token, size))
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

func (r *Router) handleRSI(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	pair, interval, ok := parsePairInterval(msg.CommandArguments())
	if !ok {
		r.reply(ctx, bot, msg.Chat.ID, "Usage: /rsi <pair> <interval>")
		return
	}
	value, err := r.api.FetchRSI(ctx, pair, interval)
	if err != nil {
		r.reply(ctx, bot, msg.Chat.ID, "RSI unavailable: "+err.Error())
		return
	}
	r.reply(ctx, bot, msg.Chat.ID, fmt.Sprintf("RSI %s %s → %.2f", pair, interval, value))
}

func (r *Router) handleMACD(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	pair, interval, ok := parsePairInterval(msg.CommandArguments())
	if !ok {
		r.reply(ctx, bot, msg.Chat.ID, "Usage: /macd <pair> <interval>")
		return
	}
	value, err := r.api.FetchMACD(ctx, pair, interval)
	if err != nil {
		r.reply(ctx, bot, msg.Chat.ID, "MACD unavailable: "+err.Error())
		return
	}
	msgText := fmt.Sprintf("MACD %s %s → MACD %.4f | Signal %.4f | Hist %.4f", pair, interval, value.MACD, value.Signal, value.Histogram)
	r.reply(ctx, bot, msg.Chat.ID, msgText)
}

func (r *Router) handleSignals(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	pair, interval, ok := parsePairInterval(msg.CommandArguments())
	if !ok {
		r.reply(ctx, bot, msg.Chat.ID, "Usage: /signals <pair> <interval>")
		return
	}
	signals, err := r.api.FetchSignals(ctx, pair, interval)
	if err != nil {
		r.reply(ctx, bot, msg.Chat.ID, "Signals unavailable: "+err.Error())
		return
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Signals %s %s:\n", pair, interval))
	for k, v := range signals {
		b.WriteString(fmt.Sprintf("• %s: %.4f\n", strings.ToUpper(k), v))
	}
	r.reply(ctx, bot, msg.Chat.ID, b.String())
}

func (r *Router) handleAutoTrade(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	parts := strings.Fields(msg.CommandArguments())
	if len(parts) == 0 {
		r.reply(ctx, bot, msg.Chat.ID, "Usage: /autotrade <on|off> [expression] [interval]")
		return
	}
	mode := strings.ToLower(parts[0])
	switch mode {
	case "off":
		if err := r.api.SetAutoTradeFilter(ctx, "", "", false); err != nil {
			r.reply(ctx, bot, msg.Chat.ID, "Failed to disable auto-trade: "+err.Error())
			return
		}
		r.reply(ctx, bot, msg.Chat.ID, "Auto-trade filters disabled")
	case "on":
		if len(parts) < 2 {
			r.reply(ctx, bot, msg.Chat.ID, "Usage: /autotrade on <expression> [interval]")
			return
		}
		expression := strings.ToLower(parts[1])
		interval := "1m"
		if len(parts) >= 3 {
			interval = parts[2]
		}
		if err := r.api.SetAutoTradeFilter(ctx, expression, interval, true); err != nil {
			r.reply(ctx, bot, msg.Chat.ID, "Failed to enable auto-trade: "+err.Error())
			return
		}
		r.reply(ctx, bot, msg.Chat.ID, fmt.Sprintf("Auto-trade filter enabled: %s @ %s", expression, interval))
	default:
		r.reply(ctx, bot, msg.Chat.ID, "Usage: /autotrade <on|off> [expression] [interval]")
	}
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

func parsePairInterval(args string) (string, string, bool) {
	parts := strings.Fields(args)
	if len(parts) < 2 {
		return "", "", false
	}
	return strings.ToUpper(parts[0]), parts[1], true
}
