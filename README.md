# Service Boilerplate

Production-grade **Golang Microservice Boilerplate** designed for building scalable **RESTful API** and **Event-Driven** services. Built with a strict 3-Tier decoupled architecture, JWT & API-Key authentication, Anti-IDOR authorization patterns, standardized validation/error responses, single-file SQL migrations, transactional database seeding, and context-propagated database transactions (featuring a built-in Blog API module as a reference implementation).

---

## Table of Contents

1. [Setup and Running Instructions](#1-setup-and-running-instructions)
2. [Architecture Explanation](#2-architecture-explanation)
3. [Technology Choices Justification](#3-technology-choices-justification)
4. [Known Limitations and Future Improvements](#4-known-limitations-and-future-improvements)
5. [Test Coverage Report](#5-test-coverage-report)
6. [CI/CD Workflow (Development, Staging, Production)](#6-cicd-workflow-development-staging-production)

---

## 1. Setup and Running Instructions

### Prerequisites
- **Go**: `1.26+`
- **Database**: MySQL `8.0+` or PostgreSQL `15+` (or SQLite for in-memory testing)
- **Make** (optional, for shortcut commands)

### Step 1: Configure Environment Variables
Copy `.env.example` to `.env` and adjust the database credentials to match your local setup:
```bash
cp .env.example .env
```

> **Dynamic Database Driver (`DB_DRIVER`)**:
> You can set `DB_DRIVER` in `.env` to `mysql`, `postgres`, or `sqlite`. When `DB_URL` is left empty, the service automatically constructs the DSN from `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME`.

### Step 2: Run Database Migrations
Each table has its own dedicated migration file in `migrations/` containing both `-- +migrate Up` and `-- +migrate Down` sections:
- `migrations/20260101000001_create_users_table.sql`
- `migrations/20260101000002_create_posts_table.sql`
- `migrations/20260101000003_create_comments_table.sql`

```bash
# Apply all pending migrations
make migrate-up
# or: go run ./src migrate up

# Check current migration version
make migrate-version

# Rollback 1 migration step
make migrate-down

# Create a new single-file migration
make migrate-create name=add_new_table
```

### Step 3: Seed Initial Data (Optional)
Populate `users`, `posts`, and `comments` with idempotent sample data wrapped inside a database transaction:
```bash
make seed
# or: go run ./src seed
```

### Step 4: Start the HTTP Server
```bash
make run-http
# or: go run ./src http serve
```

Once the server is running:
- **Swagger UI (Interactive API Docs)**: `http://localhost:8080/swagger/index.html`
- **Liveness Probe**: `GET http://localhost:8080/healthz`
- **Readiness & DB Connection Pool Stats**: `GET http://localhost:8080/readyz`

### Sample Endpoints Reference (`/api/v1`)

As a reference implementation, the boilerplate includes a complete **Blog System** module (`Auth`, `Posts`, and `Comments`) demonstrating authentication, pagination, ownership authorization, and multi-repository database transactions:

| Method | Endpoint | Auth | Description |
| :---: | :--- | :---: | :--- |
| `POST` | `/api/v1/auth/register` | Public | Register a new user account and receive a JWT access token |
| `POST` | `/api/v1/auth/login` | Public | Authenticate with email & password to receive a JWT access token |
| `GET` | `/api/v1/auth/me` | Bearer JWT | Get the authenticated user's profile |
| `GET` | `/api/v1/posts` | Public | List posts with pagination (`page`, `per_page`), `search`, & `author_id` filter |
| `GET` | `/api/v1/posts/:id` | Public | Get a single post by ID including author details |
| `POST` | `/api/v1/posts` | Bearer JWT | Create a new blog post |
| `PUT` | `/api/v1/posts/:id` | Bearer JWT | Update a post (**Owner only** — Anti-IDOR enforced) |
| `DELETE` | `/api/v1/posts/:id` | Bearer JWT | Soft-delete a post & all its comments in **1 DB transaction** (**Owner only**) |
| `GET` | `/api/v1/posts/:postId/comments` | Public | List comments for a post with pagination |
| `POST` | `/api/v1/posts/:postId/comments` | Bearer JWT | Create a comment & increment `comments_count` in **1 DB transaction** |
| `GET` | `/api/v1/comments/:id` | Public | Get a single comment by ID including author details |
| `PUT` | `/api/v1/comments/:id` | Bearer JWT | Update a comment (**Owner only** — Anti-IDOR enforced) |
| `DELETE` | `/api/v1/comments/:id` | Bearer JWT | Soft-delete a comment & decrement `comments_count` in **1 DB transaction** (**Owner only**) |

---

## 2. Architecture Explanation

This service uses a **Strict 3-Tier Decoupled Architecture** (`Handler` -> `Service` -> `Repository`) with isolated DTO contracts per layer so that HTTP transport details never leak into business logic and GORM database models never leak directly into JSON responses.

```text
go-boilerplate-srvc/
├── migrations/                              # Per-table single-file SQL migrations (-- +migrate Up & Down)
├── src/
│   ├── main.go                              # Application entrypoint & global OpenAPI annotations
│   ├── cmd/                                 # Cobra CLI commands (http serve, event serve, migrate, seed)
│   ├── docs/                                # Generated Swagger/OpenAPI 2.0 specs
│   └── internal/
│       ├── bootstrap/app.go                 # Dependency Injection wiring & Graceful Shutdown
│       ├── config/config.go                 # Environment-driven configuration (.env)
│       ├── infra/
│       │   ├── database/
│       │   │   ├── database.go              # Dynamic GORM dialector & sql.DB Connection Pool
│       │   │   ├── tx.go                    # Context-propagated Transaction Manager (WithinTransaction)
│       │   │   ├── gorm_logger.go           # Context-aware slog GORM logger + Slow Query detector
│       │   │   ├── migration_source.go      # Custom golang-migrate driver for single-file Up/Down SQL
│       │   │   └── seeder.go                # Transactional idempotent database seeder
│       │   └── logger/logger.go             # Context-aware structured logger (auto-injects request_id & user_id)
│       ├── middleware/                      # RequestID, Recover, CORS, Logger, ErrorHandler, JWTAuth, RateLimiter
│       ├── router/http.go                   # Centralized HTTP route registration
│       ├── handler/                         # Tier 1: Delivery Layer (parse HTTP, validate, return JSON Envelope)
│       ├── service/                         # Tier 2: Business Logic Layer (ownership checks, transactions)
│       ├── repository/                      # Tier 3: Data Access Layer (GORM SQL queries)
│       ├── model/                           # Pure GORM database entities (with gorm.DeletedAt soft delete)
│       ├── dto/
│       │   ├── request/                     # HTTP request payloads & query parameters
│       │   ├── service/                     # Transport-agnostic inputs/filters for Service layer
│       │   └── response/                    # Outbound JSON response structs & model mappers
│       └── pkg/
│           ├── response/                    # Standardized JSON Envelope, AppError, & HTTP status code mapping
│           └── validator/                   # Struct validation wrapper producing per-field error details
```

### Data Flow & Layer Boundaries

```
flowchart LR
    Client["HTTP Client"] -->|JSON / Query| ReqDTO["dto/request"]
    ReqDTO -->|.ToServiceInput()| SvcDTO["dto/service"]
    SvcDTO --> Service["Service Layer (TxManager & Authz)"]
    Service <-->|model.Entity| Repo["Repository Layer (GORM)"]
    Service -->|ToResponse()| RespDTO["dto/response"]
    RespDTO -->|response.Envelope| Client
```

### Key Architectural Highlights
1. **Context-Propagated Database Transactions (`TxManager`)**:
   - `database.TxManager.WithinTransaction(ctx, fn)` starts a GORM transaction and injects it into `context.Context`.
   - Repositories call `database.FromContext(ctx, r.db)` so any repository method automatically participates in an active transaction without polluting repository method signatures with `*gorm.DB`.
   - Used in `CommentService.Create` (insert comment + increment `posts.comments_count`), `CommentService.Delete` (soft-delete comment + decrement `posts.comments_count`), `PostService.Delete` (cascade soft-delete comments + soft-delete post), and `database.Seed`.
2. **Defense-in-Depth Anti-IDOR Protection**:
   - Actor identity (`user_id`) is extracted strictly from verified JWT claims in `middleware.JWTAuth`.
   - **Service Layer**: Verifies `resource.AuthorID == input.ActorID` before any mutation and returns `403 FORBIDDEN` on mismatch.
   - **Repository Layer**: Enforces `WHERE id = ? AND author_id = ?` on all SQL `UPDATE` and `DELETE` statements to prevent Time-of-Check to Time-of-Use (TOCTOU) race conditions.
3. **Standardized Response Envelope & Centralized Error Mapping**:
   - All endpoints return `response.Envelope` (`success`, `code`, `message`, `data`, `meta`, `error`).
   - Services return domain errors (`response.NewError(response.CodeForbidden, ...)`), which `middleware.ErrorHandler` maps to deterministic HTTP status codes (`400`, `401`, `403`, `404`, `409`, `422`, `429`, `500`).

---

## 3. Technology Choices Justification

| Technology | Justification |
| :--- | :--- |
| **Go (`1.26+`)** | Strong static typing, fast compilation, low memory footprint, and native concurrency primitives suitable for high-throughput microservices. |
| **Fiber v2 (`gofiber/fiber/v2`)** | Built on top of `fasthttp` with zero-allocation routing, rich middleware ecosystem, and Express-like ergonomics while maintaining top-tier throughput. |
| **GORM + `database/sql` Pool** | Supports multi-driver flexibility (`mysql`, `postgres`, `sqlite`), clean association preloading (`Preload("Author")`), automatic soft-delete handling (`gorm.DeletedAt`), and full control over underlying `sql.DB` connection pool tuning (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`). |
| **`golang-migrate` + Custom Single-File Driver** | Industry-standard schema versioning extended with a custom `source.Driver` (`singleFileSource`) so developers can keep both `-- +migrate Up` and `-- +migrate Down` inside a single per-table `.sql` file, eliminating file-pair clutter. |
| **`go-playground/validator/v10`** | Battle-tested declarative struct validation using struct tags (`required`, `email`, `min`, `max`), customized to report exact JSON field names in `422 VALIDATION_ERROR` responses. |
| **`golang-jwt/jwt/v5`** | Stateless authentication with strict signing algorithm whitelisting (`HS256`/`HS384`/`HS512`) and expiration validation. |
| **Standard Library `log/slog`** | Zero external logging dependency, structured JSON/Text output, wrapped with a custom `contextHandler` that automatically attaches `request_id` and `user_id` to every log line across Handler, Service, and GORM SQL logs. |
| **Swaggo (`swaggo/swag`)** | Generates OpenAPI 2.0 specifications directly from Go doc comments, keeping documentation in sync with code and serving interactive Swagger UI at `/swagger/index.html`. |

---

## 4. Known Limitations and Future Improvements

### Known Limitations
1. **Stateless JWT Without Immediate Revocation**: Access tokens remain valid until `JWT_EXPIRATION` elapses because there is no server-side token blacklist or refresh token rotation table yet.
2. **In-Memory Rate Limiter**: Fiber's built-in rate limiter stores counters in local process memory, which resets on restart and is not shared across multiple horizontal replicas.
3. **Offset-Based Pagination & `LIKE` Search**: `LIMIT/OFFSET` pagination and `LOWER(title) LIKE '%...%'` queries work well for small-to-medium datasets but degrade on millions of rows without full-text indexes.

### Future Improvements
1. **Refresh Token Rotation & Redis Token Blacklist**: Add `POST /api/v1/auth/refresh` and `POST /api/v1/auth/logout` backed by Redis to support instant token revocation.
2. **Distributed Rate Limiting & Caching**: Integrate Redis for distributed rate limiting across Kubernetes pods and cache hot blog posts (`GET /api/v1/posts/:id`) with write-through invalidation.
3. **Cursor-Based (Keyset) Pagination & Full-Text Search**: Support cursor pagination (`after_id`) for infinite scrolling and MySQL `FULLTEXT` / PostgreSQL `tsvector` indexes for fast relevance search.
4. **OpenTelemetry Distributed Tracing & Prometheus Metrics**: Export `/metrics` and OTLP spans to correlate HTTP latency and database transaction duration in Grafana/Tempo.

---

## 5. Test Coverage Report

Unit and integration tests cover the end-to-end HTTP pipeline, service business logic, database transactions, anti-IDOR authorization checks, single-file migration parsing, idempotent seeding, JWT/API-key middlewares, context-aware logging, and input validation.

### Running Tests & Coverage
```bash
# Run all unit & integration tests with verbose output
make test

# Run statement coverage report across tested packages
make test-cover
```

### Coverage Summary by Package

| Package | Statement Coverage | Key Tested Scenarios |
| :--- | :---: | :--- |
| `src/internal/service` | **72.1%** | Register, duplicate email (`409`), login (`200`/`401`), profile, post CRUD, comment CRUD, atomic `comments_count` increment/decrement & cascade delete transactions, **Anti-IDOR (`403 Forbidden`)** on cross-user post/comment update & delete |
| `src/internal/handler` | **71.8%** | End-to-end HTTP tests (`Register`, `Login`, `Me`, `Create/List/Get/Update/Delete Post`, `Create/List/Get/Update/Delete Comment`, `401 Unauthorized`, `403 Forbidden` Anti-IDOR, `/healthz`, `/readyz`) |
| `src/internal/pkg/validator` | **63.2%** | Struct validation rules (`required`, `email`, `min`, `max`, `gte`) and JSON field name extraction |
| `src/internal/middleware` | **57.5%** | `JWTAuth` token parsing & expiry, `RequireRoles` RBAC, `APIKeyAuth`, and IP `RateLimiter` (`429`) |
| `src/internal/infra/database` | **51.7%** | Single-file `-- +migrate Up/Down` parser, per-table migration ordering, and idempotent transactional `Seed()` |
| `src/internal/infra/logger` | **16.7%** | Automatic `request_id` and `user_id` extraction from `context.Context` into structured `slog` records |

---

## 6. CI/CD Workflow (Development, Staging, Production)

Automated CI/CD pipelines are separated into dedicated workflow files per environment under `.github/workflows/`:
- **Development**: [`.github/workflows/development.yml`](.github/workflows/development.yml) (branch `development`)
- **Staging**: [`.github/workflows/staging.yml`](.github/workflows/staging.yml) (branch `staging`)
- **Production**: [`.github/workflows/production.yml`](.github/workflows/production.yml) (branch `main` & tags `v*.*.*`)

### Pipeline Flow

```
[Pull Request / Push per Environment Branch]
        │
        ▼
┌──────────────────────────────────────────────────────────┐
│ 1. CI: Quality Gate, Test & Build                        │
│    • go mod verify & go vet ./...                        │
│    • make test (Unit & Integration Tests)                │
│    • make test-cover (Statement Coverage Check)          │
│    • make build (Compile Binary Verification)            │
└───────────────────────────┬──────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────┐
│ 2. Package: Multi-Stage Docker Build & Push              │
│    • Builds Alpine non-root image via Dockerfile         │
│    • Pushes tagged image (dev-<sha>, staging-<sha>, v*)  │
└─────────────┬───────────────────────────┬────────────────┘
              │                           │
              ▼                           ▼
┌───────────────────────────┐ ┌────────────────────────────┐ ┌────────────────────────────┐
│ development.yml           │ │ staging.yml                │ │ production.yml             │
│ • Trigger: development    │ │ • Trigger: branch staging  │ │ • Trigger: main / tag v*   │
│ • Auto-run: migrate up    │ │ • Auto-run: migrate up     │ │ • Manual Approval Gate     │
│ • Auto-run: seed          │ │ • No seeder (QA dataset)   │ │ • Auto-run: migrate up     │
│ • Deploy & verify /readyz │ │ • Deploy & verify /readyz  │ │ • Zero-downtime rollout    │
└───────────────────────────┘ └────────────────────────────┘ └────────────────────────────┘
```

### Environment Promotion Strategy

| Environment | Workflow File | Branch / Trigger | Database Migration & Seeding | Deployment & Approval Policy |
| :--- | :--- | :--- | :--- | :--- |
| **Development** | `.github/workflows/development.yml` | Push / PR to `development` | Runs `./app migrate up` + `./app seed` automatically | Automatic deployment on every merged feature PR for rapid integration testing. |
| **Staging** | `.github/workflows/staging.yml` | Push / PR to `staging` | Runs `./app migrate up` only (seeder disabled) | Automatic deployment mirroring production config (`APP_ENV=staging`) for QA & UAT sign-off. |
| **Production** | `.github/workflows/production.yml` | Push / PR to `main` or tag `v*.*.*` | Runs `./app migrate up` only (seeder strictly disabled) | Protected by GitHub Environment **Manual Approval Gate**, followed by zero-downtime rolling update and `/healthz` + `/readyz` verification. |

