# MyCHU 原生授权接口

本协议已在 CHDXK 和 ResHub 工作分支实现。MyCHU 复用宿主资料站登录，CHDXK 保留独立 Redis 用户会话；不接受资料站 Bearer 替代 CHDXK Cookie。

## 部署

ResHub 超级管理员登记 CHDXK OIDC 客户端，技术管理员设置 `OIDC_NATIVE_CLIENT_IDS` 为该 client_id。CHDXK 原有 `reshub.issuer/client_id/client_secret` 配置继续使用，客户端秘密留在服务端。生产 `session.secure=true`，HTTPS 反向代理保留 `/api/` 路由和 `Set-Cookie`、`X-CSRF-Token`。MyCHU 分别配置资料站 origin、CHDXK origin 和同一个 client_id。

## 交互顺序

1. 宿主创建随机 state（16–128 字符 base64url）、PKCE verifier（43–128 字符），向 ResHub `POST /api/auth/native/authorize` 提交 client_id、S256 challenge、method=S256、state，使用已有资料站 Bearer。取得 `{code,state,expires_in:60}` 并检查 state。
2. 使用专用内存 Cookie jar 请求 CHDXK `GET /api/auth/csrf`（204），读取 `X-CSRF-Token` 和 `jcourse_session` Cookie。Cookie Path 为 `/`，HttpOnly，生产 Secure。
3. `POST /api/auth/native/exchange`，携带 Cookie 和 CSRF header，请求 `{code,code_verifier,state}`。
4. 单账号：200 `{status:"authenticated",account:AccountDTO}`。首次绑定身份自动创建本站账号；本地封禁返回 403。
5. 多账号：200 `{status:"account_selection_required",expires_in:300}`，此时没有本站登录身份，不能访问业务接口。刷新 CSRF 后，`GET /api/auth/accounts` 取得 AccountDTO 数组，再 `POST /api/auth/accounts/select` 请求 `{id}`，成功返回 AccountDTO。5 分钟后选择失效；只能选择该身份已绑定、未封禁的账号。
6. 兑换和账号选择都会轮换会话，必须重新请求 CSRF，再读取 `/api/auth/me`。不要复用轮换前 CSRF token。
7. 取消或退出使用 `POST /api/auth/logout`，宿主也清除专用 Cookie、CSRF、pending。用户切换资料站账号时不沿用原 CHDXK 会话。

CHDXK 服务端兑码使用固定内部 `urn:reshub:native`，复用 `/oidc/token` 和已有 ID Token 校验，state 与 nonce 对齐。网页授权使用真实 callback，两种码不能混兑。主站 auth_session 撤销后，CHDXK 下一次业务请求经实时 userinfo 校验返回 401 并清除会话；Redis Cookie 本身期限较长不代表主站失效后仍能使用。

## 错误与业务

原生兑换响应为 `{code,error}`：`invalid_request` 400、`invalid_grant` 401、`account_suspended` 403、`provider_unavailable` 503、`internal_error` 500。通用写请求 CSRF 错误为 `csrf_invalid`；强制改密为 `password_change_required`；本站封禁为 `account_suspended`；禁言为 `muted`。其他既有业务错误仍可能只有 `{error}`，不要根据中文文本推断业务类型。

课程和教师列表使用 `/api/course/`、`/api/teacher/`，需保留末尾斜杠。通知等级 0 普通、1 关注、2 屏蔽。手动选课提交 section_id，必须是本站已有教学班。积分仍在 ResHub；未绑定账号没有积分账户，点评奖励由 worker 异步处理。MyCHU 原生界面一期不包含后台或课表批量导入。

## 验证

2026-10-07：相关 Go 应用、控制器、中间件和 ResHub 客户端测试及 vet 通过；新增测试覆盖 CSRF、nonce、固定兑码标记、单/多账号、待选择不可访问业务、选择超时、封禁、主站撤销。隔离 PostgreSQL/Redis 实例加项目内临时 CA 完成真实 HTTPS ResHub 登录→授权→CHDXK 兑换→me/课程→主站退出后 401，且 ResHub 管理员处于普通模式仍继承管理员身份。

MyCHU agent 通过 `test/integration/channel_course_live_test.dart` 完成无 mock 的 Flutter HTTPS 请求联调：宿主登录/授权、Cookie/me、课程/教师/filter/hot/我的点评/选课/站点内容 DTO、课程趋势、点评创建更新读取删除、投票 1/-1/0、关注恢复、选课创建修改删除及两站退出。测试写入已清理。这是客户端真实传输验证，不等同于手机设备界面验收。

生产部署、生产 mTLS 和 Android/iOS 设备体验需单独验收。临时证书与联调密码只在忽略目录保存，不纳入版本控制；v0.1.0 发布保持原样。

## 食堂评价

食堂评价复用同一套 native exchange 与本站会话，不需要新增客户端登录方式。接口契约见 [`canteen-api.md`](canteen-api.md)；当前实现只在本机验证，原生端接入时按该文档读取 `/api/canteen/*` 即可，不需要绑定 Flutter 实现或地图数据。
