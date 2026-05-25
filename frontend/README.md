# jcourse Frontend

这是 `jcourse` 的 React/Vite 前端，提供课程检索、教师检索、课程详情、点评发布与浏览、个人课程、积分、API Key、设置和后台管理页面。

## 技术栈

- React 19 + TypeScript
- Vite
- React Router
- TanStack Query
- Tailwind CSS + shadcn/ui 风格组件
- MSW 本地 mock
- Vitest、ESLint、Prettier

## 目录结构

```text
frontend/
├── public/          # 静态资源与 MSW worker
├── src/
│   ├── api/         # API client、DTO 与接口封装
│   ├── components/  # UI primitives、布局与业务组件
│   ├── config/      # 品牌与认证相关配置
│   ├── contexts/    # React context
│   ├── hooks/       # 业务数据 hooks
│   ├── lib/         # 通用工具
│   ├── mocks/       # MSW handlers 与 mock fixtures
│   ├── pages/       # 页面组件
│   ├── App.tsx
│   └── router.tsx
├── package.json
└── vite.config.ts
```

## 本地启动

### 1. 安装依赖

```bash
pnpm install
```

项目声明的包管理器为 `pnpm@11.2.2`。

### 2. 启动开发服务器

```bash
pnpm dev
```

默认访问地址为 `http://localhost:5173`。开发环境下，Vite 会将 `/api` 代理到 `http://localhost:8080`。

### 3. 配合后端开发

先在仓库根目录启动 PostgreSQL 和 Redis，再启动后端 API：

```bash
docker compose -f docker/docker-compose.yaml up -d postgres redis
cp backend/config/config.example.yaml backend/config/config.yaml
cd backend
go run cmd/api/main.go --config config/config.yaml
```

前端 API 基础路径定义在 `src/api/constants.ts`，默认为相对路径 `/api`。

## Mock 模式

开发环境可启用 MSW mock，不依赖真实后端：

```bash
VITE_ENABLE_MOCKS=true pnpm dev
```

mock 入口在 `src/main.tsx`，handlers 和 fixtures 位于 `src/mocks`。MSW worker 文件位于 `public/mockServiceWorker.js`。

## 常用命令

```bash
pnpm dev        # 启动 Vite 开发服务器
pnpm build      # TypeScript 构建检查并打包
pnpm lint       # 运行 ESLint
pnpm typecheck  # 仅运行 TypeScript 检查
pnpm test       # 运行 Vitest
pnpm preview    # 本地预览生产构建
pnpm format     # 使用 Prettier 格式化
```

## 编码约定

- 页面文件使用 kebab-case，例如 `course-detail-page.tsx`。
- React 组件使用 PascalCase，hooks 使用 `useThing` 命名。
- 优先使用已有 UI primitives：`src/components/ui`。
- 业务组件放在对应 feature 目录，例如 `src/components/course`、`src/components/review`。
- API 请求封装放在 `src/api`，数据读取和变更 hooks 放在 `src/hooks`。
- 使用 `@/` 路径别名引用 `src` 下模块，例如 `@/components/ui/button`。

## API 行为

`src/api/client.ts` 统一处理请求：

- 默认携带 `credentials: "include"`，用于 session cookie。
- 非 GET/HEAD/OPTIONS 请求会先请求 `/api/auth/csrf` 并添加 `X-CSRF-Token`。
- 非 2xx 响应会抛出 `HttpError`。

后端 CORS 默认允许 `http://localhost:5173` 和 `http://127.0.0.1:5173`。

## 测试与提交前检查

提交前建议按改动范围运行：

```bash
pnpm lint
pnpm typecheck
pnpm test
pnpm build
```

涉及 hooks、API 数据转换或复杂 UI 状态时，补充聚焦的 Vitest 测试。

## 构建产物

生产构建输出到 `dist/`：

```bash
pnpm build
pnpm preview
```

部署时需要让 `/api` 转发到后端服务，或在同域下提供后端 API。
