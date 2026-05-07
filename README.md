# Lumi 🎬

A production-ready JSON REST API for managing movie data — built in Go, backed by PostgreSQL, with zero frameworks.

Lumi handles everything you'd expect from a real backend service: user registration and authentication via stateless tokens, role-based permissions, full CRUD on movies with filtering and pagination, rate limiting, graceful shutdown, and structured error handling. Every piece is wired together by hand — no ORM, no web framework, just Go's standard library and a small set of focused dependencies.

---

## Tech Stack

| Layer | Choice |
|---|---|
| Language | Go 1.25 |
| Router | `julienschmidt/httprouter` |
| Database | PostgreSQL (`lib/pq`) |
| Migrations | `golang-migrate/migrate` |
| Mailing | `go-mail/mail` |
| Dev Tooling | `air` (live reload) |

---

## Architecture

```
cmd/
└── api/                   # Entry point — server, routing, middleware, handlers
    ├── main.go            # Config, DB pool, server bootstrap
    ├── routes.go          # All route declarations
    ├── middleware.go      # Rate limiting, auth, CORS, metrics
    ├── errors.go          # Centralised error helpers
    ├── helpers.go         # Request reading, response writing, query parsing
    ├── movies.go          # Movie handlers
    ├── users.go           # User registration handlers
    ├── tokens.go          # Authentication token handlers
    └── healthcheck.go

internal/
├── data/                  # Models, database queries, type definitions
│   ├── movies.go          # Movie model + CRUD queries
│   ├── users.go           # User model + DB queries
│   ├── tokens.go          # Token model + DB queries
│   ├── permissions.go     # Permission model + DB queries
│   ├── filters.go         # Pagination and sorting helpers
│   ├── runtime.go         # Custom JSON type for movie runtime
│   └── models.go          # Aggregate model struct passed to handlers
└── validator/             # Reusable input validation

migrations/                # Versioned up/down SQL migration files
```

The `cmd/api` layer handles only HTTP concerns — reading requests, calling into `internal/data`, and writing responses. All database logic lives in `internal/data`. Handlers never touch SQL directly. This boundary is enforced by Go's package system, not just convention.

---

## Features

**Movies CRUD** — Create, read, update, and delete movies. Each movie carries a title, year, runtime, and a set of genres. Partial updates are supported via `PATCH` — only the fields you send get changed.

**Filtering, Sorting, and Pagination** — The `GET /v1/movies` endpoint accepts query parameters for filtering by title and genre, sorting by any field, and paginating results. Page metadata (current page, page size, total records, total pages) is included in every list response. Total record count is fetched alongside the data in a single query using `count(*) OVER()`, avoiding a separate `COUNT` round trip.

**User Registration and Activation** — New users register with a name, email, and password. Passwords are hashed with bcrypt before storage. An activation token is sent to the user's email; the account is inactive until the token is submitted to the activation endpoint.

**Token-Based Authentication** — Authenticated users exchange credentials for a stateless bearer token stored in the database. Tokens are scoped (authentication, activation, password-reset) and carry an expiry. Every protected endpoint reads the `Authorization` header and looks up the token to identify the user.

**Permission-Based Authorization** — Users are granted fine-grained permissions (e.g. `movies:read`, `movies:write`). Middleware checks the required permission for each route before the handler runs.

**Rate Limiting** — Per-client IP rate limiting using a token bucket algorithm. Each IP gets its own limiter; limiters are cleaned up periodically to prevent memory growth. Exceeds the limit → `429 Too Many Requests`.

**Structured Error Handling** — All errors flow through central helpers that produce a consistent JSON envelope. Validation errors, 404s, method-not-allowed, 500s — all return the same predictable shape so clients never have to guess.

**Graceful Shutdown** — The server listens for `SIGINT` and `SIGTERM`. On signal, it stops accepting new connections and waits for in-flight requests to complete before exiting.

**CORS** — Configurable trusted origins list. Preflight requests are handled correctly.

**Expvar Metrics** — Runtime metrics (total requests received, responses sent, processing time, goroutine count, database stats) exposed at `/debug/vars` via Go's `expvar` package.

---

## API Endpoints

| Method | Path | Auth Required | Description |
|---|---|---|---|
| `GET` | `/v1/healthcheck` | No | Server health, version, environment |
| `GET` | `/v1/movies` | `movies:read` | List movies (filter, sort, paginate) |
| `POST` | `/v1/movies` | `movies:write` | Create a movie |
| `GET` | `/v1/movies/:id` | `movies:read` | Get a single movie |
| `PATCH` | `/v1/movies/:id` | `movies:write` | Partially update a movie |
| `DELETE` | `/v1/movies/:id` | `movies:write` | Delete a movie |
| `POST` | `/v1/users` | No | Register a new user |
| `PUT` | `/v1/users/activated` | No | Activate a user account |
| `POST` | `/v1/tokens/authentication` | No | Get an authentication token |

---

## Getting Started

**Prerequisites:** Go 1.25+, PostgreSQL, `migrate` CLI.

```bash
# Clone
git clone https://github.com/EDITH5607/Lumi.git
cd Lumi

# Run migrations
migrate -path ./migrations -database $LUMI_DB_DSN up

# Start with live reload
air

# Or run directly
go run ./cmd/api \
  -db-dsn=$LUMI_DB_DSN \
  -smtp-host=sandbox.mailtrap.io \
  -smtp-username=<your-username> \
  -smtp-password=<your-password>
```

The server starts on port `4000` by default. All flags have sensible defaults and can also be set via environment variables.

### Example Request

```bash
# Create a movie
curl -X POST localhost:4000/v1/movies \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Interstellar","year":2014,"runtime":"169 mins","genres":["sci-fi","drama"]}'

# List movies with filters
curl "localhost:4000/v1/movies?genres=sci-fi&sort=-year&page=1&page_size=5" \
  -H "Authorization: Bearer <token>"
```

---

## Configuration Flags

| Flag | Default | Description |
|---|---|---|
| `-port` | `4000` | HTTP server port |
| `-env` | `development` | Environment (development/staging/production) |
| `-db-dsn` | — | PostgreSQL DSN |
| `-db-max-open-conns` | `25` | DB connection pool size |
| `-limiter-rps` | `2` | Rate limiter requests per second |
| `-limiter-burst` | `4` | Rate limiter burst size |
| `-limiter-enabled` | `true` | Toggle rate limiting |
| `-cors-trusted-origins` | — | Space-separated list of trusted origins |
| `-smtp-*` | — | SMTP server config for email sending |

---

## What's Next

- Password reset flow
- Docker + Docker Compose setup
- GitHub Actions CI pipeline
- Deployment to a Linux VPS
