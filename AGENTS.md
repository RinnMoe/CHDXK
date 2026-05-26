# Repository Guidelines

## Project Structure & Module Organization

This repository contains `jcourse`, split into `backend/` and `frontend/`. Backend entrypoints live in `backend/cmd/api`, `backend/cmd/taskworker`, and `backend/cmd/importer`; core Go code is under `backend/internal` with `domain`, `application`, `infrastructure`, and `interface` layers. SQL schemas are in `backend/script`, and config examples are in `backend/config`. The React/Vite frontend lives in `frontend/src`, with pages in `src/pages`, API clients in `src/api`, hooks in `src/hooks`, UI primitives in `src/components/ui`, reusable feature components in `src/components/*`, and MSW assets in `public/` when mocks are enabled.

## Build, Test, and Development Commands

- `cd backend && go build ./...` builds all Go packages.
- `cd backend && go test -tags test ./...` runs backend tests; repository tests expect PostgreSQL. Always include the `test` build tag for Go tests because shared repository mocks are guarded by `//go:build test`.
- `cd backend && go vet -tags test ./...` runs Go static analysis with test-tagged mocks available.
- `cd backend && docker compose up -d postgres redis` starts local backend dependencies.
- `cd backend && go run cmd/api/main.go --config config/config.yaml` starts the API.
- `cd frontend && pnpm dev` starts Vite locally.
- `cd frontend && pnpm build` type-checks and builds the frontend.
- `cd frontend && pnpm typecheck` runs TypeScript without bundling.
- `cd frontend && pnpm lint` runs ESLint.
- `cd frontend && pnpm test` runs Vitest.

## Coding Style & Naming Conventions

Use `gofmt` for Go and keep package names short, lower-case nouns such as `course`, `review`, or `repository`. Backend code follows DDD and CQRS: keep domain code independent of Gin, Gorm, Redis, infrastructure, and other adapter details; keep interfaces separate from their implementations; add a mock implementation whenever introducing a new interface. Frontend code uses TypeScript, React function components, path aliases such as `@/components/ui/button`, Prettier, Tailwind CSS, and ESLint. Name components in PascalCase, hooks as `useThing`, and page files with kebab-case names such as `course-detail-page.tsx`. Prefer existing feature folders and UI primitives before adding new structure.

When changing database structure, fields, or indexes, add a new numbered SQL file under `backend/script` instead of editing an earlier migration. The initial PostgreSQL schema is `01-schema-pgsql.sql`; subsequent changes should be named `02-xxx.sql`, `03-xxx.sql`, and so on.

## Testing Guidelines

Backend tests use Go's standard `testing` package and sit beside code as `*_test.go`. Prefer table-driven tests for domain policy and application logic. Repository tests use helpers in `backend/internal/infrastructure/repository` and require local PostgreSQL. Frontend changes should pass lint and type-check; add focused Vitest coverage for nontrivial hooks, API behavior, or UI logic.

## Commit & Pull Request Guidelines

Recent commits use short imperative subjects, often with a scope or type, for example `feat: add course notification controls`, `refine course detail layout`, or `style frontend point tabs`. Keep commits focused. Pull requests should include a summary, linked issue when available, test commands run, screenshots for UI changes, and notes for schema, config, or environment changes.

## Security & Configuration Tips

Do not commit real secrets or local `backend/config/config.yaml` values. Start from `backend/config/config.example.yaml` and prefer `JCOURSE_` environment overrides for local changes. Check route auth requirements in `backend/internal/interface/web/router.go` when adding endpoints.
