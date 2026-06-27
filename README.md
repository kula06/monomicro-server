# monomicro-server

Small Go backend that exposes selected Monobank public currency rates as plain text for old Java ME phones.

## Endpoints

- `GET /health` returns `ok`
- `GET /rates.txt` returns:

```text
USD|buy|sell
EUR|buy|sell
PLN|buy|sell
```

Rates are fetched from `https://api.monobank.ua/bank/currency` and cached in memory for 5 minutes by default.

## Configuration

Environment variables:

| Name | Default |
| --- | --- |
| `APP_ADDR` | `:8080` |
| `CACHE_TTL_SECONDS` | `300` |
| `MONOBANK_API_URL` | `https://api.monobank.ua/bank/currency` |

## Run Locally

Requires Go 1.23 or newer.

```sh
go run ./cmd/server
```

Then open:

```sh
curl http://localhost:8080/health
curl http://localhost:8080/rates.txt
```

## Test

```sh
go test ./...
```

## Docker

Build and run:

```sh
docker compose up --build
```

The service will be available at `http://localhost:8080`.

## Notes

- Uses only the Go standard library.
- Uses `log/slog` JSON logs.
- Uses a 10 second HTTP client timeout for Monobank requests.
- Supports graceful shutdown on `SIGINT` and `SIGTERM`.
