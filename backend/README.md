# jcourse Backend

这是 `jcourse` 的 Go 后端，负责 HTTP API、res-hub OIDC 会话、课程与教师检索、点评、积分、系统 API Key、站点统计、异步任务和课程数据导入。

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
├── config/           # 配置结构与示例配置
├── internal/
│   ├── domain/       # 领域模型、策略与接口
│   ├── application/  # 用例服务、命令、查询和 DTO
│   ├── infrastructure/ # Gorm、Redis、邮件、res-hub、任务适配器
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

空库按编号顺序应用全部 up 迁移（包含 `canteen_*` 食堂评价表）：

```bash
for f in $(ls script/*.up.sql | sort); do
  psql "host=localhost port=5432 user=postgres password=postgres dbname=jcourse sslmode=disable" \
    -v ON_ERROR_STOP=1 -f "$f"
done
```

已有数据库只需应用新增的迁移文件。个别历史迁移不可重复执行；需要重建时先执行
`docker compose -f ../docker/docker-compose.yaml down -v`。

### 4. 启动 API

```bash
go run cmd/api/main.go --config config/config.yaml
```

服务默认监听 `http://localhost:8080`，路由前缀为 `/api`。

### 构建和部署镜像

`build_docker.sh` 默认构建本地 `chdxk/api:latest` 和 `chdxk/taskworker:latest`，不会自动推送。部署时通过环境变量指定镜像仓库和标签，需要推送时显式设置 `PUSH_IMAGES=true`：

```bash
REGISTRY=registry.example.com IMAGE_NAMESPACE=chdxk IMAGE_TAG=2026-09-14 PUSH_IMAGES=true ./build_docker.sh
```

Compose 中的 `CHDXK_API_IMAGE`、`CHDXK_TASKWORKER_IMAGE` 和容器名也可由部署环境覆盖；数据库名、表名和 `JCOURSE_` 配置前缀保持兼容。

### 5. 启动异步 worker

如果需要处理异步任务或定时站点统计，另开终端执行：

```bash
go run cmd/taskworker/main.go --config config/config.yaml
```

## 常用命令

```bash
go build ./...
go test -tags test ./...
go vet -tags test ./...
```

Go 测试统一带 `-tags test`，因为 domain 包下复用的 repository mock 使用 `//go:build test` 约束。

运行单个包或单个测试示例：

```bash
go test -tags test ./internal/application -run TestCourseQuery
go test -tags test ./internal/infrastructure/repository -run TestCourseHotRepository
```

仓库层测试依赖本地 PostgreSQL，并会创建测试数据库。

## 数据导入

课程数据由管理员在“开课记录”展开“从教务同步课程”，扫码登录本科教务并选择学期进行同步。内嵌 Go HTTP 采集器读取分页与详情，后端校验完整性，在事务中替换对应学期的开课记录及教学班，保留已有评价。普通用户可在“我的课程”扫码导入全部学期的个人课程。

无需外部 SDK、Node 或浏览器。配置 `course_source.enabled` 或环境变量 `JCOURSE_COURSE_SOURCE_ENABLED` 控制入口。采集实现位于 `internal/infrastructure/bkjw`，详见 [采集说明](../docs/http-course-collector.md)。

## 配置重点

主要配置项在 `config/config.example.yaml`：

- `server.addr`：API 监听地址
- `server.cors.allowed_origins`：允许跨域访问 API 的前端 Origin 列表
- `postgres.dsn`：PostgreSQL 连接串
- `redis`：Redis 地址、账号、密码和 DB
- `session`：会话密钥、过期时间和 secure cookie 开关
- `auth.access`：访问记录批量刷新配置
- `smtp`：点评风控等运营通知的邮件发送配置
- `reshub`：必需的统一身份、会话和积分服务配置；密钥只从部署环境注入
- `stats` / `asynq`：定时统计和异步任务配置

不要提交真实的 `config/config.yaml`、SMTP 密码、OAuth secret 或 session secret。

## API 与权限

路由集中定义在 `internal/interface/web/router.go`。公开认证接口包括资料站授权、本站账密登录、注册与账号选择；注册受后台开关控制。受保护接口校验选定本站账号、会话版本、本站处罚和初始改密状态。系统接口密钥需开发维护权限，站点设置需站点配置权限，用户增删改、权限分配和积分调整分别授权。账号模型详见 [账号与权限说明](../docs/bootstrap-and-points.md)。

前端开发服务器默认从 `http://localhost:5173` 访问后端。CORS 默认允许 `*`，可通过 `server.cors.allowed_origins` 或 `JCOURSE_SERVER_CORS_ALLOWED_ORIGINS` 覆盖，并启用 cookie 凭据。

## 数据库变更

`script/0001_schema_pgsql.up.sql` 是初始 PostgreSQL schema。修改表结构、字段或索引时，请新增四位顺序编号的 up/down SQL 迁移，例如 `0004_add_course_hot.up.sql` 和 `0004_add_course_hot.down.sql`，不要直接改旧迁移。

## 编码约定

- 使用 `gofmt` 格式化 Go 文件。
- 包名使用简短小写名词，例如 `course`、`review`、`repository`。
- domain 层保持纯净，不依赖 Gin、Gorm、Redis、Asynq 等适配器。
- 新增接口时，将接口和实现分离；引入新接口时同步补 mock 实现。
- 应用逻辑优先放在 `internal/application`，适配器细节放在 `internal/infrastructure` 或 `internal/interface`。
