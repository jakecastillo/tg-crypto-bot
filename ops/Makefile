.PHONY: build up down lint fmt test

build:
docker compose -f docker-compose.yml build

up:
docker compose -f docker-compose.yml up -d

up-dev:
TG_TRADER_BOT_TELEGRAM_TOKEN=$${TG_TRADER_BOT_TELEGRAM_TOKEN} TG_SHARED_TOKEN=local-token docker compose -f docker-compose.yml up --build

down:
docker compose -f docker-compose.yml down -v

fmt:
cargo fmt --all
cd .. && gofmt -w $(shell git ls-files '*.go')

lint:
cargo clippy --all-targets -- -D warnings

 test:
cargo test --workspace
cd .. && go test ./...
