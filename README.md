# tg-crypto-trader

A production-ready monorepo for a latency-optimized Telegram crypto trading bot. The stack separates user interaction, API validation, and execution into hardened services with 12-factor configuration and observability baked in.

## Features
- Telegram bot (Go) with one-tap buy/sell, size presets, slippage control, TA lookups (`/rsi`, `/macd`, `/signals`), auto-trade filters, and markdown trade summaries
- API gateway (Go) providing REST + WebSocket fan-out, rate limiting, auth, and Redis/NATS job dispatch
- Execution engine (Rust) with async orchestration, Redis consumer groups, safelisted Uniswap V2/V3 hooks, TA-aware auto-trade guards, and MEV/private orderflow placeholders
- Risk engine (Go) enforcing per-token max notional, slippage caps, cooldowns, and trailing-stop scaffolding
- Connectors: EVM (Rust/ethers), Solana placeholder (Rust/Jito ready), Binance Spot testnet (Go)
- Post-trade storage in Postgres/Timescale with repositories for portfolio + PnL tracking
- Telemetry via Prometheus metrics, structured logs, and OpenTelemetry hooks
- Docker Compose for local stack including Redis, TimescaleDB, Anvil devnet, and the TA microservice
- Technical analysis service (Go + Rust FFI) maintaining Binance/Uniswap candles, computing RSI/MACD/Bollinger/ATR, and exposing REST/WebSocket feeds plus CSV-driven backtesting CLI

## Quickstart
1. Copy `.env.example` to `.env` and populate secrets
2. Start local dependencies:
   ```bash
   cd ops
   make up
   ```
3. Run migrations (requires psql access):
   ```bash
   psql postgresql://trader:password@localhost:5432/trader -f ../data/migrations/0001_init.sql
   ```
4. Launch exec in dry-run mode (optional outside Docker):
   ```bash
   cargo run -p exec
   ```
5. Connect Telegram bot to your chat and issue `/buy ETHUSDT 0.01 0.5%`
6. Explore TA commands such as `/signals ETHUSDT 1m` and enable filters `/autotrade on rsi<30 1m`
6. Inspect Redis stream `trade-intents` and exec logs for lifecycle; revoke approvals via surfaced link post-trade

## Testing
- Go services: `go test ./...`
- Rust crates: `cargo test --workspace`
- Risk engine unit tests cover SL/TP, cooldown, and slippage handling.
- Integration placeholder: extend `scripts/integration.sh` to spin Anvil, deploy Uniswap router, and assert buy/sell success including revoke flows.

## Security Posture
- Stateless bot, no private keys in Telegram tier
- API verifies bearer token + allow-listed chat IDs
- Exec enforces safelisted tokens/routers, per-trade approvals, and optional dry-run signing
- Signer abstracts keystore (memory/file/HSM) and maintains per-session nonces
- Risk engine blocks trades exceeding global/token limits or slippage budget

See [`docs/threat-model.md`](docs/threat-model.md) for detailed threat analysis and [`SECURITY.md`](SECURITY.md) for reporting guidelines.

## Repository Layout
```
bot/          # Telegram interface (Go)
api/          # REST/WebSocket API (Go)
exec/         # Execution orchestrator (Rust)
ta-service/   # Real-time TA service (Go + Rust indicators + backtesting CLI)
connectors/   # DEX/CEX connectors (Rust + Go)
signer/       # Signing abstractions (Rust)
risk/         # Risk policy engine (Go)
data/         # Persistence and migrations (Go + SQL)
ops/          # Dockerfiles, compose, Makefile, CI config
```

## Roadmap Highlights
- Full Uniswap V3 pathing + mev-boost/Flashbots bundle submission
- Solana Jito integration with bundle signing and priority fees
- Copy-trading feeds with allow-listed channels and risk budget overlays
- Advanced monitoring: queue depth SLOs, structured alerting, anomaly detection

## License
MIT
