# Architecture Overview

```mermaid
flowchart LR
    TG[Telegram Client]
    TG -->|Commands| BOT
    BOT[Go Bot Service] -->|REST/WS| API
    API[Go API Gateway] -->|Redis Streams| EXEC
    API -->|WebSocket| Clients
    EXEC[Rust Exec Orchestrator] -->|RPC| EVM[EVM DEX Connectors]
    EXEC -->|RPC| SOL[Solana Connector]
    EXEC -->|Signer RPC| SIGNER
    EXEC -->|Post-Trade| DATA[(Postgres/Timescale)]
    SIGNER[Secure Signer] -->|Signatures| EXEC
    EXEC -->|Metrics| Prometheus
```

Components are isolated by responsibility and communicate over authenticated channels. The orchestrator handles latency-sensitive operations using Rust with async runtimes while stateless Go services expose user-facing APIs.
