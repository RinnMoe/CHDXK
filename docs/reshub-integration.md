# res-hub OIDC 接入

CHDXK 保留独立本站账号、密码登录、角色与处罚。res-hub 是主要授权登录提供方，也是积分来源。CHDXK 使用 Authorization Code + PKCE，密钥及授权令牌只由后端保存。绑定关系为多个本站账号对应一个 `issuer + sub`；首次授权无绑定时创建本站账号，有多个绑定时由用户选择。

## 接入边界

本站初始超级管理员可以独立账密登录；所有用户可以在本站设置密码与绑定资料站。资料站管理员继承本站管理员身份，本地管理员身份和分项权限由本站维护。注册开关与封禁、禁言只影响本站。


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

OIDC 配置用于接入资料站授权和积分；本站账密登录使用本地密码与 Redis 会话，积分不会回退为本地积分。开发环境使用 res-hub 测试 client。

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
6. userinfo 的 `sub` 与配置 Issuer 用于查找绑定的本站账号。会话保存授权令牌、选定的本站账号 ID 和认证版本；多个绑定时先显示账号选择页。前端不会接触 token。

userinfo 使用 res-hub 当前直接 claims：`sub`、`name`、
`preferred_username`、`res_hub_identity_type`、`res_hub_role`、
`res_hub_effective_role`、`res_hub_permission_mode` 和可选的
`res_hub_operator_subject`。不根据本地旧字段猜测账号状态或权限。

## 当前请求和退出

受保护请求校验选定本站账号、认证版本与本站封禁状态。资料站登录会话另外请求 userinfo 并校验该身份仍绑定选定账号；失效时清理本站会话。角色使用 `res_hub_role`，不依据普通用户模式 `res_hub_effective_role` 或 `res_hub_permission_mode`。本站主动授予的管理员身份保留，资料站超管继承本站管理员，不自动获得本站超级管理员身份。

本站账密会话不依赖资料站在线状态。退出登录只清理 CHDXK 会话，不注销资料站会话。密码更改、绑定更换、解绑使旧认证版本失效。


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


## 初始管理员和积分账户

首次启动的本地初始超级管理员、账密登录、资料站账号绑定、账号选择、分项权限，以及无积分账户的处理，参见 [初始管理员、频道账号和积分](bootstrap-and-points.md)。

频道在 `res_hub_identity` scope 的身份信息中返回 `res_hub_has_points_account`。配置管理员通常没有积分账户，用户库账号具有积分账户；课程榜会区分可管理站点与可使用积分这两种能力。
