package config

import (
	"time"

	"github.com/vrischmann/envconfig"
)

// Config holds runtime configuration for the TA service.
type Config struct {
	HTTPAddr         string        `envconfig:"default=0.0.0.0:9100"`
	WSAddr           string        `envconfig:"default=0.0.0.0:9101"`
	MetricsAddr      string        `envconfig:"default=0.0.0.0:9102"`
	PostgresURL      string        `envconfig:"required"`
	RedisURL         string        `envconfig:"optional"`
	CandleLimit      int           `envconfig:"default=1000"`
	BinanceAPIKey    string        `envconfig:"optional"`
	BinanceSecret    string        `envconfig:"optional"`
	BinanceSymbols   []string      `envconfig:"optional"`
	UniswapPairs     []string      `envconfig:"optional"`
	Interval         string        `envconfig:"default=1m"`
	BackfillLookback time.Duration `envconfig:"default=48h"`
	RustLibPath      string        `envconfig:"optional"`
}

// Load returns Config populated from environment variables.
func Load() (Config, error) {
	var cfg Config
	if err := envconfig.InitWithPrefix(&cfg, "TA_SERVICE"); err != nil {
		return Config{}, err
	}
	if cfg.CandleLimit <= 0 {
		cfg.CandleLimit = 1000
	}
	return cfg, nil
}
