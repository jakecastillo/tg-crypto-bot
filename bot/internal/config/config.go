package config

import (
    "fmt"
    "strings"

    "github.com/spf13/viper"
)

// Config holds runtime configuration for the Telegram bot.
type Config struct {
    TelegramToken   string   `mapstructure:"telegram_token"`
    APIBaseURL      string   `mapstructure:"api_base_url"`
    APIToken        string   `mapstructure:"api_token"`
    CommandPrefixes []string `mapstructure:"command_prefixes"`
    HealthAddr      string   `mapstructure:"health_addr"`
}

// Load returns a Config using viper to merge env + yaml files.
func Load() (Config, error) {
    v := viper.New()
    v.SetEnvPrefix("TG_TRADER_BOT")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    v.SetDefault("health_addr", ":9091")
    v.SetDefault("command_prefixes", []string{"/"})

    v.SetConfigName("bot")
    v.SetConfigType("yaml")
    v.AddConfigPath("./config")
    v.AddConfigPath(".")

    if err := v.ReadInConfig(); err != nil {
        // allow missing config file if env vars are used
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return Config{}, fmt.Errorf("read config: %w", err)
        }
    }

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return Config{}, fmt.Errorf("unmarshal config: %w", err)
    }

    if cfg.TelegramToken == "" {
        return Config{}, fmt.Errorf("telegram_token must be configured")
    }

    if cfg.APIBaseURL == "" {
        return Config{}, fmt.Errorf("api_base_url must be configured")
    }

    return cfg, nil
}
