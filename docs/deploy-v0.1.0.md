# CHDXK v0.1.0：Docker 部署与资料站登录配置

适用于资料站 ResHub 已上线、新部署 CHDXK 的情境。示例使用 `https://courses.example.com` 作为课程榜域名、`https://res.example.com` 作为资料站 OIDC 地址，请替换为自己的地址。本版提供源码和 Docker 构建文件，未发布公共镜像仓库；应用镜像在部署服务器本地构建。

## 1. 在资料站登记 CHDXK

由资料站超级管理员登录资料站，进入「OIDC 接入管理」（前端路由 `/oidc-clients`），创建客户端：

| 字段 | 填写内容 |
| --- | --- |
| Client ID | `chdxk`（也可自定，与 CHDXK 配置一致） |
| 名称 | 频道课程榜 |
| 回调地址 | `https://courses.example.com/api/auth/oidc/callback` |
| 登出返回地址 | `https://courses.example.com/` |
| 授权范围 | `openid profile res_hub_identity points.read points.write` |
| 授权类型 | `authorization_code client_credentials` |

点击「创建并显示一次密钥」，立即保存密钥，后续填入 `RESHUB_CLIENT_SECRET`。客户端保持启用。回调和登出地址必须与部署地址精确一致，包含路径与末尾斜杠；生产使用 HTTPS。

资料站若已经支持其他应用的 OIDC 登录，通常只需增加客户端。若尚未开启，由技术管理员检查资料站后端环境配置：

- `OIDC_ENABLED=true`。
- `OIDC_ISSUER` 为资料站实际对外 OIDC 地址，必须与发现文档及令牌中的 issuer 一致。
- `OIDC_SIGNING_KEY_PATH` 指向持久保存的签名私钥；不能每次重启重新生成。
- 生产环境按资料站策略启用 `OIDC_REQUIRE_HTTPS=true`。
- 网关将 `/.well-known/openid-configuration`、`/oidc/`、`/api/integrations/v1/` 转发到资料站后端，不能由前端 SPA 返回 HTML。

浏览器访问 `<OIDC_ISSUER>/.well-known/openid-configuration` 应得到 JSON。服务器及 API、worker 容器也必须能访问该地址。发行版不会自动修改资料站、服务器网关或系统网络配置。

若资料站已启用 `OIDC_MTLS_REQUIRED`，保留该策略，并按本文第 5 节配置客户端证书；不要为接入 CHDXK 关闭已有保护。

## 2. 获取源码与填写配置

服务器需要已安装并运行的 Docker Engine 与 Compose 插件，以及可用的 HTTPS 反向代理。

```bash
git clone --branch v0.1.0 https://gitee.com/techchu/CHDXK.git chdxk
cd chdxk/docker/release
cp .env.example .env
```

编辑 `.env`：

```dotenv
CHDXK_VERSION=0.1.0
CHDXK_PUBLIC_URL=https://courses.example.com
CHDXK_BIND_ADDRESS=127.0.0.1
CHDXK_PORT=8088
SESSION_SECURE=true
POSTGRES_PASSWORD=填写随机数据库密码
SESSION_SECRET=填写至少32位随机字符串
RESHUB_ISSUER=https://res.example.com
RESHUB_CLIENT_ID=chdxk
RESHUB_CLIENT_SECRET=填写资料站刚生成的客户端密钥
```

数据库密码使用随机字母和数字，避免空格影响连接字符串。可以执行 `openssl rand -hex 32` 生成随机值，数据库密码与会话密钥分别生成。`.env` 不提交到 Git，不放到可公开下载的目录。

`RESHUB_ISSUER` 指向后端对外 OIDC 地址：如果资料站前后端共用域名，就填该域名；如果分开，则填能提供发现文档的后端地址，不能填仅提供资料站前端的地址。

## 3. 构建并启动

以下命令均在 `docker/release` 下执行：

```bash
docker compose --env-file .env -f compose.yaml build api taskworker frontend
docker compose --env-file .env -f compose.yaml up -d
docker compose --env-file .env -f compose.yaml ps
```

共启动 PostgreSQL、Redis、API、异步 worker、前端五项服务。首次空数据库启动时，按编号执行 `backend/script` 下全部 SQL up 文件。数据库与 Redis 使用命名数据卷；重建应用镜像不会清空数据。此初始化流程用于新安装，不是未来版本的增量升级工具。

PostgreSQL 使用带 `pg_jieba` 的镜像 `dujiajun/postgres-pg-jieba:17`。若镜像无法拉取，可在项目已有 Docker 构建环境内执行 `docker compose --env-file .env -f compose.yaml build postgres`。若 Docker、网络、系统环境本身存在问题，应由服务器技术管理员处理。

API 镜像已包含教务采集器、Node 和 Chromium，无需宿主机的 CHUAuthSDK 或固定目录。worker 必须运行，点评积分奖励等异步任务依赖它。

将已有 HTTPS 网关的课程榜域名全部请求转发到 `http://127.0.0.1:8088`，保留 `Host` 并设置 `X-Forwarded-Proto: https`。容器前端自动转发 `/api/` 到 API；前端路由由 SPA 处理。仅前端端口暴露到宿主机，数据库与 Redis 不对外映射。

若反向代理也是容器，按部署环境调整网络和绑定地址，不能在代理容器内把 `127.0.0.1` 当作宿主机。只有本地 HTTP 调试才设置 `SESSION_SECURE=false`；线上 HTTPS 使用 `true`。

## 4. 首次登录、绑定与验收

```bash
docker compose --env-file .env -f compose.yaml logs api
```

首次创建数据库后，API 日志输出 ID=1 初始超级管理员的 `username`、`password`。保存首次凭证，打开课程榜登录页，点击较小的「使用本站账号登录」入口，输入日志账密。第一次登录必须修改密码。普通重启不会重置密码，也不会再次输出凭证。

进入个人设置可绑定资料站账号。未绑定的独立账号可以使用站点和对应管理功能，但不能使用资料站积分。积分由资料站保存；多个本站账号绑定同一资料站账号时共享该资料站余额。

资料站授权登录首次使用会自动创建并绑定本站账号；资料站管理员在本站继承管理员身份。一个资料站账号绑定多个本站账号时，登录显示账号选择。初始超级管理员保留本站身份与账密入口。本站的封禁、禁言仅作用于本站。

验收建议：

1. 首页、课程页正常，`/api/auth/options` 返回 JSON，匿名访问受保护接口返回未登录响应。
2. 初始账密登录、强制改密、超级管理员后台可用。
3. 使用普通资料站账号登录，检查昵称与个人资料、绑定关系；使用资料站管理员检查后台权限。
4. 绑定后积分页面显示余额；由有权限的管理员小额发放并收回积分，确认资料站流水；发表测试点评，检查 worker 奖励任务结果。
5. 教务扫码、教师资料同步需能访问学校网站；扫码需要实际用户操作，不能用镜像构建成功代替采集验收。

故障查看 `docker compose --env-file .env -f compose.yaml logs --tail 100 api taskworker frontend`。不要把包含初始密码、令牌、cookie 的完整日志公开。

## 5. 已启用 mTLS 的资料站

由资料站技术管理员按已有证书签发流程提供客户端证书、私钥与信任 CA，在客户端登记其 SHA-256 指纹。资料站网关校验客户端证书，并按资料站已配置的可信代理机制传递证书指纹；`OIDC_TRUSTED_PROXY_IPS` 必须与真实网关一致。

把证书放在 `docker/release/certs/`，新建 `compose.mtls.yaml`：

```yaml
x-mtls-env: &mtls-env
  JCOURSE_RESHUB_CLIENT_CERT_FILE: /app/certs/client.crt
  JCOURSE_RESHUB_CLIENT_KEY_FILE: /app/certs/client.key
  JCOURSE_RESHUB_CA_FILE: /app/certs/ca.crt
services:
  api:
    environment: *mtls-env
    volumes:
      - ./certs:/app/certs:ro
  taskworker:
    environment: *mtls-env
    volumes:
      - ./certs:/app/certs:ro
```

两个容器的运行用户都需要读取证书与私钥。启动时追加覆盖文件：

```bash
docker compose --env-file .env -f compose.yaml -f compose.mtls.yaml up -d
```

后续查看状态、日志和停止服务也使用相同的两个 `-f` 参数。

## 6. 日常维护

```bash
# 查看状态与 worker 日志
docker compose --env-file .env -f compose.yaml ps
docker compose --env-file .env -f compose.yaml logs --tail 100 taskworker
# 备份 PostgreSQL（在 Linux shell 中执行）
docker compose --env-file .env -f compose.yaml exec -T postgres pg_dump -U chdxk chdxk > chdxk-backup.sql
# 停止，保留数据卷
docker compose --env-file .env -f compose.yaml down
```

同时安全备份 `.env`、证书和数据卷。不要对正式环境执行 `down -v`，它会删除数据卷。后台可关闭注册、维护用户、开课记录、教师资料、公告与内容页面、主题和常用文案；不同管理员按下放权限显示可用功能。
