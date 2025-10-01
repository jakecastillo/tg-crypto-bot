# Security Policy

## Supported Versions
The `main` branch receives security fixes. Release tags should be rebuilt after applying patches.

## Reporting a Vulnerability
- Email security reports to `security@tg-crypto-trader.example` using PGP (key fingerprint `ABCD 1234`)
- Provide reproduction steps, observed impact, and environment details
- Expect initial acknowledgement within 48 hours and triage within 5 business days

## Incident Guardrails
- Immediately rotate Telegram bot token, API bearer tokens, Redis credentials, and signer keystore passwords
- Switch exec service to `dry_run=true` to prevent broadcast while diagnosing issues
- Disable copy-trading feeds and revert to safelisted routers only
- Pause TA auto-trade filters (`/autotrade off`) and gate TA service API access until signal integrity is confirmed
- Use Redis intent IDs and Postgres audit tables for forensics

## Hardening Checklist
- Deploy API and exec behind mTLS-protected ingress
- Restrict Telegram bot access to allow-listed chat IDs only
- Ensure keystore directories have `0700` permissions; prefer HSM integration
- Enable Prometheus/OTel exporters and centralize logs for anomaly detection
- Run `go test ./...` and `cargo test --workspace` before promoting builds
