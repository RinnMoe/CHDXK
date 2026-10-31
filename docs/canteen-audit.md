# 食堂评价交付审计说明

本文件供独立审计 agent 使用，说明本轮交付的提交拆分、验证证据、已知限制与建议审查点。

## 提交与 PR 拆分

四个提交按依赖顺序堆叠，后一个 PR 以前一个分支为 base；合并时按 1 → 2 → 3 → 4 顺序进行。

| 顺序 | 提交 | 范围 |
| --- | --- | --- |
| 1 | `feat(canteen): add schema, domain and persistence foundation` | 迁移 0022–0024、领域模型、仓储、媒体存储/图像处理、配置与设置键、领域/仓储测试 |
| 2 | `feat(canteen): add application services and HTTP API` | 应用服务、控制器、路由、容器接线、异步任务、应用层测试 |
| 3 | `feat(canteen): add web UI for discovery, reviews, dishes and moderation` | 前端页面、组件、hooks、API 客户端、路由、文案目录 |
| 4 | `docs(canteen): document API, operations and workflows` | API/运维/后台/原生授权文档、README、缓存键文档、本审计说明 |

## 验证证据

本机执行：

```bash
cd backend
gofmt -l .                      # 仅剩两个历史未格式化文件
go vet -tags test ./...
go test -tags test ./...        # 全部通过

cd frontend
pnpm typecheck
pnpm lint                       # 0 error，仅历史 warning
pnpm test                       # 17 文件 / 46 用例
pnpm build
```

迁移：

- 空库按序应用 0001–0024 成功。
- 0022/0023/0024 的 down 与重新 up 在 `jcourse_migration_check` 验证通过。

本机 API 冒烟（`jcourse_smoke`）：

- 榜单：`confidence`、`summary_tags`、`dish_count`、`dish_photo_count` 返回。
- 评价流：默认每人最新一条；`history=true` 展开；`filter=with_media|high|low`；`next_cursor` 第二页不重复。
- 菜品：创建/三维评分；菜品详情返回封面、实拍与评价；上传→待审核→审核通过→图集可见。
- 首页：`hero`、`feeds`、`hot_dishes` 返回；统一搜索返回店铺/菜品。
- 媒体：JPEG/PNG 重编码去 EXIF、直链鉴权、待审核不可见。
- 举报：创建、管理员列表、处理完成。

## 审计整改（fix/canteen-audit）

针对独立审计报告，已在 `fix/canteen-audit` 分支处理：

- 评价编辑：Repository 在事务内锁定评价行并校验 `expected_updated_at`，缺少版本按 400 拒绝；新增真实 PostgreSQL 双请求并发测试。
- 媒体直链：审核通过不再等于公开；必须存在可见父内容（active 评价，或 active 菜品 + 可用店铺）。隐藏评价/菜品后图片退回所有者/管理员可见；新增仓储可见性测试。
- 店铺合并：事务内迁移到店评价、用户摘要、菜品与实拍关联，重算源/目标统计；目标店铺不能已合并或处于已拒绝状态；新增合并聚合测试。
- 菜品合并：补齐 `canteen_review_dishes` 关联迁移；新增关联迁移测试。
- 权限：禁言/暂停用户不能新增菜品；菜品评分关联的到店评价必须属于当前用户、同一店铺且为 active。
- 发布：`docker/release/compose.yaml` 增加 `canteen_media` 持久化卷。
- 契约：CORS 补 `PATCH` 与 `Idempotency-Key`；幂等键校验 UUID；`visited_at` 限制未来 24 小时至一年前；创建/更新评价返回完整 DTO；前端区分评价加载失败与空列表。

## 已知限制

- 全部只在本机验证，未部署生产，未接入真实 MyCHU 设备。
- 浏览器 UI 走查未完成：本地 OAuth 会话注入被重定向到登录页，因此 UI 以类型检查、lint、Vitest、构建和 API 冒烟验证；建议审计时补一次真实登录后的手工页面检查。
- 文案目录只更新了现有 catalog，没有执行全量 `manage-copy.mjs --extract`；仓库原有 500+ 未纳管文案会产生大量无关 diff。
- 缓存键已升级为 `jcourse:canteen:*:v2:*`，升级部署时需要清理旧 `jcourse:canteen:*`（本机通过写操作失效验证）。
- 榜单标签在冒烟数据中为空，因为测试评价未带标签；仓储测试已覆盖带标签场景。

## 建议审计重点

1. `canteen_review.go` / `canteen_dish.go` 的聚合 SQL 与行锁、软删重算、菜品合并迁移。
2. `canteen_query.go` 的 latest-per-user 子查询、游标条件、排序稳定性。
3. 媒体上传的路径安全、大小/像素限制、EXIF 去除、待审核可见性。
4. 管理端路由的 `RequireAdmin` 与作者/管理员 guardian 判定。
5. 榜单解释字段的统计口径（30 天标签、菜品/实拍计数、置信度阈值）。
6. 前端无限滚动在切换排序/筛选时的游标重置与竞态。
