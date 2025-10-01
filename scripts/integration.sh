#!/usr/bin/env bash
set -euo pipefail

# Placeholder integration test harness. Spins up Anvil and simulates trade intents.
ANVIL_PID=""
cleanup() {
  if [[ -n "$ANVIL_PID" ]]; then
    kill "$ANVIL_PID" || true
  fi
}
trap cleanup EXIT

anvil --block-time 1 --fork-url ${FORK_URL:-""} &
ANVIL_PID=$!
sleep 2

cargo run -p exec &
EXEC_PID=$!
sleep 2

echo "Submitting sample intent"
curl -s -X POST "http://localhost:8080/v1/trades" \
  -H "Authorization: Bearer ${TG_SHARED_TOKEN:-local-token}" \
  -H "Content-Type: application/json" \
  -d '{"mode":"market","token":"WETH","size":0.01,"slippage_bps":50,"side":"buy","trigger":"integration","paper_trading":true}'

wait $EXEC_PID
