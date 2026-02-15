# URL Shortener (Go + Postgres/Redis)

A simple, production-structured URL shortener with clear layers:

- `cmd/server`: entrypoint only
- `internal/config`: environment/config loading
- `internal/app`: app wiring and lifecycle management
- `internal/repository/postgres`: database repository
- `internal/repository/redis`: redis repository
- `internal/service`: business logic
- `internal/httpapi`: HTTP handlers
- `internal/middleware`: HTTP middleware

## Endpoints

- `POST /shorten` creates a short URL
- `GET /{code}` redirects to original URL
- `GET /healthz` returns health status

## Requirements

- Go 1.22+
- Docker

## 1) Add .env variables
```
DATABASE_URL='postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable'
REDIS_URL='localhost:6379'
```

## 2) Run app

```bash
make run
```

## 3) Quick test

Create short URL:

```bash
curl -s -X POST http://localhost:8080/shorten \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://golang.org"}'
```

Redirect:

```bash
curl -i http://localhost:8080/<code>
```

## Optional configuration

- `SERVER_READ_TIMEOUT` (default `10s`)
- `SERVER_WRITE_TIMEOUT` (default `10s`)
- `SERVER_IDLE_TIMEOUT` (default `60s`)
- `SHUTDOWN_TIMEOUT` (default `10s`)
- `DB_TIMEOUT` (default `5s`)
- `CODE_LENGTH` (default `6`)
- `MAX_GENERATE_RETRIES` (default `5`)
