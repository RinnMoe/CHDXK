# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build ./...
go vet -tags test ./...
go test -tags test ./...                                            # all tests
go test -tags test ./internal/application -run TestCourseQuery       # focused
go test -tags test ./internal/infrastructure/repository -run TestX   # repository test
go run cmd/api/main.go --config config/config.yaml                   # API (:8080)
go run cmd/taskworker/main.go --config config/config.yaml            # async worker (Asynq)
go run cmd/importer/main.go --config config/config.yaml --semester 2025-2026-1
go run cmd/migrate_v1/main.go --config config/config.yaml            # legacy v1 → v2 data migration
```

**Always pass `-tags test` to `go test` and `go vet`.** Shared repository mocks live in `internal/domain/*/testutil.go` files guarded by `//go:build test`; without the tag, code that imports them fails to compile.

Local dependencies (run from repo root): `docker compose -f docker/docker-compose.yaml up -d postgres redis`. PostgreSQL is `dujiajun/postgres-pg-jieba:17` (custom image with `pg_jieba`). Redis is `redis:7`. Initialize with `psql ... -f script/01-schema-pgsql.sql`.

Config: copy `config/config.example.yaml` → `config/config.yaml`. Any field overridable via `JCOURSE_` env vars (e.g. `JCOURSE_SERVER_ADDR=:9090`, `JCOURSE_POSTGRES_DSN=...`). Never commit `config/config.yaml`.

CI (`.github/workflows/ci.yml`, Go 1.26) runs `go vet -tags test`, `go test -tags test`, and `go build` against PostgreSQL 17 and Redis 7. Postgres credentials are injected via `JCOURSE_POSTGRES_*`.

## Architecture

Hexagonal / Clean DDD layering:

```
cmd/
  api/          HTTP server entrypoint
  taskworker/   Asynq worker; SIGINT/SIGTERM graceful shutdown
  importer/     CSV course importer (data/<semester>.csv)
  migrate_v1/   one-shot data migration from the v1 schema
config/         AppConfig types + Viper loader (JCOURSE_ env override)
script/         numbered SQL migrations (01-, 02-, 03-, ...)
internal/
  domain/       models, repository + query interfaces, policies, mocks (//go:build test)
    announcement, auth, course, point, review (incl. policy/), stat, task, teacher, account, audit
  application/  use-case services, CQRS-style command/query split, DTOs
  infrastructure/
    persistence/  gorm + redis connection setup
    repository/   gorm-based implementations of domain repository + query interfaces
    smtp/         SMTP verification-code sender (gopkg.in/gomail.v2)
    jaccount/     JAccount OAuth + course enrollment sync
    moderation/   review safety / moderation backend
    task/         Asynq Enqueuer + Server adapters
  interface/
    web/          Gin: controller/, middleware/ (auth, CSRF, API key), router.go
    async/        Asynq ServeMux + handler registration
  app/container.go  manual DI wiring
pkg/              small generic utilities
```

### Key patterns

- **Domain interfaces in domain packages** (e.g. `review.ReviewRepository`, `review.ReviewQuery`); infrastructure implements them. New interfaces always ship with a build-tagged mock in `testutil.go`.
- **Dual interfaces per aggregate**: a write-oriented `*Repository` (CRUD on domain models) and a read-oriented `*Query` (returns `*ForQuery` DTOs with eager-loaded joins). A single gorm struct in `infrastructure/repository/` satisfies both.
- **CQRS in application layer**: paired `*CommandService` / `*QueryService` per aggregate. Review creation runs a `CreatePolicy` chain (safety + frequency).
- **Authorization via Guardian objects** (e.g. `review.Guardian`) for owner / admin checks. Admin gating also happens at the route level via `middleware.RequireAdmin()` and `middleware.RequireSelfOrAdmin(param)`.
- **GORM entities are separate from domain models** — explicit `new*Domain()` / `new*Query()` converters, no auto-mapping.
- **Soft deletes** via `deleted_at` unix-timestamp column; queries filter `deleted_at = 0`.
- **Review revisions**: each update writes a `Revision` snapshot in the same transaction as the review update.
- **Course aggregate**: `Course` ↔ `OfferedCourse` (per-semester offering) ↔ `TeacherGroup` (ordered instructor list).
- **PostgreSQL features**: tsvector + GIN indexes for full-text and pinyin search (via `pg_jieba`); `int[]` arrays for categories/teacher IDs.
- **Redis**: session store (`jcourse:session:` prefix), hot-course sorted sets, verification codes, login-attempt lockout, and repository JSON caches with a 30-min default TTL. Cache keys and invalidation rules are documented in `CACHE_KEYS.md` — keep that file in sync when adding caches.
- **Async tasks**: `domain/task` defines pure `Task` / `Enqueuer` ports plus a package-level `Enqueue` function. `infrastructure/task` provides the Asynq backend; the API process calls `task.SetEnqueuer` at startup; `cmd/taskworker` runs the Asynq server.
- **Auth**: email/password registration with verification codes, Django-compatible PBKDF2-SHA256 hashes, `Auth.EmailWhitelist` domain restriction, password reset flow, login lockout via `login_attempt` repository. JAccount OAuth handles SJTU course enrollment sync (`/api/course-enrollment-sync/{start,callback}`). External clients use `APIKeyAuth` for `/api/ext/*`.
- **Points & transfers**: per-user balance with transfers between users; fee = `max(amount * RateBps / 10000, MinFee)` configured under `point.transfer_fee_rate_bps` and `point.transfer_min_fee`.
- **Audit logs**: admin actions (suspensions, role grants, system API keys, moderator remarks) are recorded via the `audit` domain.

### Routes

All routes mounted under `/api` in `internal/interface/web/router.go`. Auth tiers:

- **Public**: `/api/auth/{csrf,register,register/code,login,password-reset,password-reset/code}`, `/api/course-enrollment-sync/callback`
- **API key (`APIKeyAuth`)**: `/api/ext/*` (e.g. `GET /api/ext/point`)
- **Session auth (`RequireAuth`)**: most `/api/*` routes — course/teacher/review browsing and writes, points, API keys, user settings, announcements
- **Admin (`RequireAdmin`)**: `/api/site-stats/*`, `/api/admin/{users,api-keys,audit}/*`, review revision listing, moderator remark updates
- **Self-or-admin (`RequireSelfOrAdmin("userID")`)**: `/api/user/:userID/{point,review}`

When adding a route, set the auth requirement here — handlers should not gate access alone. Treat `router.go` as the canonical route reference; this list will go stale.

### Tech stack

- **HTTP**: Gin
- **ORM**: GORM + PostgreSQL driver
- **Cache / sessions**: go-redis/v9 + `gin-contrib/sessions` Redis store
- **Async**: hibiken/asynq (Redis-backed queues)
- **Config**: spf13/viper + spf13/pflag
- **Email**: gopkg.in/gomail.v2

## Database migrations

`script/01-schema-pgsql.sql` is the initial schema; subsequent changes are sequentially numbered (`02-add-course-rating-score.sql`, `03-add-audit-logs.sql`, ...). **Add a new numbered file rather than editing an existing migration**, even during development.

## Testing

Tests use Go's standard `testing` package and live beside code as `*_test.go`. Prefer table-driven tests for domain policy and application logic. Repository tests use helpers in `internal/infrastructure/repository/testhelper_test.go` and expect local PostgreSQL (`postgres/postgres`); the `jcourse_test` database is created and dropped automatically per test run. Always run with `-tags test`.
