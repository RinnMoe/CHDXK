# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build ./...                                     # build all packages
go run cmd/api/main.go                             # run API server (uses config/config.yaml by default)
go run cmd/api/main.go --config path               # run with specific config path
go run cmd/taskworker/main.go                      # run async task worker (asynq)
go run cmd/importer/main.go --semester 2025-2026-1 # import course CSV from data/<semester>.csv
go vet ./...                                       # static analysis
go test ./...                                      # run all tests
go test ./internal/infrastructure/repository/...   # run repository tests
go test ./internal/infrastructure/repository/... -run TestFoo  # run a single test by name
docker compose up -d postgres redis                # start local Postgres (5432) and Redis (6379)
```

No Makefile or CI config exists yet. Config template: `config/config.example.yaml` — copy to `config/config.yaml` for local runs. Any config value can be overridden by env vars with `JCOURSE_` prefix (e.g. `JCOURSE_SERVER_ADDR=:9090`). SQL schema lives in `script/schema_pgsql.sql` (and `script/schema_sqlite.sql`).

## Architecture

This is a Go backend for a course review platform (jcourse) using **hexagonal/ports-and-adapters** architecture with Clean DDD layering.

### Layer structure

```
cmd/
  api/main.go                → Entrypoint: loads config, wires container, starts HTTP server
  taskworker/main.go         → Entrypoint: async task worker (asynq); graceful shutdown on SIGINT/SIGTERM
  importer/                  → CSV importer (reads data/<semester>.csv into Postgres)
  seed/                      → DB seeding entrypoint
config/                      → Config types and viper-based loading
script/                      → SQL schemas (pgsql/sqlite)
internal/
  domain/                    → Domain models, repository/query interfaces, domain logic
    announcement/            → Announcement query interface
    auth/                    → User, AuthService, password hashing, verification, password reset, login attempt (lockout), API keys, ctx helpers
    course/                  → Course model, CourseRepository, CourseQuery
    point/                   → Point balance and Transfer models with fee/policy logic
    review/                  → Review/Revision models, ReviewRepository, ReviewQuery, Guardian (authz)
      policy/                → CreatePolicy implementations (SafetyPolicy, FrequencyPolicy)
    stat/                    → Site stats model and task
    task/                    → Task interface, Enqueuer interface, global Enqueue function (pure abstraction)
    teacher/                 → TeacherQuery interface with pinyin search DTOs
  application/               → Application services (use cases) operating on domain interfaces
                               (CQRS split: *_query.go reads, *_command.go writes; DTOs in *_dto.go)
  infrastructure/
    persistence/             → Postgres (gorm) and Redis connection setup
    repository/              → Gorm-based implementations of domain repository/query interfaces
    email/                   → Verification code senders: `LogVerificationCodeSender`, `SMTPVerificationCodeSender` (gopkg.in/gomail.v2)
    task/                    → Asynq-based Enqueuer and Server implementations of the task port
  interface/
    web/                     → HTTP layer (Gin framework)
      controller/            → Request handlers
      middleware/            → Auth, CSRF, API key middleware (Redis-backed sessions)
      router.go              → Route definitions
    async/                   → Async task handler registration; builds the asynq ServeMux
  app/container.go           → DI container wiring everything together
```

### Key patterns

- **Domain interfaces are defined in domain packages** (e.g. `review.ReviewRepository`, `review.ReviewQuery`). Infrastructure implements them.
- **Two query interfaces per aggregate**: a write-oriented `Repository` (CRUD on domain models) and a read-oriented `Query` (returns `*ForQuery` structs with eager-loaded relations). Both are satisfied by the same infrastructure struct.
- **CQRS split in application layer**: `*QueryService` (reads, returns `*ForQuery` structs) and `*CommandService` (writes). Review uses `ReviewQueryService`/`ReviewCommandService` (with CreatePolicy chain on create). Course similarly has `CourseQueryService` (incl. `notification_level`) and `CourseCommandService` (incl. `SetNotificationLevel`).
- **Authorization uses Guardian objects** (`review.Guardian`) for owner/admin checks, plus **CreatePolicy chain** for review creation rules (safety, frequency).
- **Soft deletes**: `deleted_at` unix timestamp column; queries filter `deleted_at = 0`.
- **Review revisions**: on update, a `Revision` snapshot is created in a transaction alongside the review update.
- **OfferedCourse and TeacherGroup**: `Course` aggregates `OfferedCourse` (a specific semester offering) which contains a `TeacherGroup` (ordered list of instructors).
- **Entity mapping**: infrastructure layer defines `*Entity` structs (gorm models) with explicit `new*Domain()`/`new*Query()` converters — no auto-mapping.
- **Config**: `AppConfig` with `Server`, `Postgres`, `Redis`, `Session`, `Auth`, `Point`, `Asynq`, `SMTP` sections; env override via `JCOURSE_` prefix.
- **Auth**: email/password registration with verification codes (Redis-backed, configurable TTL/interval). Passwords use Django-compatible PBKDF2-SHA256 hashing. Sessions stored in Redis via `gin-contrib/sessions`. Email domains restricted via `Auth.EmailWhitelist`. Password reset flow uses verification codes. Login lockout tracks failed attempts via `login_attempt` repository.
- **API keys**: external clients authenticate via `APIKeyAuth` middleware for `/api/ext/*` routes.
- **Points & transfers**: users have a point balance; transfers between users charge a fee computed as `max(amount * RateBps / 10000, MinFee)`. Configured via `Point.TransferFeeRateBps` and `Point.TransferMinFee`.
- **Async tasks**: `domain/task` defines `Task` and `Enqueuer` interfaces and a package-level `Enqueue` function backed by a global enqueuer (set via `task.SetEnqueuer`). `infrastructure/task` provides the asynq-backed implementation. `interface/async` builds the asynq `ServeMux` and registers handlers. The API process injects the enqueuer at startup; `cmd/taskworker` runs the asynq server with graceful shutdown on SIGINT/SIGTERM.

### Routes (all under `/api`)

#### Auth (`/api/auth`)
- `POST /api/auth/register/code` — send verification code for registration
- `POST /api/auth/register` — register with email verification code
- `POST /api/auth/login` — email/password login (lockout after repeated failures)
- `POST /api/auth/logout` — logout
- `POST /api/auth/password-reset/code` — send password reset verification code
- `POST /api/auth/password-reset` — reset password with verification code

#### Course (`/api/course`)
- `GET /api/course/filters` — course filter options
- `GET /api/course/` — list courses (pagination, filtering)
- `GET /api/course/followed` — user's followed courses (auth)
- `GET /api/course/ignored` — user's ignored courses (auth)
- `GET /api/course/:courseID` — course detail (with stats, related data)
- `GET /api/course/:courseID/review` — course reviews
- `POST /api/course/:courseID/notification` — set notification level (0=normal, 1=follow, 2=ignored)

#### Teacher (`/api/teacher`)
- `GET /api/teacher/filters` — teacher filter options
- `GET /api/teacher/` — list teachers (pinyin search, filtering)
- `GET /api/teacher/:teacherID/courses` — teacher's courses

#### Review (`/api/review`)
- `GET /api/review/latest` — latest reviews (excludes ignored courses for logged-in users)
- `GET /api/review/followed` — reviews from followed courses (auth)
- `GET /api/review/:reviewID` — review detail
- `POST /api/review/` — create review
- `POST /api/review/:reviewID/vote` — vote on a review
- `PUT /api/review/:reviewID` — update review
- `DELETE /api/review/:reviewID` — delete review

#### User (`/api/user`)
- `GET /api/user/:userID/points` — user's points balance
- `GET /api/user/:userID/reviews` — user's reviews

#### Points (`/api/point`)
- `POST /api/point/transfers/preview` — preview transfer with fee calculation
- `POST /api/point/transfers` — create point transfer

#### Announcement (`/api/announcement`)
- `GET /api/announcement/` — list announcements

#### Site stats (`/api/site-stats`, admin only)
- `GET /api/site-stats/daily/yesterday` — yesterday's site stats
- `GET /api/site-stats/daily` — list daily site stats

#### External (`/api/ext`, API key auth)
- `GET /api/ext/points` — query user points by email

### Tech stack

- **Web framework**: Gin
- **ORM**: Gorm with PostgreSQL driver
- **Cache/Sessions**: go-redis/v9 (Redis-backed session store)
- **Async tasks**: hibiken/asynq (Redis-backed queue)
- **Config**: spf13/viper + spf13/pflag
- **Email**: gopkg.in/gomail.v2 (SMTP sender)

## Testing

Tests in `internal/infrastructure/repository/` require a running PostgreSQL instance. Default connection: `localhost:5432`, user/password: `postgres/postgres`, database: `jcourse_test` (created/dropped automatically per test run via `testhelper_test.go`). Start dependencies with `docker compose up -d postgres redis`.
