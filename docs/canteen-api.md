# 食堂评价 API（v1）

本文件描述 CHDXK 餐饮领域的对外 HTTP 契约，供 MyCHU 原生端和脚本复用。当前实现与验证均在本机完成，接口前缀为 `/api`。

## 认证与通用约定

- 所有接口要求登录会话（Cookie）；写请求还需要 `X-CSRF-Token`，与现有 CHDXK 前端一致。
- MyCHU 通过现有 `/api/auth/native/exchange` 换取本站会话，不另建账号体系。
- 分页统一返回 `{ items, total, page, page_size }`，`page` 从 1 开始，`page_size` 上限 100。
- 时间使用 RFC3339 字符串。
- 错误返回 `{"error": "..."}`，以 HTTP 状态区分：`400` 参数、`403` 权限/未审核、`404` 不存在、`409` 冲突（含乐观锁）、`413/415` 媒体限制、`429` 频率限制。
- 创建到店评价必须携带 `Idempotency-Key` 请求头（UUID）。重复提交返回 `created:false` 与既有评价。
- 修改评价使用 `expected_updated_at` 做乐观锁；版本不一致返回 409。

## 用户端

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/canteen/home` | 食堂/店铺/评价数聚合与推荐 |
| GET | `/api/canteen/halls` | 校区食堂列表 |
| GET | `/api/canteen/stores` | 店铺检索：`q, campus, hall_id, category, sort, page, page_size` |
| GET | `/api/canteen/stores/{id}` | 店铺详情（统计、五维、我的摘要、菜品） |
| GET | `/api/canteen/rankings` | 排行榜：`campus, hall_id, sort, page, page_size` |
| GET | `/api/canteen/stores/{id}/dishes` | 店铺菜品（含封面与实拍数） |
| GET | `/api/canteen/stores/{id}/dishes/{dishID}` | 菜品详情：三维评分、实拍、菜品评价 |
| POST | `/api/canteen/stores/{id}/dishes` | 提交菜品（用户为待审核，管理员直接上架） |
| GET | `/api/canteen/dishes/suggest?store_id=&q=` | 菜品搜索建议（名称/别名 + 编辑距离） |
| POST | `/api/canteen/dishes/{id}/reviews` | 菜品三维评分，可带 `media_ids` |
| GET | `/api/canteen/stores/{id}/reviews` | 评价流：`order=latest\|best\|worst`、`filter=all\|with_media\|high\|low`、`history=true`、`cursor`、`limit` |
| POST | `/api/canteen/stores/{id}/reviews` | 创建到店评价（五维、可选文字/标签/图片） |
| PATCH | `/api/canteen/reviews/{id}` | 修改到店评价 |
| DELETE | `/api/canteen/reviews/{id}` | 作者软删除 |
| POST | `/api/canteen/reviews/{id}/vote` | 有用/无用投票，`vote_type` 为 `1`、`-1` 或 `0`（取消） |
| GET | `/api/canteen/reviews/mine` | 我的到店评价 |
| GET | `/api/canteen/contributions` | 我的菜品评分与实拍 |
| POST | `/api/canteen/reports` | 举报：`subject_type=store\|review\|dish_review\|media` |
| GET | `/api/canteen/search?q=` | 统一搜索，返回 `{stores, dishes}` |
| POST | `/api/canteen/media/upload-intents` | 申请上传：`{purpose, mime, bytes}` |
| PUT | `/api/canteen/media/{id}/content` | 代理上传原始字节（JPEG/PNG） |
| POST | `/api/canteen/media/{id}/complete` | 确认上传完成（幂等） |
| GET | `/api/canteen/media/{id}` | 读取图片；仅 `approved` 对他人可见 |

### 创建到店评价请求体

```json
{
  "visited_at": "2026-10-31T12:00:00+08:00",
  "taste": 5,
  "value": 4,
  "queue": 3,
  "hygiene": 4,
  "service": 5,
  "content": "可选文字，最多 2000 字",
  "tags": ["分量足", "上菜快"],
  "media_ids": [12, 13]
}
```

五维必须为 1–5；标签最多 5 个且必须在后端允许列表内；`media_ids` 最多 3 个，只能引用本人上传的图片。

## 媒体上传流程

1. `POST /api/canteen/media/upload-intents` 返回 `{media_id, upload:{mode:"proxy", method:"PUT", url, headers}}`，30 分钟内有效。
2. 客户端 `PUT` 到返回的 `url`，携带原始图片字节；服务端限制 MIME、大小与像素数，重编码去除 EXIF。
3. `POST /api/canteen/media/{id}/complete` 确认；图片进入 `pending` 等待管理员审核。
4. 审核通过后，图片出现在评价列表与店铺页；未通过或待审核图片仅上传者与管理员可见。
5. 评价创建时通过 `media_ids` 关联图片；图片不会自动公开。

限制（可通过 `canteen.media.*` 配置覆盖）：单张 5 MiB、最多 3 张/评价、每日 20 次上传、最大 8 MP，仅 JPEG/PNG。

菜品评分同样支持 `media_ids`，图片会进入菜品实拍图鉴；只有 `approved` 图片对其他人可见。

## 榜单解释字段

`GET /api/canteen/rankings` 的每条店铺额外返回：

- `confidence`：`样本较少`（<3 人）、`参考`（3–9 人）、`较可信`（≥10 人），仅描述样本量。
- `summary_tags`：近 30 天评价中出现频率最高的 3 个标签。
- `dish_count`：在架菜品数。
- `dish_photo_count`：已通过审核的菜品实拍数。

综合分仍为贝叶斯平滑，菜品与实拍数量不参与计算。

## 评价流与游标

`GET /api/canteen/stores/{id}/reviews` 默认只返回每位用户最新一条到店评价；`history=true` 展开全部历史。

- `filter=with_media` 只看有已审核图片的评价；`high` 为综合分 ≥4；`low` 为综合分 ≤2。
- `limit` 默认 20，最大 100。
- 响应包含 `next_cursor`；把该值作为 `cursor` 传回即可加载下一页。游标与排序绑定，切换 `order` 后必须丢弃旧游标。

## 菜品图鉴

`GET /api/canteen/stores/{id}/dishes/{dishID}` 返回：

```json
{
  "dish": {"id": 1, "name": "牛肉面", "cover_url": "/api/canteen/media/12", "photo_count": 3},
  "photos": [{"id": 12, "url": "/api/canteen/media/12", "width": 1280, "height": 960}],
  "reviews": [],
  "review_total": 0
}
```

菜品图片走同一套上传/审核流程，`purpose=dish`；只有审核通过的图片出现在 `photos`。

## 首页发现

`GET /api/canteen/home` 在原有 halls/top_stores/new_stores/latest_reviews/stats 之外返回：

- `hero`：今日推荐（店铺/菜品、理由、图片、综合分）。
- `feeds`：`recommended_store`、`trending`、`new_store`、`stable_choice`、`recent_photo` 决策卡片。
- `hot_dishes`：按菜品综合分排序的热门菜品。

`GET /api/canteen/search?q=` 一次返回匹配的店铺与菜品，用于首页搜索框。

## 管理端

管理端路径前缀 `/api/admin/canteen`，要求管理员身份。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST/PUT | `/halls[/{id}]` | 食堂维护 |
| GET/POST/PUT | `/stores[/{id}]` | 店铺维护 |
| POST | `/stores/{id}/status` | 店铺状态（`pending/active/offline/rejected`） |
| POST | `/stores/{id}/merge` | 合并重复店铺 |
| GET | `/dishes` | 菜品列表：`store_id, status, q` |
| POST | `/dishes/{id}/status` | 菜品审核/上下架 |
| POST | `/dishes/{id}/merge` | 合并菜品（历史评分迁移到目标菜品） |
| GET | `/media` | 图片审核队列：`status, page, page_size` |
| POST | `/media/{id}/review` | 审核：`{status:"approved"\|"rejected", reason}` |
| GET | `/reports` | 举报队列：`status, page, page_size` |
| POST | `/reports/{id}/resolve` | 处理：`{status:"resolved"\|"dismissed", note}` |
| POST | `/reviews/{id}/hide` | 隐藏评价 |

## 排序与评分语义

- 单次到店综合分：五维权重 `35/20/15/20/10`。
- 用户摘要：最近 3 次有效到店按 `0.6/0.3/0.1` 加权，事件不足按权重和归一。
- 店铺排行：`(n*R + 5*C) / (n + 5)`，`n` 为去重评价人数，`C` 为全平台用户摘要均值。
- 菜品：口味/分量/性价比 `0.5/0.25/0.25`，同样使用最近 3 次与贝叶斯平滑。
- 店铺 `rating_count` 为去重用户数，`review_count` 为到店事件数，二者不同。

## 变更策略

- v1 只做字段新增；删除或改义字段需要提升路径版本。
- 功能开关：`canteen.enabled`、`canteen.reviews.enabled`、`canteen.media.enabled`；关闭后接口返回 404 或 403。
- 所有接口不依赖 MyCHU 地图或 Flutter 实现，CHDXK 可独立运行。
