# Operations Runbook

## Incident Response
1. Use `/ops down` to stop execution in emergencies; bot and API are stateless.
2. Rotate API tokens and Redis credentials; invalidate Telegram bot token if compromised.
3. Inspect Prometheus metrics for anomalous latency or error spikes.
4. Review structured logs (zerolog/tracing) for failing intents. Use intent IDs from API responses.
5. Verify TA service health (`/healthz`) and disable auto-trade filters if indicator latency or signal drift is detected.
6. Replay intents in dry-run mode before re-enabling live trading.

## Deployment
- Build images via `make build` in `ops/`
- Deploy stack using docker-compose or Kubernetes manifests (todo)
- Provide `.env` with Telegram token, API token, Redis URL.

## Backups
- Postgres/Timescale retains trade history; run daily dumps via `pg_dump`.
- Redis streams are ephemeral; rely on API idempotency + Postgres for reconciliation.

## On-call Checklist
- Health endpoints: `/healthz`, `/readyz` on bot, API, exec, and TA service (plus `/ws` stream).
- Metrics: Prometheus scrape of API/exec, alerts on queue backlog and failed intents.
- Ensure signer keystore storage path has restricted permissions.
- Run TA backtests using `go run ./ta-service/cmd/backtest --file data.csv` before rolling out new filter expressions.
