package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all runtime parameters for the API service.
type Config struct {
	HTTPAddr      string        `mapstructure:"http_addr"`
	MetricsAddr   string        `mapstructure:"metrics_addr"`
	RedisURL      string        `mapstructure:"redis_url"`
	NATSServerURL string        `mapstructure:"nats_url"`
	APIToken      string        `mapstructure:"api_token"`
	AllowedChats  []int64       `mapstructure:"allowed_chats"`
	RateLimitRPS  int           `mapstructure:"rate_limit_rps"`
	RequestTTL    time.Duration `mapstructure:"request_ttl"`
	TAServiceURL  string        `mapstructure:"ta_service_url"`
}

// Load reads configuration from env and optional config file.
func Load() (Config, error) {
	v := viper.New()
	v.SetEnvPrefix("TG_TRADER_API")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigName("api")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	v.SetDefault("http_addr", ":8080")
	v.SetDefault("metrics_addr", ":9100")
	v.SetDefault("rate_limit_rps", 5)
	v.SetDefault("request_ttl", "15s")
	v.SetDefault("ta_service_url", "http://ta-service:9100")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.RedisURL == "" {
		return Config{}, fmt.Errorf("redis_url must be set")
	}

	if cfg.APIToken == "" {
		return Config{}, fmt.Errorf("api_token must be set")
	}

	if cfg.TAServiceURL == "" {
		return Config{}, fmt.Errorf("ta_service_url must be set")
	}

	if cfg.RequestTTL == 0 {
		cfg.RequestTTL = 15 * time.Second
	}

	return cfg, nil
}
