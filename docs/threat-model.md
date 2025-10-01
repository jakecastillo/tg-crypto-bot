# Threat Model

## Assets
- User trade intents (confidentiality, integrity)
- Signing keys and nonces
- Portfolio and trade history
- Connectivity to DEXes, CEXes, and RPC providers

## Adversaries
- Malicious Telegram users attempting to spoof commands
- Compromised API tokens or Redis streams
- RPC provider tampering or response manipulation
- Exchange API credential theft
- Chain reorgs and sandwich attacks

## Controls
- Bot is stateless and only communicates with API over mTLS/TLS and bearer tokens
- API validates allow-listed chat IDs and enforces rate limits + safelist tokens
- Exec service operates with safelisted routers and per-trade approvals; default dry-run
- Signer maintains per-session nonces to prevent replay
- Redis streams require unique consumer group per instance to avoid duplicate execution
- Metrics and logs are structured and shipped to detect anomalies
- TA service enforces safelisted Binance/Uniswap markets, stores short-lived candles, and supports disabling auto-trade filters remotely

## Residual Risks
- Telegram account takeover of authorized users
- Latency-sensitive race conditions during liquidity launches
- RPC outages or censorship

## Mitigations in Roadmap
- Hardware security module integration for signer
- MEV protection and private orderflow defaults
- Additional anomaly detection for copy-trading feeds
