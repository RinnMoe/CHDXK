# Repository Guidelines

## Project Structure & Module Organization

This directory is the Go backend for `jcourse`. Entrypoints live in `cmd/api` for the HTTP server, `cmd/taskworker` for the Asynq worker and schedulers, and `cmd/importer` for course CSV import jobs. Configuration loading is in `config/`, with a template at `config/config.example.yaml`. Core code follows a clean/hexagonal layout under `internal/`: `domain/` defines models, policy, repository/query interfaces, and test mocks; `application/` contains use-case services, commands, queries, DTOs, and templates; `infrastructure/` contains Gorm repositories, Redis persistence, SMTP, JAccount, moderation, and Asynq adapters; `interface/web` contains Gin routes, controllers, and middleware; `interface/async` registers task handlers. Shared backend helpers live in `pkg/`. SQL migrations are in `script/`, starting with `0001_schema_pgsql.up.sql`.

## Build, Test, and Development Commands

- `go build ./...` builds every package.
- `go vet -tags test ./...` runs Go static analysis with test-tagged mocks available.
- `go test -tags test ./...` runs all tests; repository tests require PostgreSQL. Always include the `test` build tag because shared repository mocks are guarded by `//go:build test`.
- `go test -tags test ./internal/infrastructure/repository/... -run TestName` runs a focused repository test.
- `docker compose -f ../docker/docker-compose.yaml up -d postgres redis` starts local dependencies on ports `5432` and `6379`; run this from `backend/`. From the repository root, use `docker compose -f docker/docker-compose.yaml up -d postgres redis`.
- `go run cmd/api/main.go --config config/config.yaml` starts the API server.
- `go run cmd/taskworker/main.go --config config/config.yaml` starts the async worker.
- `go run cmd/importer/main.go --target-dsn "$TARGET_DSN" --semester 2025-2026-1` runs the importer for `data/<semester>.csv`.

Create `config/config.yaml` from `config/config.example.yaml` for local runs. Override config with `JCOURSE_` environment variables, for example `JCOURSE_SERVER_ADDR=:9090`.

## Coding Style & Naming Conventions

Use standard Go formatting: run `gofmt` on edited Go files and keep imports organized. Package names are short, lower-case nouns such as `course`, `review`, and `repository`. Exported types and methods use PascalCase; unexported helpers use camelCase. Backend code follows DDD and CQRS: keep domain interfaces and policy in `internal/domain/*`, use-case orchestration in `internal/application/*`, and adapter details in `internal/infrastructure/*`; do not leak Gorm, Gin, Redis, Asynq, or any infrastructure concerns into domain packages. Keep interfaces separate from their concrete implementations, and add a `//go:build test` mock in the relevant domain `testutil.go` when introducing a new interface that tests will consume.

When changing database structure, fields, extensions, or indexes, add a new four-digit numbered up/down SQL migration pair under `script/` instead of editing an earlier migration. Keep numbering sequential: `0001_schema_pgsql.up.sql` is the initial PostgreSQL schema, and the next migration should follow the current highest pair, such as `0007_xxx.up.sql` and `0007_xxx.down.sql` if `0006` remains latest.

## Testing Guidelines

Tests use Go's standard `testing` package and live beside the code as `*_test.go`. Prefer table-driven tests for policy, validation, and application logic. Repository tests use helpers in `internal/infrastructure/repository/testhelper_test.go` and expect PostgreSQL with user/password `postgres/postgres`; `jcourse_test` is created and dropped automatically. The repository tests create `pg_jieba` and `pg_trgm`, so use the custom PostgreSQL service from `docker/docker-compose.yaml` or another server with those extensions installed. Run `go test -tags test ./...` before handing off backend changes when local dependencies are available.

## Commit & Pull Request Guidelines

Recent commits use lower-case, imperative subjects such as `add support for review voting` and `refactor repositories ...`. Keep subjects specific and mention the affected area when useful. Pull requests should describe behavior changes, list test commands, call out config or schema changes, and link related issues.

## Security & Configuration Tips

Do not commit real `config/config.yaml` secrets, database credentials, Redis passwords, SMTP credentials, OAuth client secrets, or session secrets. Keep local-only overrides in environment variables. When adding routes, check auth middleware requirements in `internal/interface/web/router.go` and preserve owner/admin checks in the application or domain layer.
