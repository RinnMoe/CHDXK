# xk 接入 res-hub

这份文档描述 xk 侧的实现边界。`res-hub` 仍是独立仓库和独立服务，负责账号、角色、有效会话、账户状态和积分余额；xk 只保存本站业务所需的本地投影，并通过内网 HTTP 调用主站集成接口。

## 登录流程

生产环境将 `reshub.enabled` 设为 `true` 后，xk 不再提供本地注册、密码登录、密码重置或用户 API Key 路由。未登录用户访问 `/login` 或任意需要登录的页面时，前端跳转到 `GET /api/auth/sso/start?return_to=<本站路径>`。

`reshub.enabled: false` 仅用于客户端尚未在主站登记时的安全关闭状态，不是支持独立本地账号登录的运行模式；启用前，浏览器登录会保持不可用。

1. xk 生成一次性 `state` 和 PKCE `code_verifier`，只放入服务端会话。
2. xk 重定向到配置的主站授权页（res-hub 当前为 `/#/sso/authorize`），携带 `client_id`、精确回调地址、`state` 和 S256 challenge。授权参数必须放在 hash 路由内部的查询字符串中。
3. 主站完成登录（已登录时跳过密码输入）后，只把授权码和 `state` 返回 xk 的 `/api/auth/sso/callback`。
4. xk 后端通过内网 `POST /api/sso/v1/token` 兑换不透明集成会话令牌，再调用 `POST /api/sso/v1/introspect` 获取当前身份。
5. xk 在 `user_identities` 建立 `(issuer, subject)` 映射，现有课程、点评和关联表继续使用本地整数 `users.id`。令牌仅保存在服务端 Redis 会话，浏览器不保存令牌。
6. 之后每个请求都以内省结果为准；主站降权、删除主体、退出或会话失效会立即影响本站。主站不可用返回 503，失效会话返回 401，前端不会把这两类错误当成同一种情况反复跳转。

回调失败只重定向到本站 `/login?error=...`，不会把授权码、令牌或主站错误详情放进前端页面。本站退出时调用 `POST /api/sso/v1/logout`，再清除本站会话。

## 身份和权限

`subject` 是主站返回的带命名空间主体 ID，xk 不根据邮箱或昵称匹配账户。`effective_role` 映射为 xk 的 `user`、`admin` 或 `super_admin`，每次请求以内省快照为准；xk 不提供修改主站用户、管理员、密码或封禁状态的入口。管理员仍可使用 xk 自己的内容审核、公告、系统设置和站点统计能力，这些路由使用当前主站角色做鉴权。

配置型管理员可能没有积分账户。xk 展示积分时会把主站的 `no_points_account` 稳定错误直接返回，不自动创建账户。

## 积分流程

xk 的点评奖励仍先写入本地 `point_rewards`，再由现有异步任务调用主站积分接口：

- 查询余额：`GET /api/integrations/v1/points/{subject}`
- 查询流水：`GET /api/integrations/v1/points/ledger?subject=...&page=...&page_size=...`
- 变动积分：`POST /api/integrations/v1/points/operations`，请求包含 `subject`、带符号 `amount`、`reason` 和 `operation_key`
- 查询操作：`GET /api/integrations/v1/points/operations/{operation_key}`
- 完整冲正：`POST /api/integrations/v1/points/operations/{operation_id}/reverse`

这些路径对应 res-hub 的第一方集成契约；冲正路径使用主站返回的不透明 `operation_id`，业务键只用于幂等查询。若主站最终路径或字段名调整，只需要修改 `backend/internal/infrastructure/reshub/client.go`，不改业务层。当前客户端使用独立的 `client_id`/`client_secret` Basic 认证，所有请求只从 xk 后端发往内网地址。

每个奖励使用稳定键 `xk:point-reward:<本地奖励 ID>`。重复任务会得到主站幂等结果；网络失败会保留待处理奖励、记录尝试次数和下一次重试时间。删除点评时，已发放奖励调用主站完整冲正，待发放奖励只在本站标记取消。积分流水页明确展示主站返回的操作记录，不通过本站历史记录推算主站余额。

## 配置和数据库

从 `backend/config/config.example.yaml` 复制配置并填写：

```yaml
reshub:
  enabled: true
  authorize_url: "https://qbot.evian.asia/#/sso/authorize"
  internal_base_url: "http://127.0.0.1:3421"
  client_id: "xk"
  client_secret: "由部署系统注入"
  redirect_uri: "https://xk.example.com/api/auth/sso/callback"
  logout_return_uri: "https://xk.example.com/"
  timeout: "3s"
  points_base_path: "/api/integrations/v1"
```

主站登记的回调地址必须与 `redirect_uri` 完全一致。res-hub 授权页的实际地址是 `https://qbot.evian.asia/#/sso/authorize`；`client_secret` 不提交仓库。反向代理需公开主站 Web、登录及其所需的公开认证路由；兑换、内省、退出和积分请求走服务器内网（默认监听 `127.0.0.1:3421`）。

全新部署须按项目迁移机制从 `0001` 顺序应用到 `0014`，不能只执行 `backend/script/0014_add_reshub_identity_and_point_operations.up.sql`，因为它依赖前置的 `users` 和 `point_rewards` 表。`0014` 只创建 xk 的主体映射表和积分操作状态字段，不回填旧用户或旧积分数据，也不连接主站数据库。回滚时先禁用客户端入口，保留已发生的积分变动和新增表，不能只恢复本站余额。

## 联调顺序

1. 在主站登记 `client_id`、密钥、回调地址及积分权限。
2. 在 xk 应用迁移并填写内网地址，确认两个服务均使用 HTTPS 外部入口和独立服务会话密钥。
3. 用测试主体验证：已登录主站直达、首次登录、取消授权、PKCE 重放、主站退出、角色降权、无积分账户、奖励幂等、余额不足和完整冲正。
4. 再开启点评奖励和生产客户端。真实主站网络联调需要单独记录，仓库测试不会替代这一步。
