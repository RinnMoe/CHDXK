# Repository Guidelines

## Project Structure & Module Organization

This repository contains `jcourse`, split into `backend/` and `frontend/`, with shared local dependency Compose files under `docker/`. Backend entrypoints live in `backend/cmd/api`, `backend/cmd/taskworker`, and `backend/cmd/importer`; core Go code is under `backend/internal` with `domain`, `application`, `infrastructure`, and `interface` layers. SQL migrations are in `backend/script`, config examples are in `backend/config`, and reusable backend packages are in `backend/pkg`. The React/Vite frontend lives in `frontend/src`, with pages in `src/pages`, API clients in `src/api`, hooks in `src/hooks`, routing in `src/router.tsx`, utilities in `src/lib`, UI primitives in `src/components/ui`, reusable feature components in `src/components/*`, and MSW handlers/fixtures in `src/mocks` plus the generated worker in `public/`.

## Build, Test, and Development Commands

- `cd backend && go build ./...` builds all Go packages.
- `cd backend && go test -tags test ./...` runs backend tests; repository tests expect PostgreSQL. Always include the `test` build tag for Go tests because shared repository mocks are guarded by `//go:build test`.
- `cd backend && go vet -tags test ./...` runs Go static analysis with test-tagged mocks available.
- `docker compose -f docker/docker-compose.yaml up -d postgres redis` starts local PostgreSQL and Redis from the repository root; PostgreSQL uses the custom image with `pg_jieba` support.
- `cd backend && go run cmd/api/main.go --config config/config.yaml` starts the API.
- `cd backend && go run cmd/taskworker/main.go --config config/config.yaml` starts the async worker and scheduled tasks.
- `cd backend && go run cmd/importer/main.go --target-dsn "$TARGET_DSN" --semester 2025-2026-1` imports course CSV data from `backend/data/<semester>.csv`.
- `cd frontend && pnpm dev` starts Vite locally.
- `cd frontend && pnpm build` type-checks and builds the frontend.
- `cd frontend && pnpm typecheck` runs TypeScript without bundling.
- `cd frontend && pnpm lint` runs ESLint.
- `cd frontend && pnpm test` runs Vitest.

## Coding Style & Naming Conventions

Use `gofmt` for Go and keep package names short, lower-case nouns such as `course`, `review`, or `repository`. Backend code follows DDD and CQRS: keep domain code independent of Gin, Gorm, Redis, Asynq, and other adapter details; keep interfaces separate from their implementations; add a build-tagged mock implementation whenever introducing a new domain interface used by tests. Frontend code uses TypeScript, React function components, TanStack Router/Form/Query, Tailwind CSS, shadcn-style primitives, Remix Icon, Prettier, and ESLint. Name components in PascalCase, hooks as `useThing`, and page files with kebab-case names such as `course-detail-page.tsx`. Prefer existing feature folders and UI primitives before adding new structure.

When changing database structure, fields, extensions, or indexes, add a new four-digit numbered up/down SQL migration pair under `backend/script` instead of editing an earlier migration. The initial PostgreSQL schema is `0001_schema_pgsql.up.sql`; keep numbering sequential after the current highest migration, for example `0007_xxx.up.sql` and `0007_xxx.down.sql` if `0006` is still latest.

## Testing Guidelines

Backend tests use Go's standard `testing` package and sit beside code as `*_test.go`. Prefer table-driven tests for domain policy and application logic. Repository tests use helpers in `backend/internal/infrastructure/repository`, require local PostgreSQL with `pg_jieba` and `pg_trgm`, and create/drop `jcourse_test`. Frontend changes should pass lint and type-check; add focused Vitest coverage for nontrivial hooks, API behavior, or UI logic.

## Commit & Pull Request Guidelines

Recent commits use short imperative subjects, often with a scope or type, for example `feat: add course notification controls`, `refine course detail layout`, or `style frontend point tabs`. Keep commits focused. Pull requests should include a summary, linked issue when available, test commands run, screenshots for UI changes, and notes for schema, config, or environment changes.

## Security & Configuration Tips

Do not commit real secrets or local `backend/config/config.yaml` values. Start from `backend/config/config.example.yaml` and prefer `JCOURSE_` environment overrides for local changes. Check route auth requirements in `backend/internal/interface/web/router.go` when adding endpoints.
