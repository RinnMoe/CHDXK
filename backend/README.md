# jcourse Backend

这是 `jcourse` 的 Go 后端，负责 HTTP API、认证会话、课程与教师检索、点评、积分、API Key、站点统计、异步任务和课程数据导入。

## 技术栈

- Go 1.26
- Gin HTTP 路由与中间件
- Gorm + PostgreSQL
- Redis session、限流和 Asynq 队列
- Viper 配置加载

## 目录结构

```text
backend/
├── cmd/
│   ├── api/          # HTTP API 服务
│   ├── taskworker/   # Asynq worker 与定时任务
│   ├── importer/     # 课程 CSV 导入
│   └── migrate_v1/   # 旧版本数据迁移工具
├── config/           # 配置结构与示例配置
├── data/             # 本地导入数据目录
├── internal/
│   ├── domain/       # 领域模型、策略与接口
│   ├── application/  # 用例服务、命令、查询和 DTO
│   ├── infrastructure/ # Gorm、Redis、邮件、JAccount、任务适配器
│   └── interface/    # Web controller/middleware 与 async handler
├── pkg/              # 通用包
└── script/           # PostgreSQL schema 与迁移 SQL
```

## 本地启动

### 1. 启动 PostgreSQL 和 Redis

从仓库根目录执行：

```bash
docker compose -f docker/docker-compose.yaml up -d postgres redis
```

默认连接信息：

- PostgreSQL：`localhost:5432`，用户 `postgres`，密码 `postgres`，数据库 `jcourse`
- Redis：`localhost:6379`

### 2. 准备配置

```bash
cd backend
cp config/config.example.yaml config/config.yaml
```

按需修改 `config/config.yaml`。任意配置项都可以用 `JCOURSE_` 前缀环境变量覆盖，例如：

```bash
JCOURSE_SERVER_ADDR=:9090 go run cmd/api/main.go --config config/config.yaml
```

### 3. 初始化数据库 schema

```bash
psql "host=localhost port=5432 user=postgres password=postgres dbname=jcourse sslmode=disable" -f script/0001_schema_pgsql.up.sql
```

### 4. 启动 API

```bash
go run cmd/api/main.go --config config/config.yaml
```

服务默认监听 `http://localhost:8080`，路由前缀为 `/api`。

### 5. 启动异步 worker

如果需要处理异步任务或定时站点统计，另开终端执行：

```bash
go run cmd/taskworker/main.go --config config/config.yaml
```

## 常用命令

```bash
go build ./...
go test -tags test ./...
go vet ./...
```

Go 测试统一带 `-tags test`，因为 domain 包下复用的 repository mock 使用 `//go:build test` 约束。

运行单个包或单个测试示例：

```bash
go test -tags test ./internal/application -run TestCourseQuery
go test -tags test ./internal/infrastructure/repository -run TestCourseHotRepository
```

仓库层测试依赖本地 PostgreSQL，并会创建测试数据库。

## 数据导入

课程导入工具默认读取 `data/<semester>.csv`：

```bash
go run cmd/importer/main.go --config config/config.yaml --semester 2025-2026-1
```

对应文件路径为 `backend/data/2025-2026-1.csv`。

## 配置重点

主要配置项在 `config/config.example.yaml`：

- `server.addr`：API 监听地址
- `server.cors.allowed_origins`：允许跨域访问 API 的前端 Origin 列表
- `postgres.dsn`：PostgreSQL 连接串
- `redis`：Redis 地址、账号、密码和 DB
- `session`：会话密钥、过期时间和 secure cookie 开关
- `auth`：注册、重置密码、登录限制和密码哈希参数
- `smtp`：邮件验证码发送配置
- `jaccount`：JAccount OAuth 与课程同步配置
- `stats` / `asynq`：定时统计和异步任务配置

不要提交真实的 `config/config.yaml`、SMTP 密码、OAuth secret 或 session secret。

## API 与权限

路由集中定义在 `internal/interface/web/router.go`。公开接口包括注册、登录、密码重置、CSRF token 和课程同步 callback；大多数 `/api` 接口需要登录；后台用户、站点统计和部分点评管理接口需要管理员权限。

前端开发服务器默认从 `http://localhost:5173` 访问后端。CORS 默认允许 `http://localhost:5173` 和 `http://127.0.0.1:5173`，可通过 `server.cors.allowed_origins` 或 `JCOURSE_SERVER_CORS_ALLOWED_ORIGINS` 覆盖，并启用 cookie 凭据。

## 数据库变更

`script/0001_schema_pgsql.up.sql` 是初始 PostgreSQL schema。修改表结构、字段或索引时，请新增四位顺序编号的 up/down SQL 迁移，例如 `0004_add_course_hot.up.sql` 和 `0004_add_course_hot.down.sql`，不要直接改旧迁移。

## 编码约定

- 使用 `gofmt` 格式化 Go 文件。
- 包名使用简短小写名词，例如 `course`、`review`、`repository`。
- domain 层保持纯净，不依赖 Gin、Gorm、Redis、Asynq 等适配器。
- 新增接口时，将接口和实现分离；引入新接口时同步补 mock 实现。
- 应用逻辑优先放在 `internal/application`，适配器细节放在 `internal/infrastructure` 或 `internal/interface`。
