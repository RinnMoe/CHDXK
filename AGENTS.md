# Repository Guidelines

## Project Structure & Module Organization

This repository contains `jcourse`, split into `backend/` and `frontend/`. Backend entrypoints live in `backend/cmd/api`, `backend/cmd/taskworker`, and `backend/cmd/importer`; core Go code is under `backend/internal` with `domain`, `application`, `infrastructure`, and `interface` layers. SQL schemas are in `backend/script`, and config examples are in `backend/config`. The React/Vite frontend lives in `frontend/src`, with pages in `src/pages`, API clients in `src/api`, hooks in `src/hooks`, UI primitives in `src/components/ui`, and MSW mocks in `src/mocks`.

## Build, Test, and Development Commands

- `cd backend && go build ./...` builds all Go packages.
- `cd backend && go test ./...` runs backend tests; repository tests expect PostgreSQL.
- `cd backend && go vet ./...` runs Go static analysis.
- `cd backend && docker compose up -d postgres redis` starts local backend dependencies.
- `cd backend && go run cmd/api/main.go --config config/config.yaml` starts the API.
- `cd frontend && pnpm dev` starts Vite locally.
- `cd frontend && pnpm build` type-checks and builds the frontend.
- `cd frontend && pnpm lint` runs ESLint.

## Coding Style & Naming Conventions

Use `gofmt` for Go and keep package names short, lower-case nouns such as `course`, `review`, or `repository`. Keep domain code independent of Gin, Gorm, Redis, and other adapter details. Frontend code uses TypeScript, React function components, path aliases such as `@/components/ui/button`, Prettier, Tailwind CSS, and ESLint. Name components in PascalCase, hooks as `useThing`, and page files with kebab-case names such as `course-detail-page.tsx`.

## Testing Guidelines

Backend tests use Go's standard `testing` package and sit beside code as `*_test.go`. Prefer table-driven tests for domain policy and application logic. Repository tests use helpers in `backend/internal/infrastructure/repository` and require local PostgreSQL. The frontend currently relies on lint and type-check coverage; add focused component or hook tests when behavior becomes complex.

## Commit & Pull Request Guidelines

Recent commits use short imperative subjects, often with a scope or type, for example `feat: add course notification controls`, `refine course detail layout`, or `style frontend point tabs`. Keep commits focused. Pull requests should include a summary, linked issue when available, test commands run, screenshots for UI changes, and notes for schema, config, or environment changes.

## Security & Configuration Tips

Do not commit real secrets or local `backend/config/config.yaml` values. Start from `backend/config/config.example.yaml` and prefer `JCOURSE_` environment overrides for local changes. Check route auth requirements in `backend/internal/interface/web/router.go` when adding endpoints.
