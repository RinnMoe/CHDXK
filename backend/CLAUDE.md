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
    auth/                     → User model, UserRepository, AuthService, password hashing, verification codes, context helpers
    course/                   → Course model, CourseRepository, CourseQuery
    point/                    → Point balance and Transfer models with fee/policy logic
    review/                   → Review/Revision models, ReviewRepository, ReviewQuery, Guardian (authz)
      policy/                 → CreatePolicy implementations (SafetyPolicy, FrequencyPolicy)
    teacher/                  → TeacherQuery interface with pinyin search DTOs
    task/                     → Task interface, Enqueuer interface, global Enqueue function (pure abstraction)
  application/                → Application services (use cases) operating on domain interfaces
  infrastructure/
    persistence/              → Postgres (gorm) and Redis connection setup
    repository/               → Gorm-based implementations of domain repository/query interfaces
    email/                    → Verification code sender implementations: log-based (`LogVerificationCodeSender`) and SMTP-based (`SMTPVerificationCodeSender`, via `gopkg.in/gomail.v2`)
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
- **CQRS split in application layer**: review domain is split into `ReviewQueryService` (read, returns `*ForQuery` structs) and `ReviewCommandService` (write, handles creation with CreatePolicy chain). Course domain similarly has `CourseQueryService` (reads, including `notification_level`) and `CourseCommandService` (writes, including `SetNotificationLevel`). All live in `application/`.
- **OfferedCourse and TeacherGroup**: `Course` aggregates `OfferedCourse` (a specific semester offering) which contains a `TeacherGroup` (ordered list of instructors).
- **Entity mapping**: infrastructure layer defines `*Entity` structs (gorm models) with explicit `new*Domain()`/`new*Query()` converters — no auto-mapping.
- **Config**: `AppConfig` with `Server`, `Postgres`, `Redis`, `Session`, `Auth`, `Point`, `Asynq`, `SMTP` sections; env override via `JCOURSE_` prefix.
- **Auth**: email/password registration with verification codes (Redis-backed, configurable TTL/interval). Passwords use Django-compatible PBKDF2-SHA256 hashing. Sessions stored in Redis via `gin-contrib/sessions`. Email domains restricted via `Auth.EmailWhitelist`.
- **Points & transfers**: users have a point balance; transfers between users charge a fee computed as `max(amount * RateBps / 10000, MinFee)`. Configured via `Point.TransferFeeRateBps` and `Point.TransferMinFee`.
- **Async tasks**: `domain/task` defines `Task` and `Enqueuer` interfaces and a package-level `Enqueue` function backed by a global enqueuer (set via `task.SetEnqueuer`). `infrastructure/task` provides the asynq-backed implementation. `interface/async` builds the asynq `ServeMux` and registers handlers. The API process injects the enqueuer at startup; `cmd/taskworker` runs the asynq server with graceful shutdown on SIGINT/SIGTERM.

### Routes (all under `/api`)

#### Auth (`/api/auth`)
- `POST /api/auth/register/code` — send verification code for registration
- `POST /api/auth/register` — register with email verification code
- `POST /api/auth/login` — email/password login
- `POST /api/auth/logout` — logout

#### Course (`/api/course`)
- `GET /api/course/` — list courses (with pagination, filtering)
- `GET /api/course/followed` — user's followed courses (auth required)
- `GET /api/course/ignored` — user's ignored courses (auth required)
- `GET /api/course/:courseID` — course detail (with stats, related data)
- `GET /api/course/:courseID/review` — course reviews
- `POST /api/course/:courseID/notification` — set notification level (0=normal, 1=follow, 2=ignored)

#### Teacher (`/api/teacher`)
- `GET /api/teacher/` — list teachers (with pinyin search, filtering)
- `GET /api/teacher/:teacherID/courses` — teacher's courses

#### Review (`/api/review`)
- `GET /api/review/latest` — latest reviews (excludes ignored courses for logged-in users)
- `GET /api/review/followed` — reviews from followed courses (auth required)
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

### Tech stack

- **Web framework**: Gin
- **ORM**: Gorm with PostgreSQL driver
- **Cache/Sessions**: go-redis/v9 (Redis-backed session store)
- **Async tasks**: hibiken/asynq (Redis-backed queue)
- **Config**: spf13/viper + spf13/pflag

## Testing

Tests in `internal/infrastructure/repository/` require a running PostgreSQL instance. Default connection: `localhost:5432`, user/password: `postgres/postgres`, database: `jcourse_test` (created/dropped automatically per test run via `testhelper_test.go`).
