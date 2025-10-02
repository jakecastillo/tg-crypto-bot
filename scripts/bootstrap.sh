#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_EXAMPLE="$ROOT_DIR/.env.example"
ENV_FILE="$ROOT_DIR/.env"

if [[ ! -f "$ENV_EXAMPLE" ]]; then
  echo "[bootstrap] .env.example is missing. Please pull the repo again." >&2
  exit 1
fi

ENV_CREATED=false

if [[ ! -f "$ENV_FILE" ]]; then
  cp "$ENV_EXAMPLE" "$ENV_FILE"
  ENV_CREATED=true
  echo "[bootstrap] Created .env from template."
else
  echo "[bootstrap] Using existing .env (values will be updated in-place)."
fi

prompt() {
  local var_name="$1"
  local prompt_text="$2"
  local default_value="${3-}"

  local existing_value=""
  if [[ "$ENV_CREATED" == false ]]; then
    existing_value="$(grep -E "^${var_name}=" "$ENV_FILE" | head -n1 | cut -d'=' -f2-)"
  fi
  if [[ -z "$existing_value" && -n "$default_value" ]]; then
    existing_value="$default_value"
  fi

  if [[ -n "$existing_value" ]]; then
    read -rp "$prompt_text [$existing_value]: " input_value
    if [[ -z "$input_value" ]]; then
      input_value="$existing_value"
    fi
  else
    while true; do
      read -rp "$prompt_text: " input_value
      if [[ -n "$input_value" ]]; then
        break
      fi
      echo "Please provide a value." >&2
    done
  fi

  printf '%s' "$input_value"
}

update_env_var() {
  local key="$1"
  local value="$2"

  python3 - "$ENV_FILE" "$key" "$value" <<'PY'
import re
import sys
from pathlib import Path

path = Path(sys.argv[1])
key = sys.argv[2]
value = sys.argv[3]

lines = path.read_text().splitlines() if path.exists() else []
pattern = re.compile(rf"^{re.escape(key)}=")
for idx, line in enumerate(lines):
    if pattern.match(line):
        lines[idx] = f"{key}={value}"
        break
else:
    lines.append(f"{key}={value}")
path.write_text("\n".join(lines) + "\n")
PY
}

generate_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 16
  else
    python3 - <<'PY'
import secrets
print(secrets.token_hex(16))
PY
  fi
}

cat <<'MSG'
============================
Telegram Bot Bootstrap Wizard
============================
This script will prepare environment variables and explain how to start the stack.
Press enter to accept defaults from .env.example.
MSG

TELEGRAM_TOKEN=$(prompt "TG_TRADER_BOT_TELEGRAM_TOKEN" "Enter your Telegram bot token (from @BotFather)")
update_env_var "TG_TRADER_BOT_TELEGRAM_TOKEN" "$TELEGRAM_TOKEN"

API_TOKEN_DEFAULT=$(generate_secret)
API_TOKEN=$(prompt "TG_TRADER_BOT_API_TOKEN" "Shared API token used between services" "$API_TOKEN_DEFAULT")
update_env_var "TG_TRADER_BOT_API_TOKEN" "$API_TOKEN"
update_env_var "TG_TRADER_API_API_TOKEN" "$API_TOKEN"
update_env_var "TG_SHARED_TOKEN" "$API_TOKEN"

CHAT_IDS=$(prompt "TG_TRADER_API_ALLOWED_CHATS" "Comma separated Telegram chat IDs allowed to trade (from @userinfobot)" "")
if [[ -n "$CHAT_IDS" ]]; then
  update_env_var "TG_TRADER_API_ALLOWED_CHATS" "$CHAT_IDS"
fi

API_BASE_URL=$(prompt "TG_TRADER_BOT_API_BASE_URL" "Bot API base URL" "http://localhost:8080")
update_env_var "TG_TRADER_BOT_API_BASE_URL" "$API_BASE_URL"

REDIS_URL=$(prompt "TG_TRADER_API_REDIS_URL" "Redis connection string" "redis:6379")
update_env_var "TG_TRADER_API_REDIS_URL" "$REDIS_URL"
if [[ "$REDIS_URL" == redis://* ]]; then
  EXEC_REDIS_URL="$REDIS_URL"
else
  EXEC_REDIS_URL="redis://$REDIS_URL"
fi
update_env_var "TG_TRADER_EXEC__REDIS_URL" "$EXEC_REDIS_URL"

cat <<'NEXT'

Next steps:
1. If you skipped the chat ID prompt, message @userinfobot and append the numeric "Id" to TG_TRADER_API_ALLOWED_CHATS in your .env file.
2. Start the stack with Docker Compose:
     cd ops
     make up
3. Run the database migration once (after the containers are healthy):
     cd ops
     make migrate
4. Open Telegram, search for your bot, press /start, and run /buy or /signals commands.

To tear down the stack run `make down` from the same directory.
NEXT

printf '\n[bootstrap] Environment is ready.\n'
