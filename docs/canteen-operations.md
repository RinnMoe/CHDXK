# 食堂评价本机运维

本文件覆盖本机开发与验收所需的启动、迁移、开关、媒体与排障步骤；不包含生产部署流程。

## 启动

```bash
docker compose -f docker/docker-compose.yaml up -d postgres redis

# 空库按序应用迁移（含 canteen_*）
for f in $(ls backend/script/*.up.sql | sort); do
  psql "host=localhost port=5432 user=postgres password=postgres dbname=jcourse sslmode=disable" \
    -v ON_ERROR_STOP=1 -f "$f"
done

cd backend
cp -n config/config.example.yaml config/config.yaml
go run cmd/api/main.go --config config/config.yaml
go run cmd/taskworker/main.go --config config/config.yaml   # 日任务与异步任务
```

前端：`cd frontend && pnpm dev`（Vite 代理 `/api` 到 8080）。

## 功能开关

系统设置中三项：

| Key | 默认 | 作用 |
| --- | --- | --- |
| `canteen.enabled` | `false` | 食堂入口与全部餐饮接口总开关 |
| `canteen.reviews.enabled` | `true` | 到店/菜品评价与投票 |
| `canteen.media.enabled` | `true` | 图片上传与审核 |

本机可通过后台「站点配置 → 系统设置」修改，或直接调用管理员设置接口。

## 媒体

- 目录：`canteen.media.dir`（默认 `.local/canteen-media`，已被 `.gitignore` 忽略）。
- 发布编排 `docker/release/compose.yaml` 已把 `canteen_media` 卷挂载到 `/app/.local/canteen-media`；API 与 taskworker 镜像都在构建阶段预创建 `.local/canteen-media` 并设置 `app:app` 所有权。重建 API 容器不会丢图片；已用命名卷验证非 root 写入、跨容器读取，并实际启动 API + taskworker 确认两者持续运行。备份/恢复时请与 PostgreSQL 备份一起处理该卷。
- 单张上限、每日上传、像素与每评价张数见 `config/config.example.yaml` 的 `canteen.media`。
- 上传经过服务端重编码，EXIF（含 GPS）不会保留。
- 未审核图片不进入公开评价列表；直链读取会校验所有者/管理员。

## 缓存与任务

- 缓存键：`jcourse:canteen:home:v2`、`jcourse:canteen:store:{id}:detail:v2:{viewer}`、`jcourse:canteen:rankings:v2:*`。`v2` 用于避免旧版本缓存缺少榜单解释/菜品封面字段。
- 任何评价、店铺、食堂、菜品写操作都会按 `jcourse:canteen:*` 清理缓存。
- 日任务 `canteen:refresh_rankings` 全量重算店铺统计并清缓存；由 `taskworker` 按 `canteen.rating.refresh_cron` 调度。
- 手动验证任务：启动 taskworker 后观察日志，或调用管理员接口前先清理 Redis 中 `jcourse:canteen:*`。

## 健康检查

```sql
-- 统计与摘要一致性
SELECT id, rating_count, rating_avg, rating_score, review_event_count
FROM canteen_stores ORDER BY id;

SELECT store_id, user_id, weighted_score, window_size FROM canteen_user_rating_summaries LIMIT 20;

-- 孤立媒体（有记录但未挂到评价）
SELECT m.id, m.status FROM canteen_media m
LEFT JOIN canteen_review_media rm ON rm.media_id = m.id
WHERE rm.media_id IS NULL AND m.status IN ('pending','approved');

-- 已合并但仍有评价的菜品（合并实现会迁移事件，正常情况下应为 0）
SELECT dish_id, count(*) FROM canteen_dish_review_events
JOIN canteen_dishes d ON d.id = dish_id
WHERE d.status = 'merged' GROUP BY dish_id;
```

## 排障

| 现象 | 排查 |
| --- | --- |
| 用户端仍显示入口但接口 404 | `canteen.enabled` 是否为 false，或 API 未重启加载设置 |
| 发布评价返回 429 | 24 小时日限或短时间频率策略；检查 `canteen.review.*` |
| 发布评价返回 409 | 页面的 `expected_updated_at` 过期，刷新后重试 |
| 图片上传 415 | 仅 JPEG/PNG；确认文件未损坏（Go 解码会校验 PNG CRC） |
| 图片上传 413 | 超过 5 MiB；前端会先压缩到长边 2048 |
| 审核后图片仍不显示 | 检查 `canteen_media.status` 与 `canteen_review_media` 关联；缓存是否清理 |
| 排行榜不变 | 日任务未运行可接受；写操作应即时更新统计，检查缓存 |
| 菜品合并后评分异常 | 合并会迁移 `canteen_dish_review_events` 并重算摘要/统计 |

## 回滚

本机验证以重建为主：

```bash
docker compose -f docker/docker-compose.yaml down -v
```

生产不在本阶段范围内；迁移脚本的 down 仅用于开发与测试。
