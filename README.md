<p align="center">
  <img src="internal/web/static/logo.svg" width="120" alt="micro-rest-events logo"/>
</p>

# micro-rest-events

[![Build](https://github.com/jtrw/micro-rest-events/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/jtrw/micro-rest-events/actions)
[![codecov](https://codecov.io/gh/jtrw/micro-rest-events/graph/badge.svg?token=MXC3NMIN2V)](https://codecov.io/gh/jtrw/micro-rest-events)

Lightweight microservice for managing user notifications and events. Provides a **web dashboard** (htmx, session auth) and a **JSON REST API** (JWT auth). Supports SQLite and PostgreSQL.

## Features

- Create, update, and query events per user
- Mark events as seen (sets `is_seen = true` and `status = seen`)
- Bulk operations: batch create, batch status update
- Web dashboard with filters, pagination, and status management
- Dual auth: form-based session cookie for the web UI, JWT for the API

## Quick start

```bash
# SQLite (default — no extra setup needed)
go run ./cmd/server

# PostgreSQL
go run ./cmd/server --conn "postgres://user:pass@localhost/events?sslmode=disable"

# Custom listen address and credentials
go run ./cmd/server --listen :8080 --auth-login admin --auth-password secret
```

The server starts on `:8181` by default. Open `http://localhost:8181` to access the dashboard.

## Configuration

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| `-l` / `--listen` | `LISTEN` | `:8181` | Listen address |
| `-s` / `--secret` | `EVENT_SECRET_KEY` | `123` | JWT secret + HMAC key for session cookies |
| `--auth-login` | `AUTH_LOGIN` | `admin` | Web UI username |
| `--auth-password` | `AUTH_PASSWORD` | `admin` | Web UI password (empty = auth disabled) |
| `--conn` | `CONNECTION_DSN` | `micro_events.db` | DB connection string |
| `--dsn` | `POSTGRES_DSN` | — | PostgreSQL DSN (alternative to `--conn`) |
| `--dbg` | `DEBUG` | `false` | Enable colored debug logging |

**Connection string detection:**
- `postgres://...` → PostgreSQL
- `file:/...` / `*.sqlite` / `*.db` → SQLite

## API

All API endpoints are under `/api/v1` and require a JWT in the `Api-Token` header. The token must contain a `user_id` claim.

### Create event

```
POST /api/v1/events
```

```json
{
    "type": "notification",
    "user_id": "12345",
    "caption": "New message",
    "body": "Hello!",
    "status": "new"
}
```

Response: `201 Created` — `{"status": "ok", "uuid": "<uuid>"}`

### Create batch events

```
POST /api/v1/events/batch
```

```json
{
    "type": "alert",
    "users": ["user1", "user2", "user3"]
}
```

Response: `201 Created` — `{"status": "ok"}`

### Update event

```
POST /api/v1/events/{uuid}
```

```json
{
    "status": "done",
    "message": "Processed successfully"
}
```

### Get events by user

```
GET /api/v1/events/users/{id}?status=new&status=done&date_from=2024-01-01
```

Query params:
- `status` — filter by status (repeatable)
- `date_from` — filter by `updated_at >= date`

### Mark as seen

```
POST /api/v1/events/{uuid}/seen
```

Sets `is_seen = true` and `status = seen`.

### Bulk status update

```
POST /api/v1/events/change/batch
```

```json
{
    "uuids": ["uuid1", "uuid2"],
    "status": "done"
}
```

## Web dashboard

Open `http://localhost:8181` in your browser. Login with the configured credentials.

- Filter events by user ID, status, date
- Create events via form
- Change event status inline
- Mark events as seen
- Pagination (20 events per page)

CORS is controlled by the `ALLOWED_ORIGINS` environment variable:
- Not set → `*` (allow all)
- Set → comma-separated list, e.g. `https://app.example.com,https://admin.example.com`

## Docker

```bash
docker pull ghcr.io/jtrw/micro-rest-events:latest

docker run -p 8181:8181 \
  -e EVENT_SECRET_KEY=mysecret \
  -e AUTH_LOGIN=admin \
  -e AUTH_PASSWORD=changeme \
  -v $(pwd)/data:/data \
  ghcr.io/jtrw/micro-rest-events:latest --conn /data/events.db
```

Multi-platform image (`linux/amd64`, `linux/arm64`) is published to GHCR on every tagged release.

## Running tests

```bash
# All tests (sequential packages to avoid OOM on macOS)
go test -p 1 -count=1 ./... -timeout 120s

# With coverage
go test -p 1 -count=1 ./... -timeout 120s -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Repository tests use a real SQLite `:memory:` database. Web handler tests use a testify mock.
