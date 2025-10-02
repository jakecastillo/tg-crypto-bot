#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODULES=("bot" "api" "connectors/cex" "data" "risk" "ta-service")
LOCAL_PROXY="file://$(go env GOMODCACHE)/cache/download"

for module in "${MODULES[@]}"; do
    echo "==> building ${module}"
    (
        cd "${ROOT_DIR}/${module}" && \
        GOWORK=off GOSUMDB=off GOPROXY="${LOCAL_PROXY},https://proxy.golang.org,direct" go build ./...
    )
    echo "    ✓ ${module}"
    echo
done

echo "All Go modules compiled successfully."
if command -v cargo >/dev/null 2>&1; then
    echo "Rust workspace targets can be built with 'cargo build --release' once crates.io access is available."
fi
