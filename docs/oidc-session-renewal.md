# OIDC 登录续期

CHDXK 在服务端自动轮换主站访问令牌，不把刷新令牌发送给浏览器或原生客户端。保留现有 Cookie、CSRF、账号选择和每次请求的 userinfo/账号权限检查。

## 生命周期与兼容

- 当前示例配置 `session.max_age: 604800` 为 7 天；没有配置该字段时，代码默认 30 天。此次不更改配置。
- 登录时记录本站绝对截止时间。访问令牌剩余不足 30 秒时，在执行业务处理前续期；续期不延长本站截止时间，也不延长主站父登录会话。
- 浏览器 OIDC callback 和 native exchange 都保存刷新凭证；账号选择/密码变更引起的 session ID 轮换保留这些凭证与截止时间。
- 旧会话、主站尚未支持刷新或 client 未启用刷新授权时，继续原有行为；用户重新登录后才能获得刷新凭证。本站密码登录不调用主站刷新。
- 登录、普通授权入口、回调和退出不被过期的刷新凭证阻挡；绑定操作和账号选择会先续期。

## 并发和故障

复用现有 Redis，通过刷新令牌摘要协调多个 API 实例。一次轮换结果缓存 2 分钟；发出轮换请求前记录消费标记，保留本站配置的会话期限。浏览器请求取消后，已经开始的轮换仍在独立、最多 15 秒的 context 内完成并发布结果。

明确的暂不可用或客户端配置错误返回 503 并保留会话。明确失效、撤销、绝对截止时间已到返回 401 并清除会话。网络中断、响应丢失或成功响应不完整时，无法证明一次性刷新令牌是否已消费，因此要求重新登录，禁止重发旧令牌。短期结果缓存到期后仍携带旧刷新令牌的异常会话也要求重新登录。

本站 session 保存与注销/ID 轮换通过 Redis 锁和撤销标记协调。后台续期若晚于注销或账号切换完成，将拒绝保存旧 session，避免恢复已退出的登录态。底层连接设有明确超时，退出操作不会因浏览器取消请求而跳过撤销。

刷新发生在业务 handler 前，服务端不会自动重放点评、积分等业务写请求。续期后仍由现有认证逻辑重新校验身份和权限。

## 主站契约及后续发布

`POST /oidc/token` 使用 `client_secret_basic` 和表单字段 `grant_type=refresh_token`、`refresh_token`。沿用原 scope。成功响应包含 `access_token`、`token_type=Bearer`、正数 `expires_in` 和轮换后的 `refresh_token`；不要求新的 ID Token，本站保留原 ID Token。

主站须显式为 CHDXK client 启用 `refresh_token` grant，浏览器和 native 授权码交换才返回刷新令牌。主站本地交付与依赖详见 [主站交接](oidc-refresh-provider-handoff.md)。

本次仅本地修改和验证，不部署、推送或修改线上 client 配置。后续人工批准发布时，应先确认主站所需提交和 client 配置就绪，再发布 CHDXK；新登录后验证跨访问令牌期限仍可浏览与写入，退出、父会话撤销和账号封禁仍生效。不得把本地测试等同于线上登录验收。

## 本地验证

基础命令：

~~~powershell
cd backend
go test -tags test ./internal/interface/web/middleware ./internal/interface/web/controller ./internal/application ./internal/infrastructure/reshub
go build ./...
go vet -tags test ./...
~~~

Redis 集成测试需要显式指定隔离实例（不要指向生产）：

~~~powershell
$env:JCOURSE_TEST_REDIS_ADDR = "127.0.0.1:<isolated-port>"
go test -race -tags test ./internal/infrastructure/reshub -run TestRefreshCoordinator -count=1
~~~

未设置该变量时 Redis 集成测试会跳过，普通包测试通过不能替代这项证据。

跨服务本地联调需要 res-hub agent 启动真实 Python app、隔离 SQLite、两个合成刷新链的 loopback fixture。通过环境变量 `JCOURSE_TEST_OIDC_FIXTURE` 指定其临时凭据 JSON（不入 Git），分别运行 `TestRefreshAgainstLocalResHub` 和 `TestOIDCRefreshAgainstLocalProviderAndRedis`。这两项测试会消费一次性凭据，重复运行需创建新的 fixture。

2026-10-19 实际结果：上述两项跨服务联调通过；隔离 Redis 下 middleware、controller、application、reshub 四个相关包的竞态测试通过，另有跨实例续期、请求取消、失效重放、缓存发布失败和注销/账号轮换交错回归。后端 `go build ./...`、`go vet -tags test ./...` 通过。主站 47 项 OIDC/native/refresh 回归通过。临时 Python fixture 和合成凭据已清理。

未运行依赖 PostgreSQL 扩展的全仓库数据库测试，也未做公网 HTTPS、浏览器长期停留或真机验收。
