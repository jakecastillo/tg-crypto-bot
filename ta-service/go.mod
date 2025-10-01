module github.com/example/tg-crypto-trader/ta-service

go 1.20

require (
    github.com/adshao/go-binance/v2 v2.5.0
    github.com/go-chi/chi/v5 v5.0.10
    github.com/go-chi/cors v1.2.1
    github.com/joho/godotenv v1.5.1
    github.com/markcheno/go-talib v0.0.0-20190307022043-ec69c2c24f25
    github.com/rs/zerolog v1.30.0
    github.com/thoas/go-funk v0.9.3
    github.com/vrischmann/envconfig v1.3.0
    github.com/jackc/pgx/v5 v5.5.4
    github.com/gorilla/websocket v1.5.1
)

replace github.com/example/tg-crypto-trader/ta-service/rustlib => ./rustlib
