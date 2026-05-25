# jcourse

`jcourse` 是一个课程点评与选课辅助系统，包含 Go 后端和 React/Vite 前端。项目提供课程与教师检索、课程点评、评分趋势、关注课程、课程通知、个人课程、积分转账、API Key、站点统计和后台用户管理等能力。

## 目录结构

```text
.
├── backend/   # Go API、异步任务、数据导入与数据库脚本
├── frontend/  # React/Vite 单页应用
└── docker/    # 本地 PostgreSQL 与 Redis 编排
```

更多细节见 [backend/README.md](backend/README.md) 和 [frontend/README.md](frontend/README.md)。

## 技术栈

- 后端：Go、Gin、Gorm、PostgreSQL、Redis、Asynq
- 前端：React、TypeScript、Vite、TanStack Query、Tailwind CSS、shadcn/ui、MSW
- 本地依赖：Docker Compose、PostgreSQL 17、Redis 7

## 快速开始

### 1. 启动依赖

```bash
docker compose -f docker/docker-compose.yaml up -d postgres redis
```

PostgreSQL 默认监听 `localhost:5432`，Redis 默认监听 `localhost:6379`。数据库默认账号为 `postgres/postgres`，数据库名为 `jcourse`。

### 2. 初始化数据库

```bash
psql "host=localhost port=5432 user=postgres password=postgres dbname=jcourse sslmode=disable" -f backend/script/01-schema-pgsql.sql
```

### 3. 配置并启动后端

```bash
cp backend/config/config.example.yaml backend/config/config.yaml
cd backend
go run cmd/api/main.go --config config/config.yaml
```

API 默认监听 `http://localhost:8080`，接口前缀为 `/api`。需要异步任务或定时统计时，另开终端启动 worker：

```bash
cd backend
go run cmd/taskworker/main.go --config config/config.yaml
```

### 4. 启动前端

```bash
pnpm --dir frontend install
pnpm --dir frontend dev
```

前端默认监听 `http://localhost:5173`。Vite 开发服务器会把 `/api` 代理到 `http://localhost:8080`。

## 常用命令

### 后端

```bash
cd backend
go build ./...
go test -tags test ./...
go vet ./...
```

后端 Go 测试统一带 `-tags test`，用于启用测试环境的共享 repository mock。

### 前端

```bash
cd frontend
pnpm lint
pnpm typecheck
pnpm test
pnpm build
```

## 配置说明

后端配置模板在 `backend/config/config.example.yaml`。本地开发时复制为 `backend/config/config.yaml`，不要提交真实密钥或本地配置。

配置可通过 `JCOURSE_` 前缀的环境变量覆盖，例如：

```bash
JCOURSE_SERVER_ADDR=:9090 go run cmd/api/main.go --config config/config.yaml
```

前端默认使用相对路径 `/api` 访问后端。本地开发代理配置在 `frontend/vite.config.ts`。

## 数据导入

课程导入入口位于 `backend/cmd/importer`，默认读取 `backend/data/<semester>.csv`：

```bash
cd backend
go run cmd/importer/main.go --config config/config.yaml --semester 2025-2026-1
```

## 开发约定

- Go 代码使用 `gofmt`，后端按 domain、application、infrastructure、interface 分层。
- 前端使用 TypeScript、React 函数组件、Tailwind CSS 和 `@/` 路径别名。
- 修改数据库结构时，在 `backend/script` 下新增顺序编号 SQL 文件，不修改已有迁移。
- 提交前根据改动范围运行对应的测试、lint 和 typecheck。

## 安全提示

不要提交真实的数据库密码、Redis 密码、SMTP 凭据、OAuth 密钥或 session secret。新增接口时请检查 `backend/internal/interface/web/router.go` 中的认证与管理员权限要求。
