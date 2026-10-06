# 频道课程榜

频道课程榜（项目标识：CHDXK）是一个课程点评与选课辅助系统，包含 Go 后端和 React/Vite 前端。项目提供课程与教师检索、课程点评、评分趋势、关注课程、课程通知、个人课程、积分转账、系统 API Key、站点统计和后台运营能力。`jcourse` 仅作为内部 Go module、数据库和配置命名保留。

## 目录结构

```text
.
├── backend/   # Go API、异步任务、数据导入与数据库脚本
├── frontend/  # React/Vite 单页应用
└── docker/    # 本地 PostgreSQL 与 Redis 编排
```

更多细节见 [backend/README.md](backend/README.md) 和 [frontend/README.md](frontend/README.md)。

API Docker 镜像包含教务扫码登录与开课计划采集所需的运行依赖，无需宿主机 SDK 目录。在项目根目录执行 `docker compose -f backend/docker-compose.yml build api` 构建；数据库、Redis、SSO 仍按部署环境配置。教务同步操作与容器配置见 [后台管理指南](docs/admin-workflows.md#教务同步)。

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
psql "host=localhost port=5432 user=postgres password=postgres dbname=jcourse sslmode=disable" -f backend/script/0001_schema_pgsql.up.sql
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

登录契约：本站支持独立账密登录，主要入口为资料站授权登录；一个资料站账号可绑定多个本站账号。初始超级管理员、注册开关、分项权限与资料站积分见 [账号与权限说明](docs/bootstrap-and-points.md)，客户端登记见 [res-hub OIDC 接入说明](docs/reshub-integration.md)。

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
go vet -tags test ./...
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

管理员可在后台“开课记录”展开“从教务同步课程”，扫码连接本科教务并同步学期数据，在“开课记录”按课程代码、名称、主讲教师、开课单位、课程序号和教学班数量组合筛选，按列排序、分页维护课程与教学班。已移除手动 ZIP 上传；教务同步仍校验采集完整性并替换对应学期开课数据，已有课程评价保留。教师信息缺失的教学班也可导入。

```powershell
Compress-Archive -Path "tools/bkjw-course-source/data/bkjw/2026-2027-1/semester.json", "tools/bkjw-course-source/data/bkjw/2026-2027-1/report.json" -DestinationPath "$env:TEMP/2026-2027-1.zip"
```

## 长大教务课程数据源

`tools/bkjw-course-source` 提供可复用的 TypeScript + Playwright CLI，从长大教务系统的全校开课查询页按学期保存列表与详情 JSON。首次使用运行 `pnpm --dir tools/bkjw-course-source start login`，按终端提示输入账号密码；需要验证码时查看浏览器中的图片并手动输入。登录验证成功后再使用 `semesters`、`fetch --semester` 或 `fetch --all`；详见 [工具说明](tools/bkjw-course-source/README.md)。

## 开发约定

- Go 代码使用 `gofmt`，后端按 domain、application、infrastructure、interface 分层。
- 前端使用 TypeScript、React 函数组件、Tailwind CSS 和 `@/` 路径别名。
- 修改数据库结构时，在 `backend/script` 下新增四位顺序编号的 up/down SQL 迁移，例如 `0004_add_course_hot.up.sql` 和 `0004_add_course_hot.down.sql`，不修改已有迁移。
- 提交前根据改动范围运行对应的测试、lint 和 typecheck。

## 安全提示

不要提交真实的数据库密码、Redis 密码、SMTP 凭据、OAuth 密钥或 session secret。新增接口时请检查 `backend/internal/interface/web/router.go` 中的认证与管理员权限要求。

个人课程扫码导入与 MyCHU 预留入口见 [我的课程导入](docs/my-courses-import.md)。
