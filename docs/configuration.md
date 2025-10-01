# Configuration Reference

## Bot (`bot/config/bot.yaml`)
```yaml
telegram_token: "${TG_TRADER_BOT_TELEGRAM_TOKEN}"
api_base_url: "http://localhost:8080"
api_token: "${TG_SHARED_TOKEN}"
health_addr: ":9091"
```

## API (`api/config/api.yaml`)
```yaml
http_addr: ":8080"
metrics_addr: ":9100"
redis_url: "redis:6379"
api_token: "${TG_SHARED_TOKEN}"
rate_limit_rps: 10
ta_service_url: "http://ta-service:9100"
```

## Exec (`exec/config/default.yaml`)
```yaml
redis_url: "redis://redis:6379"
stream: "trade-intents"
group: "exec"
consumer: "exec-1"
http_addr: "0.0.0.0:8081"
metrics_addr: "0.0.0.0:9101"
dry_run: true
ta_service_url: "http://ta-service:9100"
```

## TA Service (`ta-service`)
```yaml
http_addr: "0.0.0.0:9100"
ws_addr: "0.0.0.0:9101"
metrics_addr: "0.0.0.0:9102"
postgres_url: "postgres://trader:password@postgres:5432/trader?sslmode=disable"
binance_symbols:
  - ETHUSDT
  - BTCUSDT
uniswap_pairs:
  - 0x0000000000000000000000000000000000000000
candle_limit: 1000
```
