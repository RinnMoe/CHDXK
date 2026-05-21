# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build ./...                                     # build all packages
go run cmd/api/main.go                             # run API server (uses config/config.yaml by default)
go run cmd/api/main.go --config path               # run with specific config path
go run cmd/taskworker/main.go                      # run async task worker (asynq)
go vet ./...                                       # static analysis
go test ./...                                      # run all tests
go test ./internal/infrastructure/repository/...    # run repository tests
go test ./internal/infrastructure/repository/... -run TestFoo  # run a single test by name
```

No Makefile or CI config exists yet. Config template: `config/config.example.yaml`. Any config value can be overridden by env vars with `JCOURSE_` prefix (e.g. `JCOURSE_SERVER_ADDR=:9090`).

## Architecture

This is a Go backend for a course review platform (jcourse) using **hexagonal/ports-and-adapters** architecture with Clean DDD layering.

### Layer structure

```
cmd/api/main.go              → Entrypoint: loads config, wires container, starts HTTP server
cmd/taskworker/main.go       → Entrypoint: async task worker (asynq); starts task server with graceful shutdown
config/                       → Config types and viper-based loading
internal/
  domain/                     → Domain models, repository/query interfaces, domain logic
    auth/                     → User model, UserRepository, AuthService, context helpers
    course/                   → Course model, CourseRepository, CourseQuery
    review/                   → Review/Revision models, ReviewRepository, ReviewQuery, Guardian (authz)
      policy/                 → CreatePolicy implementations (SafetyPolicy, FrequencyPolicy)
    teacher/                  → TeacherQuery interface with pinyin search DTOs
    task/                     → Task interface, Enqueuer interface, global Enqueue function (pure abstraction)
  application/                → Application services (use cases) operating on domain interfaces
  infrastructure/
    persistence/              → Postgres (gorm) and Redis connection setup
    repository/               → Gorm-based implementations of domain repository/query interfaces
    task/                     → Asynq-based Enqueuer and Server implementations of the task port
  interface/
    web/                      → HTTP layer (Gin framework)
      controller/             → Request handlers
      middleware/             → Auth and CSRF middleware (Redis-backed sessions)
      router.go               → Route definitions
    async/                    → Async task handler registration; builds the asynq ServeMux
  app/container.go            → DI container wiring everything together
```

### Key patterns

- **Domain interfaces are defined in domain packages** (e.g. `review.ReviewRepository`, `review.ReviewQuery`). Infrastructure implements them.
- **Two query interfaces per aggregate**: a write-oriented `Repository` (CRUD on domain models) and a read-oriented `Query` (returns `*ForQuery` structs with eager-loaded relations). Both are satisfied by the same infrastructure struct.
- **Authorization uses Guardian objects** (`review.Guardian`) for owner/admin checks, plus **CreatePolicy chain** for review creation rules (safety, frequency).
- **Soft deletes**: `deleted_at` unix timestamp column; queries filter `deleted_at = 0`.
- **Review revisions**: on update, a `Revision` snapshot is created in a transaction alongside the review update.
- **CQRS split in application layer**: review domain is split into `ReviewQueryService` (read, returns `*ForQuery` structs) and `ReviewCommandService` (write, handles creation with CreatePolicy chain). Both live in `application/`.
- **OfferedCourse and TeacherGroup**: `Course` aggregates `OfferedCourse` (a specific semester offering) which contains a `TeacherGroup` (ordered list of instructors).
- **Entity mapping**: infrastructure layer defines `*Entity` structs (gorm models) with explicit `new*Domain()`/`new*Query()` converters — no auto-mapping.
- **Config**: `AppConfig` with `Server`, `Postgres`, `Redis`, `Session`, `Asynq` sections; env override via `JCOURSE_` prefix.
- **Async tasks**: `domain/task` defines `Task` and `Enqueuer` interfaces and a package-level `Enqueue` function backed by a global enqueuer (set via `task.SetEnqueuer`). `infrastructure/task` provides the asynq-backed implementation. `interface/async` builds the asynq `ServeMux` and registers handlers. The API process injects the enqueuer at startup; `cmd/taskworker` runs the asynq server with graceful shutdown on SIGINT/SIGTERM.

### Routes (all under `/api`)

- `GET /api/course/` — list courses (with pagination, filtering)
- `GET /api/course/:courseID` — course detail (with stats, related data)
- `GET /api/course/:courseID/review` — course reviews
- `GET /api/teacher/` — list teachers (with pinyin search, filtering)
- `GET /api/teacher/:teacherID/courses` — teacher's courses
- `GET /api/review/latest` — latest reviews
- `GET /api/review/:reviewID` — review detail
- `POST /api/review/` — create review
- `POST /api/review/:reviewID/vote` — vote on a review
- `PUT /api/review/:reviewID` — update review
- `DELETE /api/review/:reviewID` — delete review
- `GET /api/user/:userID/reviews` — user's reviews

### Tech stack

- **Web framework**: Gin
- **ORM**: Gorm with PostgreSQL driver
- **Cache/Sessions**: go-redis/v9 (Redis-backed session store)
- **Async tasks**: hibiken/asynq (Redis-backed queue)
- **Config**: spf13/viper + spf13/pflag

## Testing

Tests in `internal/infrastructure/repository/` require a running PostgreSQL instance. Default connection: `localhost:5432`, user/password: `postgres/postgres`, database: `jcourse_test` (created/dropped automatically per test run via `testhelper_test.go`).
