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

### Prerequisites
- Docker + Docker Compose
- GNU Make
- A Telegram account (to create a bot with [@BotFather](https://t.me/BotFather))

### 1. Configure secrets interactively
Run the bootstrap helper and follow the prompts for your Telegram token, shared API secret, and allowed chat IDs:

```bash
./scripts/bootstrap.sh
```

The script copies `.env.example`, keeps the generated API tokens in sync across services (including `docker-compose`), and stores your chat allowlist.

### 2. Launch the Docker stack
Bring up Redis, Postgres, TA service, API, exec engine, and the Telegram bot in the background. Leave this terminal open; the
next commands assume you remain inside `ops/`:

```bash
cd ops
make up
```

### 3. Apply the database schema
After the containers report healthy status, load the default schema from inside `ops/` (run once):

```bash
make migrate
```

### 4. Talk to your Telegram bot
Search for your bot in Telegram, press **/start**, and try commands such as:

```
/buy ETHUSDT 0.01 0.5%
/signals ETHUSDT 1m
```

Use `make down` (inside `ops/`) to stop the stack, or rerun `./scripts/bootstrap.sh` anytime you need to update secrets.

## Testing
- Go services: `go test ./...`
- Rust crates: `cargo test --workspace`
- Risk engine unit tests cover SL/TP, cooldown, and slippage handling.
- Integration placeholder: extend `scripts/integration.sh` to spin Anvil, deploy Uniswap router, and assert buy/sell success including revoke flows.

## Building
- Use `./scripts/build-all.sh` to compile every Go module with an offline-friendly proxy fallback. The script iterates through each sub-module (`bot`, `api`, `data`, `risk`, `connectors/cex`, `ta-service`) and runs `go build` with `GOWORK=off`.
- The Rust workspace (`cargo build --release`) still requires crates.io access; once network access is restored the same script output highlights the command to run.

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
