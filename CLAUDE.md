# CLAUDE.md — micro-rest-events

## Project Overview

A lightweight REST microservice for managing user notifications/events. Provides both a **web UI dashboard** (htmx, session-based auth) and a **JSON API** (JWT auth). Supports SQLite and PostgreSQL.

---

## Architecture

```
cmd/server/           — entry point (Options, main)
internal/
  logger/             — custom slog.Handler (plain + colored debug)
  repository/         — data layer: StoreProvider, Event CRUD
    mocks/            — testify mock for StoreProviderInterface
  web/                — HTTP: server, routes, handlers, templates, static
    templates/        — HTML (dashboard.html, login.html, partials/)
    static/           — style.css, app.js
```

### Key patterns

- **Repository pattern** — `StoreProviderInterface` abstracts SQLite/PostgreSQL
- **Embedded resources** — templates and static files compiled into the binary via `//go:embed`
- **Dual auth** — session cookie for web UI, JWT for API
- **Graceful shutdown** — context cancelled on SIGTERM/Interrupt

---

## Server struct (`internal/web/server.go`)

```go
type Server struct {
    Listen        string
    Secret        string   // JWT secret (API) + HMAC key for sessions (derived)
    Version       string
    AuthLogin     string   // web UI username
    AuthPassword  string   // web UI password (empty = auth disabled)
    StoreProvider provider.StoreProviderInterface
    tmpl          *template.Template
}
```

---

## Routes

### Public (no auth)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/ping` | Health check |
| GET | `/static/*` | Embedded CSS/JS |
| GET | `/login` | Login form |
| POST | `/login` | Submit credentials |
| GET | `/logout` | Clear session |
| GET | `/robots.txt` | Disallow all |

### Web UI (session cookie required)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Dashboard |
| GET | `/web/events` | Events table (htmx partial) |
| POST | `/web/events` | Create event via form |
| POST | `/web/events/{uuid}/status` | Change event status |
| POST | `/web/events/{uuid}/seen` | Mark event as seen |

### API (JWT required — header `Api-Token`, claim `user_id`)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/events` | Create single event |
| POST | `/api/v1/events/batch` | Create events for multiple users |
| POST | `/api/v1/events/{uuid}` | Update event status/message |
| GET | `/api/v1/events/users/{id}` | Get events by user (filters: `status`, `date_from`) |
| POST | `/api/v1/events/{uuid}/seen` | Mark event as seen |
| POST | `/api/v1/events/change/batch` | Bulk status update |

---

## Authentication

### Web UI — form-based session
- Login via `POST /login` with `username` / `password` form fields
- On success: sets `HttpOnly` session cookie (24-hour TTL)
- Session token format: `"username:expires:hmac-sha256"`
- HMAC key derived from `Secret`: `"session:" + Secret`
- Constant-time comparison (`hmac.Equal`) prevents timing attacks
- If `AuthPassword` is empty — auth is disabled entirely

### API — JWT Bearer
- Header: `Api-Token: <jwt>`
- Must contain `user_id` claim
- Secret configured via `--secret` / `EVENT_SECRET_KEY`

---

## Configuration (`cmd/server/main.go`)

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| `-l` / `--listen` | `LISTEN` | `:8181` | Listen address |
| `-s` / `--secret` | `EVENT_SECRET_KEY` | `123` | JWT + session HMAC secret |
| `--auth-login` | `AUTH_LOGIN` | `admin` | Web UI username |
| `--auth-password` | `AUTH_PASSWORD` | `admin` | Web UI password |
| `--conn` | `CONNECTION_DSN` | `micro_events.db` | DB connection string |
| `--dsn` | `POSTGRES_DSN` | — | PostgreSQL DSN (alternative) |
| `--dbg` | `DEBUG` | false | Enable colored debug logging |

### Connection string detection
- `postgres://...` → PostgreSQL
- `...@tcp(...` → MySQL
- `file:/...` / `*.sqlite` / `*.db` → SQLite

---

## Database

Table `events` is auto-created on startup:

```sql
id, uuid, user_id, type, status, caption, message,
is_seen BOOLEAN DEFAULT 0,
created_at, updated_at
```

---

## Running tests

```bash
# All tests (run packages sequentially to avoid OOM)
go test -p 1 -count=1 ./... -timeout 120s

# Single package
go test -p 1 -count=1 ./internal/web/... -timeout 60s

# With coverage
go test -p 1 -count=1 ./... -timeout 120s -coverprofile=coverage.out
go tool cover -func=coverage.out
```

> **Important:** always use `-p 1` — running packages in parallel causes OOM on macOS.

---

## Test structure

| File | What it tests |
|------|---------------|
| `cmd/server/main_test.go` | Full integration: start server, create/query events |
| `internal/logger/logger_test.go` | slog handler: levels, format, attrs, colors |
| `internal/repository/event_test.go` | Event CRUD against real SQLite in-memory DB |
| `internal/repository/provider_test.go` | DB type detection, table creation, interface contract |
| `internal/web/events_test.go` | API handlers with mock repository |
| `internal/web/web_test.go` | Web UI handlers with mock repository |
| `internal/web/middleware_test.go` | CORS and Bearer `Auth` middleware |
| `internal/web/server_test.go` | Server.Run, robots.txt, JWT middleware |
| `internal/web/server_routes_test.go` | Full route tests: login flow, session, redirects, API auth |

Repository tests use a real SQLite `:memory:` database — no mocks at that layer.

---

## CI/CD (`.github/workflows/go.yml`)

- **build job**: `go test -p 1 ./... -timeout=120s` + `go build ./cmd/server` + Codecov upload
- **docker job**: triggered on `workflow_dispatch` or tag `v*`; builds multi-platform image (`linux/amd64`, `linux/arm64`) and pushes to GHCR
- Tag `latest` is added on push to `master`

---

## Docker

```dockerfile
# Multi-stage: golang:1.22-alpine → scratch
# CGO disabled, vendored deps, version injected via ldflags
# Exposes port 8080
ENTRYPOINT ["/srv/rest-events"]
```

Build arg `version` is set from `GIT_BRANCH-SHA-timestamp` in CI.

---

## CORS

Controlled by `ALLOWED_ORIGINS` env var:
- Not set → `*` (allow all, for backward compatibility)
- Set → comma-separated list of allowed origins (e.g. `https://app.example.com,https://admin.example.com`)
- Preflight `OPTIONS` requests always return `204 No Content`
