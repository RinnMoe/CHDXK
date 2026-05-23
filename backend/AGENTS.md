# Repository Guidelines

## Project Structure & Module Organization

This directory is the Go backend for `jcourse`. Entrypoints live in `cmd/api` for the HTTP server, `cmd/taskworker` for the Asynq worker, and `cmd/importer` for data import jobs. Configuration loading is in `config/`, with a template at `config/config.example.yaml`. Core code follows a clean/hexagonal layout under `internal/`: `domain/` defines models, policy, and interfaces; `application/` contains use-case services and DTOs; `infrastructure/` contains Gorm, Redis, email, repository, persistence, and Asynq adapters; `interface/web` contains Gin routes, controllers, and middleware; `interface/async` registers task handlers. SQL schemas are in `script/`.

## Build, Test, and Development Commands

- `go build ./...` builds every package.
- `go vet ./...` runs Go static analysis.
- `go test ./...` runs all tests; repository tests require PostgreSQL.
- `go test ./internal/infrastructure/repository/... -run TestName` runs a focused repository test.
- `docker compose up -d postgres redis` starts local dependencies on ports `5432` and `6379`.
- `go run cmd/api/main.go --config config/config.yaml` starts the API server.
- `go run cmd/taskworker/main.go --config config/config.yaml` starts the async worker.
- `go run cmd/importer/main.go --config config/config.yaml` runs the importer.

Create `config/config.yaml` from `config/config.example.yaml` for local runs. Override config with `JCOURSE_` environment variables, for example `JCOURSE_SERVER_ADDR=:9090`.

## Coding Style & Naming Conventions

Use standard Go formatting: run `gofmt` on edited Go files and keep imports organized. Package names are short, lower-case nouns such as `course`, `review`, and `repository`. Exported types and methods use PascalCase; unexported helpers use camelCase. Keep domain interfaces and policy in `internal/domain/*`, use-case orchestration in `internal/application/*`, and adapter details in `internal/infrastructure/*`; do not leak Gorm, Gin, Redis, or Asynq types into domain packages.

## Testing Guidelines

Tests use Go's standard `testing` package and live beside the code as `*_test.go`. Prefer table-driven tests for policy, validation, and application logic. Repository tests use helpers in `internal/infrastructure/repository/testhelper_test.go` and expect PostgreSQL with user/password `postgres/postgres`; `jcourse_test` is created and dropped automatically. Run `go test ./...` before handing off backend changes when local dependencies are available.

## Commit & Pull Request Guidelines

Recent commits use lower-case, imperative subjects such as `add support for review voting` and `refactor repositories ...`. Keep subjects specific and mention the affected area when useful. Pull requests should describe behavior changes, list test commands, call out config or schema changes, and link related issues.

## Security & Configuration Tips

Do not commit real `config/config.yaml` secrets, database credentials, Redis passwords, or session secrets. Keep local-only overrides in environment variables. When adding routes, check auth middleware requirements in `internal/interface/web/router.go` and preserve owner/admin checks in the application or domain layer.
