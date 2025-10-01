package ta

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client queries the TA service for indicator data.
type Client struct {
	baseURL string
	client  *http.Client
}

// New creates a client.
func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// IndicatorResponse describes a single value indicator.
type IndicatorResponse struct {
	Value float64 `json:"value"`
}

// MACDResponse describes the MACD payload.
type MACDResponse struct {
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
}

// SignalsResponse is the aggregated signals structure.
type SignalsResponse struct {
	Signals map[string]float64 `json:"signals"`
}

// FetchRSI returns the RSI value.
func (c *Client) FetchRSI(pair, interval string) (IndicatorResponse, error) {
	var resp IndicatorResponse
	err := c.get(fmt.Sprintf("%s/v1/indicators/rsi/%s/%s", c.baseURL, pair, interval), &resp)
	return resp, err
}

// FetchMACD returns the MACD data.
func (c *Client) FetchMACD(pair, interval string) (MACDResponse, error) {
	var resp MACDResponse
	err := c.get(fmt.Sprintf("%s/v1/indicators/macd/%s/%s", c.baseURL, pair, interval), &resp)
	return resp, err
}

// FetchSignals returns summary data.
func (c *Client) FetchSignals(pair, interval string) (SignalsResponse, error) {
	var resp SignalsResponse
	err := c.get(fmt.Sprintf("%s/v1/indicators/signals/%s/%s", c.baseURL, pair, interval), &resp)
	return resp, err
}

func (c *Client) get(url string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("ta service status %d", res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(out)
}
