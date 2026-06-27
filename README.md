# MonoMicro Server

MonoMicro Server is a small Go backend for Java ME currency-rate clients. It reads public currency data from Monobank, keeps a short in-memory cache, and exposes a tiny plain-text API that old Nokia and other MIDP phones can parse without JSON support.

The project is intentionally minimal: no database, no external Go dependencies, and no browser-facing frontend.

## Features

- `GET /rates.txt` endpoint optimized for Java ME clients.
- USD, EUR, and PLN exchange rates against UAH.
- Plain-text pipe-delimited response format.
- In-memory cache with configurable TTL.
- Monobank upstream timeout and response-size limits.
- Health check endpoint for deployment monitoring.
- JSON structured logs through Go `log/slog`.
- Graceful shutdown on `SIGINT` and `SIGTERM`.
- Docker image and Docker Compose setup for server deployment.

## Architecture

```text
cmd/server
  Application entrypoint, configuration loading, HTTP server lifecycle.

internal/config
  Environment variable parsing and defaults.

internal/monobank
  Monobank public currency API client.

internal/rates
  Rate selection, normalization, formatting, and cache.

internal/http
  HTTP routes, response headers, request logging, and error mapping.
```

Runtime flow:

```text
Java ME client -> /rates.txt -> rates service -> Monobank API
                                |
                                +-> in-memory cache
```

The server uses only the Go standard library. This keeps the binary portable and reduces deployment requirements for small VPS or container hosts.

## Build Instructions

Requirements:

- Go 1.23 or newer.

Run locally:

```sh
go run ./cmd/server
```

Build a local binary:

```sh
go build -trimpath -o monomicro-server ./cmd/server
```

Run tests:

```sh
go test ./...
```

Configuration is provided through environment variables:

| Name | Default | Description |
| --- | --- | --- |
| `APP_ADDR` | `:8080` | HTTP listen address. |
| `CACHE_TTL_SECONDS` | `300` | In-memory rates cache lifetime. |
| `MONOBANK_API_URL` | `https://api.monobank.ua/bank/currency` | Monobank currency API URL. |

## Docker Usage

Build and start the server:

```sh
docker compose up --build
```

The service will be available at:

```text
http://localhost:8080
```

Useful checks:

```sh
curl http://localhost:8080/health
curl http://localhost:8080/rates.txt
```

Build the image without Compose:

```sh
docker build -t monomicro-server .
docker run --rm -p 8080:8080 monomicro-server
```

## Java ME Build Instructions

This repository contains the server. A Java ME client can consume it with any MIDP 2.0 compatible project that performs an HTTP `GET` request to `/rates.txt`.

Recommended client requirements:

- Java ME / MIDP 2.0.
- CLDC 1.1 where available, CLDC 1.0 if the target phone requires it.
- A simple `HttpConnection` request to the server URL.
- Text parsing by lines and `|` separators.

Typical build flow with the Java ME SDK or Wireless Toolkit:

```sh
# Compile the MIDlet sources with the Java ME SDK or WTK tools.
# Package the result as JAR/JAD.
# Configure the client server URL, for example:
# http://your-server.example.com/rates.txt
```

Keep the Java ME client simple: avoid JSON parsers, reflection, floating UI dependencies, and APIs that are unavailable on MIDP phones.

## MicroEmulator Instructions

[MicroEmulator](https://microemu.org/) can run a Java ME MIDlet on a desktop machine for quick testing.

1. Start the server locally:

```sh
docker compose up --build
```

2. Build the Java ME client JAR/JAD.

3. Launch the client with MicroEmulator:

```sh
java -jar microemulator.jar path/to/client.jad
```

4. Configure the client to call:

```text
http://localhost:8080/rates.txt
```

If the emulator cannot reach `localhost`, use the host LAN IP address instead, for example `http://192.168.1.10:8080/rates.txt`.

## Real Nokia Deployment

For real Nokia Series 40, Symbian, or other Java ME devices:

1. Deploy this server to a public HTTP endpoint.
2. Make sure the phone can access that endpoint over mobile data or Wi-Fi.
3. Use a URL that the phone firmware can handle. Older devices are often more reliable with plain HTTP than HTTPS.
4. Build and sign the MIDlet if the target device requires fewer permission prompts.
5. Transfer the JAR/JAD to the phone with Bluetooth, USB, memory card, or OTA download.
6. Open the MIDlet and verify that it can fetch `/rates.txt`.

Network notes:

- Some older Nokia phones have limited TLS/certificate support.
- Keep the API response small and text-only.
- Avoid redirects where possible.
- If you use a custom domain, test it on the actual device, not only in an emulator.

## API Format

### `GET /health`

Returns:

```text
ok
```

### `GET /rates.txt`

Returns one line per currency:

```text
USD|buy|sell
EUR|buy|sell
PLN|buy|sell
```

Example:

```text
USD|40.1|41.2
EUR|43|44.25
PLN|10.123457|10.5
```

Format rules:

- Encoding: UTF-8 text.
- Content type: `text/plain; charset=utf-8`.
- Field separator: `|`.
- Line separator: `\n`.
- Order: USD, EUR, PLN.
- Values are decimal numbers trimmed to a compact Java ME friendly representation.
- If Monobank provides only a cross rate, the server uses that value for both buy and sell.

Error responses are plain text too:

| Status | Body |
| --- | --- |
| `405` | `method not allowed` |
| `502` | `upstream unavailable` or `invalid upstream rates` |
| `504` | `upstream timeout` |

## Screenshots

Screenshots are not included yet. Add emulator and real-device screenshots here before publishing a tagged release:

- MicroEmulator running the MIDlet.
- Real Nokia device showing the rates screen.
- Optional terminal screenshot with `/rates.txt` output.

## Roadmap

- Add a companion Java ME client repository or sample MIDlet.
- Add emulator screenshots and real-device photos.
- Add optional support for more currencies.
- Add release artifacts for Docker images.
- Add deployment examples for a small VPS.
- Add uptime and upstream health documentation.

## Contributing

Contributions are welcome.

Before opening a pull request:

```sh
gofmt -w ./cmd ./internal
go test ./...
```

Please keep the project compatible with the target Java ME use case:

- Preserve the plain-text API format unless a new versioned endpoint is added.
- Keep responses small.
- Do not require JSON parsing on the phone.
- Avoid adding server dependencies unless they solve a clear maintenance or runtime problem.

## License

This project is released under the MIT License. See [LICENSE](LICENSE).
