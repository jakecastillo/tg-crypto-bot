# Architecture Overview

```mermaid
flowchart LR
    TG[Telegram Client]
    TG -->|Commands| BOT
    BOT[Go Bot Service] -->|REST/WS| API
    BOT -->|TA Queries| TA[TA Service]
    API[Go API Gateway] -->|Redis Streams| EXEC
    API -->|WebSocket| Clients
    API -->|Indicator Proxy| TA
    EXEC[Rust Exec Orchestrator] -->|RPC| EVM[EVM DEX Connectors]
    EXEC -->|RPC| SOL[Solana Connector]
    EXEC -->|Signer RPC| SIGNER
    EXEC -->|Post-Trade| DATA[(Postgres/Timescale)]
    EXEC -->|Filter Checks| TA
    SIGNER[Secure Signer] -->|Signatures| EXEC
    EXEC -->|Metrics| Prometheus
    TA -->|Candles| DATA
    TA -->|Signals WS| Clients
```

Components are isolated by responsibility and communicate over authenticated channels. The TA service maintains Binance and Uniswap candles with Go ingestion plus Rust indicator cores, serving both bot/API queries and exec auto-trade filters. The orchestrator handles latency-sensitive operations using Rust with async runtimes while stateless Go services expose user-facing APIs.
