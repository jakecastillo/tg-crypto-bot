package model

import "time"

type Candle struct {
    Exchange string    `json:"exchange"`
    Pair     string    `json:"pair"`
    Interval string    `json:"interval"`
    Open     float64   `json:"open"`
    High     float64   `json:"high"`
    Low      float64   `json:"low"`
    Close    float64   `json:"close"`
    Volume   float64   `json:"volume"`
    Start    time.Time `json:"start"`
}
