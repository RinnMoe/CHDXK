# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**jcourse** — a course review platform. Go backend (Gin + GORM + PostgreSQL), React/TypeScript frontend (Vite + shadcn + TanStack Query).

## Development Commands

### Backend
```bash
cd backend
docker compose up -d postgres redis          # start dependencies
go run cmd/api/main.go --config config/config.yaml   # API server
go run cmd/taskworker/main.go --config config/config.yaml  # async task worker
go build ./...                                # build all packages
go test ./...                                 # run all tests (requires PostgreSQL + Redis)
go test ./internal/domain/...                 # run tests for a specific package
go vet ./...                                  # static analysis
```

### Frontend
```bash
cd frontend
pnpm dev                                      # Vite dev server (proxies /api to localhost:8080)
pnpm build                                    # typecheck + production build
pnpm lint                                     # ESLint
pnpm format                                   # Prettier
pnpm typecheck                                # TypeScript check only
```

### CI (Go 1.26)
CI runs `go vet`, `go test`, and `go build` on the backend. Tests require PostgreSQL 18 and Redis 8 services. Config is injected via `JCOURSE_` env vars.

## Backend Architecture

Hexagonal/clean architecture under `backend/internal/`:

```
domain/        → entities, repository interfaces, domain services, policies
application/   → command/query services (CQRS), DTOs
infrastructure/→ GORM repositories, Redis client, Asynq tasks, SMTP email
interface/     → Gin HTTP controllers, middleware, router
app/           → service container (manual DI wiring)
```

**Key patterns:**
- Dual read/write interfaces: `*Repository` (write, domain models) vs `*Query` (read, `*ForQuery` DTOs), implemented by a single struct
- GORM entity structs with converter functions separate from domain models
- PostgreSQL tsvector + GIN indexes for full-text search, array types for categories/teacher_ids
- Redis-backed sessions, hot course sorted sets, verification codes, login rate limiting
- Asynq for async tasks (daily stats cron, etc.)
- Config via Viper with `JCOURSE_` env overrides; never commit `config/config.yaml`

**Entrypoints:** `cmd/api` (HTTP), `cmd/taskworker` (async), `cmd/importer` (CSV data import with pinyin generation)

**Routing:** All routes under `/api` in `internal/interface/web/router.go`. Auth via session middleware (Redis) + optional auth. External API routes use API key auth.

**Domain models:** Course → OfferedCourse ↔ TeacherGroup, Review → ReviewRevision/ReviewVote, User → PointRecord/PointTransfer. Review creation goes through policy checks (frequency, safety).

## Frontend Architecture

React 19 + TypeScript (strict) + Vite. Path alias: `@/` → `src/`.

```
src/api/        → HTTP client + per-domain API modules (auth, course, review, teacher, point, etc.)
src/hooks/      → TanStack Query hooks wrapping API modules (useQuery + useMutation with cache invalidation)
src/pages/      → Route page components (kebab-case filenames)
src/components/ → feature-based folders (auth/, course/, review/, teacher/, point/) + ui/ (shadcn)
src/mocks/      → MSW handlers + fixtures for offline development (auto-starts in dev)
```

**Key patterns:**
- `apiClient<T>()` wrapper with `HttpError` class, cookie-based auth
- TanStack Query for all server state; optimistic updates for voting
- shadcn/Radix UI components in `components/ui/`
- Tailwind CSS v4 with CSS variable theming (light/dark)
- Vite dev proxy: `/api` → `http://localhost:8080`

## Configuration

Copy `backend/config/config.example.yaml` to `backend/config/config.yaml` for local dev. Override any value with `JCOURSE_` prefixed env vars (e.g., `JCOURSE_POSTGRES_DSN`, `JCOURSE_REDIS_ADDR`).

## Coding Conventions

- **Go:** `gofmt`, short lowercase package names (`course`, `review`, `repository`), keep domain code free of framework imports (Gin, GORM, Redis)
- **Frontend:** TypeScript strict, React function components, PascalCase components, `useThing` hooks, kebab-case page filenames, Prettier + ESLint
- **Commits:** short imperative subjects with scope/type prefix (`feat:`, `refine`, `style`, `docs:`)