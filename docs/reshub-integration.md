# res-hub OIDC 接入

res-hub 是 CHDXK 的唯一账户、角色、会话和积分来源。CHDXK 作为 OIDC
Relying Party 使用标准 Authorization Code + PKCE；浏览器只参与授权跳转，
client secret、access token 和 ID token 均由 CHDXK 后端保存和使用。

## 接入边界

CHDXK 不再提供独立的登录账户来源。用户登录、注册、密码修改和账号状态由
res-hub 管理；CHDXK 只维护本地课程、评价等业务数据，以及 `issuer + sub`
到本地整数用户 ID 的投影。

配置中的 `reshub.issuer` 必须是 res-hub 的稳定公共 HTTPS Issuer。其余 OIDC
端点一律从 `/.well-known/openid-configuration` 读取，不在 CHDXK 配置中复制
部署地址。

```yaml
reshub:
  issuer: "https://res-hub.example.com"
  client_id: "chdxk"
  client_secret: "由 res-hub OIDC 接入管理签发的密钥"
  redirect_uri: "https://chdxk.example.com/api/auth/oidc/callback"
  logout_return_uri: "https://chdxk.example.com/"
  timeout: "3s"
  points_base_path: "/api/integrations/v1"
```

在 res-hub 中为同一个 client 登记：

- `authorization_code` grant；
- `openid profile res_hub_identity` scope；
- 与配置逐字符一致的回调地址；
- `post_logout_redirect_uri` 为 `logout_return_uri`；
- 积分功能所需的 `client_credentials` grant，以及 `points.read`、
  `points.write` scope。

OIDC 配置是 CHDXK 启动所需的必填配置；CHDXK 不会回退到本地注册、密码登录、
用户 API Key 或本地积分实现。开发环境也应使用 res-hub 的测试 client。

## 登录流程

1. 前端访问 `GET /api/auth/oidc/start?return_to=/目标路径`。
2. CHDXK 后端生成不可预测的 `state`、PKCE `code_verifier` 和 `nonce`，将它们
   与安全的站内返回路径保存到 Redis 会话。
3. 后端从 Discovery 得到 authorization endpoint，并请求
   `response_type=code`、`scope=openid profile res_hub_identity`、`state`、
   `nonce`、`code_challenge` 和 `code_challenge_method=S256`。
4. res-hub 完成登录后，将一次性 authorization code 和 `state` 回调到
   `/api/auth/oidc/callback`。
5. CHDXK 后端通过 token endpoint 的 `client_secret_basic` 兑换 code，验证
   ID Token 的签名、`iss`、`aud`、`exp`、`nonce`、`sid` 和 `at_hash`，再用
   access token 请求 userinfo。
6. userinfo 的 `sub` 与配置的 Issuer 一起写入本地身份投影；后端会话只保存
   OIDC access token 和 ID token。前端不会接触 token。

userinfo 使用 res-hub 当前直接 claims：`sub`、`name`、
`preferred_username`、`res_hub_identity_type`、`res_hub_role`、
`res_hub_effective_role`、`res_hub_permission_mode` 和可选的
`res_hub_operator_subject`。不根据本地旧字段猜测账号状态或权限。

## 当前请求和退出

除 OIDC 开始与回调外，受保护请求都从 HttpOnly Redis 会话取出 OIDC access
token，并以 `Authorization: Bearer` 请求 userinfo。userinfo 失败或令牌失效时，
CHDXK 清理本站会话并要求重新登录；身份投影每次以 res-hub 返回的当前角色更新。

退出时，CHDXK 使用 Discovery 的 `end_session_endpoint`，携带
`id_token_hint`、`client_id` 和已登记的 `post_logout_redirect_uri`，然后清理
本站会话。不会把 access token 当作注销凭证。

## 积分 API

CHDXK 使用 OIDC `client_credentials` 获取服务 access token，并按资源请求使用
`points.read` 或 `points.write` scope。获取 token 时使用
`client_secret_basic`；调用积分资源时只发送：

```http
Authorization: Bearer <OIDC access token>
```

当前资源路径为：

- `GET /api/integrations/v1/points/user:<id>`：查询余额；
- `GET /api/integrations/v1/points/ledger?subject=user:<id>`：查询流水；
- `POST /api/integrations/v1/points/operations`：创建积分操作；
- `GET /api/integrations/v1/points/operations/<operation_key>`：按业务键查询；
- `POST /api/integrations/v1/points/operations/<operation_id>/reverse`：冲正。

创建操作的 JSON 必须包含 `subject`、`amount`、`reason` 和
`operation_key`。同一 client 的业务键用于幂等重试；客户端不会自动重放普通
写请求。冲正使用 res-hub 返回的不可变 `operation_id`。

## 反向代理要求

反向代理必须在 SPA fallback 之前转发 Discovery、`/oidc/` 和积分 API，并按
res-hub 部署要求提供 HTTPS、mTLS 或可信代理配置。CHDXK 不再依赖 res-hub 的
私有 API listener、共享会话 Cookie 或非标准登录路径。
