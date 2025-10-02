module github.com/example/tg-crypto-trader/ta-service

go 1.20

require (
	github.com/adshao/go-binance/v2 v2.5.0
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.1
	github.com/gorilla/websocket v1.5.1
	github.com/jackc/pgx/v5 v5.5.4
	github.com/rs/zerolog v1.31.0
	github.com/vrischmann/envconfig v1.3.0
)

require (
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace (
	github.com/json-iterator/go => ../third_party/github.com/json-iterator/go
	github.com/kr/pretty v0.3.0 => github.com/kr/pretty v0.3.1
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 => github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
)
