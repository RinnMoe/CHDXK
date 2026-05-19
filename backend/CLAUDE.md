# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build ./...                          # build all packages
go run cmd/api/main.go                  # run API server (uses config/config.yaml by default)
go run cmd/api/main.go --config path    # run with specific config path
go test ./...                           # run all tests
go test ./internal/domain/review/...    # run tests for a specific package
```

No Makefile or CI config exists yet. Config template: `config/config.example.yaml`. Any config value can be overridden by env vars with `JCOURSE_` prefix (e.g. `JCOURSE_SERVER_ADDR=:9090`).

## Architecture

This is a Go backend for a course review platform (jcourse) using **hexagonal/ports-and-adapters** architecture with Clean DDD layering.

### Layer structure

```
cmd/api/main.go              → Entrypoint: loads config, wires container, starts HTTP server
config/                       → Config types and viper-based loading
internal/
  domain/                     → Domain models, repository/query interfaces, domain logic
    auth/                     → User model, UserRepository, AuthService, context helpers
    course/                   → Course model, CourseRepository, CourseQuery
    review/                   → Review/Revision models, ReviewRepository, ReviewQuery, Guardian (authz)
      policy/                 → CreatePolicy implementations (SafetyPolicy, FrequencyPolicy)
    teacher/                  → TeacherQuery interface with pinyin search DTOs
  application/                → Application services (use cases) operating on domain interfaces
  infrastructure/
    persistence/              → Postgres (gorm) and Redis connection setup
    repository/               → Gorm-based implementations of domain repository/query interfaces
  interface/web/              → HTTP layer (Gin framework)
    controller/               → Request handlers
    middleware/                → Auth and CSRF middleware (Redis-backed sessions)
    router.go                 → Route definitions
  app/container.go            → DI container wiring everything together
```

### Key patterns

- **Domain interfaces are defined in domain packages** (e.g. `review.ReviewRepository`, `review.ReviewQuery`). Infrastructure implements them.
- **Two query interfaces per aggregate**: a write-oriented `Repository` (CRUD on domain models) and a read-oriented `Query` (returns `*ForQuery` structs with eager-loaded relations). Both are satisfied by the same infrastructure struct.
- **Authorization uses Guardian objects** (`review.Guardian`) for owner/admin checks, plus **CreatePolicy chain** for review creation rules (safety, frequency).
- **Soft deletes**: `deleted_at` unix timestamp column; queries filter `deleted_at = 0`.
- **Review revisions**: on update, a `Revision` snapshot is created in a transaction alongside the review update.
- **Entity mapping**: infrastructure layer defines `*Entity` structs (gorm models) with explicit `new*Domain()`/`new*Query()` converters — no auto-mapping.
- **Config**: `AppConfig` with `Server`, `Postgres`, `Redis`, `Session` sections; env override via `JCOURSE_` prefix.

### Routes (all under `/api`)

- `GET /api/course/` — list courses (with pagination, filtering)
- `GET /api/course/:courseID` — course detail (with stats, related data)
- `GET /api/course/:courseID/review` — course reviews
- `GET /api/teacher/` — list teachers (with pinyin search, filtering)
- `GET /api/review/latest` — latest reviews
- `GET /api/review/:reviewID` — review detail
- `POST /api/review/` — create review
- `PUT /api/review/:reviewID` — update review
- `DELETE /api/review/:reviewID` — delete review
- `GET /api/user/:userID/reviews` — user's reviews

### Tech stack

- **Web framework**: Gin
- **ORM**: Gorm with PostgreSQL driver
- **Cache/Sessions**: go-redis/v9 (Redis-backed session store)
- **Config**: spf13/viper + spf13/pflag
