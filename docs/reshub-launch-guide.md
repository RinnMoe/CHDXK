# 已上线 res-hub 接入频道课程榜

适用情境：res-hub 已经运行，频道课程榜（CHDXK）准备首次上线。课程榜复用主站账号、当前权限和积分，不需要复制账号数据库，也不需要重新注册用户。

## 普通管理员：日常管理

技术管理员完成首次接入后，在主站“管理功能 → OIDC 接入管理”中找到“频道课程榜”。可查看配置、启用或停用接入；停用会使课程榜用户无法继续使用主站登录。不要更改回调地址或授权范围，除非技术管理员提供了新的准确值。

课程榜中，主站使用管理员权限的账号可管理课程、公告、点评和站点设置，其他账号作为普通用户。主站管理员若主动选择普通用户模式登录，课程榜也遵循该次会话的普通用户权限；模拟其他用户时遵循被模拟者权限。无需在课程榜另行授予管理员身份。

需要调整课程榜文案、导入长安大学课程、处理点评时，在课程榜后台操作。密码、账号封禁和管理员身份在主站处理。发现无法登录时，记录时间、页面提示和具体操作，交给技术管理员；不要收集用户密码。

## 技术管理员：首次登记

1. 确认主站已有可用的 OIDC 服务。浏览器访问 `https://主站域名/.well-known/openid-configuration` 应返回 JSON，其 `issuer` 必须是稳定公共地址。若返回首页 HTML，交由负责主站代理配置的技术管理员处理。
2. 使用主站超管权限进入“管理功能 → OIDC 接入管理”，创建客户端：

| 字段 | 填写值 |
| --- | --- |
| 客户端标识 | `chdxk`（稳定且唯一） |
| 名称 | 频道课程榜 |
| 回调地址 | `https://课程榜域名/api/auth/oidc/callback` |
| 退出返回地址 | `https://课程榜域名/` |
| 授权类型 | `authorization_code`、`client_credentials` |
| 授权范围 | `openid`、`profile`、`res_hub_identity`、`points.read`、`points.write` |

3. 保存创建时显示的客户端密钥，交给课程榜技术管理员写入后端环境变量或未跟踪的配置文件。密钥只显示一次；丢失后使用轮换功能。不要填写在前端变量、公告或工单正文中。
4. 如果现有主站启用了 mTLS，由负责网关的技术管理员提供课程榜服务客户端证书和私钥，在主站客户端登记证书 SHA-256 指纹，并确认网关转发可信证书断言。CHDXK 已支持客户端证书，API 与 worker 都必须配置下面的文件路径。此任务没有修改生产网关或系统证书库，这部分上线验证列为范围外阻塞。
5. 将以下值写入 CHDXK 后端配置；前端使用同源 `/api`，不填写客户端密钥：

```yaml
reshub:
  issuer: "https://主站域名"
  client_id: "chdxk"
  client_secret: "创建时取得的密钥"
  redirect_uri: "https://课程榜域名/api/auth/oidc/callback"
  logout_return_uri: "https://课程榜域名/"
  timeout: "3s"
  points_base_path: "/api/integrations/v1"
  client_cert_file: "/run/secrets/chdxk-client.pem" # mTLS 时填写，普通 HTTPS 留空
  client_key_file: "/run/secrets/chdxk-client-key.pem"
  ca_file: "" # 使用私有服务端 CA 时填写 PEM 文件；公共 CA 留空
```

证书和私钥必须同时配置。也可使用 `JCOURSE_RESHUB_CLIENT_CERT_FILE`、`JCOURSE_RESHUB_CLIENT_KEY_FILE`、`JCOURSE_RESHUB_CA_FILE` 环境变量。证书只由后端读取，不能放进前端资源或 Git；无效文件会阻止服务启动。私有 CA 在进程内加入系统信任池，不修改操作系统证书库，所有 HTTPS 请求仍校验服务端证书和主机名。参见 [Go TLS 配置](https://pkg.go.dev/crypto/tls#Config)。

6. 在独立 PostgreSQL 库按编号运行全部 `backend/script/*.up.sql`；启动 Redis、CHDXK API 与 taskworker。生产设置独立的 `session.secret`、`session.secure=true` 和准确的 CORS 来源。点评热度和积分奖励由 worker 处理，不能只运行 API。
7. 主站原有 OIDC 功能已经启用时无需修改全局 Issuer、签名密钥或重启主站。若未启用，由主站技术管理员按其 `docs/oidc-admin-configuration.md` 完成部署配置；普通管理员只负责站点运维。

## 上线验收与故障处理

- 普通用户：进入课程榜 → 跳转主站 → 登录 → 回到原页面；不能进入后台。
- 管理员：在主站选择管理员权限后进入课程榜，可访问课程、公告和设置管理。主动降为普通用户权限后，课程榜下一次请求应撤销管理能力。
- 积分：普通用户能查看主站余额；发布满足奖励条件的点评后，由 worker 发放并在主站流水确认。主站配置管理员没有积分账户属于正常情形。
- 退出：退出课程榜后 `/api/auth/me` 返回 401；重新访问需重新建立本站会话。
- 停用客户端、撤销主站会话或变更账号权限后，课程榜下一次受保护请求遵循主站当前状态。

回调不一致：核对协议、域名、端口、路径和尾部斜杠。`invalid_client`：核对标识和密钥，确认是否已经轮换。Discovery 返回 HTML：检查代理路由。用户能登录但积分失败：检查 `client_credentials`、积分 scope、worker 和主站流水。mTLS 拒绝：由网关技术管理员检查证书和可信代理，普通管理员不修改系统配置。

## 本地验证记录（2026-10-05）

两个真实后端以 HTTP 在本机运行：res-hub `127.0.0.1:3421`，CHDXK `127.0.0.1:18080`。独立数据库为 res-hub 测试 SQLite 和 PostgreSQL `chdxk_integration`，Redis 使用 DB 5。

已通过实际 HTTP 请求验证授权码 + PKCE、签名 ID Token、userinfo、本站 Redis 会话、普通用户/超管角色、后台 403/200、普通用户积分查询及退出后 401。主站切换权限后，已有课程榜会话下一次后台请求失去管理权限。发布点评后，worker 实际发放 1 积分到主站，余额与流水均确认，测试开关随后恢复原值。

客户端证书通过项目内临时 TLS 服务验证：有可信客户端证书时成功，无客户端证书时握手失败；同时校验私有服务端 CA。此记录不代表生产网关 mTLS 或全校课程采集验证。8080 本机绑定失败后仅调整了测试端口，没有修改系统端口配置。

res-hub 在 `OIDC_REQUIRE_HTTPS=false` 时支持 loopback HTTP 回调，仅限 `localhost`、`127.0.0.1`、`::1`；生产 HTTPS 校验保持启用。Vite 本地代理可通过进程级 `VITE_API_PROXY_TARGET=http://127.0.0.1:18080` 调整。
