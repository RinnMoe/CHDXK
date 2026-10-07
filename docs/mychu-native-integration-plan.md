# MyCHU 原生课程榜接入规划

状态：原生授权桥接已在两端工作分支实现并完成本地 HTTPS 联调；基线为 CHDXK v0.1.0（372b8fc），发布版本未修改。MyCHU 协作任务：`01a1162a-00d9-79b0-b003-b3a2ff557126`，实现工作树 `C:/Users/Rinn/.codex/worktrees/channel-course-rank/MyCHU`。准确协议与验证范围见 [原生授权接口](mychu-native-authorization.md)。

配套文档：[MyCHU 原生接入规划](C:/Users/Rinn/.codex/worktrees/channel-course-rank/MyCHU/doc/plans/chdxk_native_integration.md)。双方已核对授权方式、接口与验收边界；授权码 60 秒一次性、账号选择 300 秒、宿主内存 Cookie jar。

## 产品范围

在 MyCHU 注册原生课程榜功能，使用 Dart 模型、服务和页面；登录、凭据和会话由宿主能力提供。课程榜业务不以 WebView 包装。界面名称、图标遵循 MyCHU 已有应用注册方式，建议显示「频道课程榜」。

推荐一期：频道授权与本站账号选择、课程搜索筛选、课程详情和教学班、教师检索与精简详情及教师主页链接、课程点评及统计、发布/编辑/删除自己的点评、有用/无用票、课程和教师热度榜、关注/屏蔽、我的课程与手动选择已有教学班。积分复用宿主已有资料站展示，点评页面解释奖励由服务端处理；不另外建立积分账户。

独立账密注册、绑定管理、内容运营、站点配置及其他后台操作继续在 CHDXK 网站维护。MyCHU 用户拥有管理员身份也不会因此获得新的原生后台页面。推荐一期范围不是删除网页功能。

第二阶段：MyCHU 已有课表导入本站。当前网站只有预留入口，尚不存在可用的 MyCHU 批量导入协议。教务扫码导入是另一条既有路径，不等同于 MyCHU 课表同步。

## 已核对的实际约束

- MyCHU `ChannelAccountCapability` 包装 `ChannelResourcesSessionService`，访问既有资料站同源路径；内存 access token、安全存储 refresh cookie、登录刷新与退出均由宿主管理，功能页面不拿凭据。
- CHDXK 业务路由位于登录校验之后，课程、教师和点评读取也要求登录。公开配置接口不能代表业务 API 可匿名调用。
- CHDXK 用户使用 Redis Cookie 会话，写操作使用 `X-CSRF-Token`；现有 Authorization Bearer 分支解析系统 API Key，不是原生用户令牌。
- ResHub OIDC 发现文档当前提供 authorization_code/client_credentials 与 client_secret_basic。Flutter 不能嵌入 CHDXK confidential client secret，也不能把资料站现有 Bearer 发到 CHDXK 冒充本站会话。
- CHDXK 账号与资料站身份是绑定关系，必须按 issuer + subject 对齐。多绑定账号需要选择；本站封禁、禁言独立；资料站管理员继承身份不依赖其前端权限模式。
- 积分在 ResHub 保存，未绑定账号不可用。发布点评成功与 worker 奖励到账是两个状态，客户端不能立即累加余额。

## 授权流程

1. MyCHU 宿主恢复已有频道会话，未登录则展示已有频道登录页面。由宿主建立 CHDXK 授权请求及随机 PKCE verifier/challenge，不由功能页面处理秘密。
2. ResHub 校验当前频道会话，签发短时、单次、限定 CHDXK 客户端并绑定 S256 challenge 的授权票据。ResHub 采用已有 Bearer，CHDXK 写接口使用自身 Cookie/CSRF；接口与 DTO 见原生授权文档。
3. 宿主仅向已配置的 CHDXK origin 提交票据与 verifier。CHDXK 后端通过服务端凭据向 ResHub 兑码，核对受众、有效期和请求绑定，获得与网页 OIDC 一致的身份及后续验证/撤销关联。不能仅依赖一次登录时的静态身份 JSON 永久信任会话。
4. CHDXK 复用已有创建/绑定账号规则：零绑定时自动创建并绑定新本站账号，单绑定可直接登录；多绑定进入 pending 状态，复用账号列表和选择接口。pending 不能调用业务接口，也不能选择当前资料站身份以外的账号。
5. 由 CHDXK 建立现有 Redis 用户 Cookie 会话。MyCHU 宿主持有专用 CHDXK cookie jar 和 CSRF，同源范围与资料站严格分开，随后获取 `/api/auth/me`。会话状态包含需要授权、需要选账号、已登录、失效及服务暂不可用。
6. 退出课程榜清除其会话、pending 与缓存；切换或退出频道账号同时清除 CHDXK 会话，隔离旧异步请求结果。未获用户指令不替用户退出网页端的独立会话。

推荐 Cookie + CSRF 是为了复用本站账号与权限机制；若改用原生 Bearer，必须新增明确用户会话验证通道、续期及撤销协议，不能使用现有 API Key 分支。若 ResHub 桥接暂无法交付，可另评估系统浏览器 OIDC、CHDXK 后端回跳单次兑换码；业务界面仍保持原生，该备用路线不视为已有实现。

## 可复用接口

以 `backend/internal/interface/web/router.go`、controller 与 `frontend/src/api/*.ts` 为实际契约来源；以下路径均相对于站点根地址。

| 功能 | 接口/约束 |
| --- | --- |
| 用户与 CSRF | GET `/api/auth/csrf`；GET `/api/auth/me`；POST `/api/auth/logout` |
| 多绑定选择 | GET `/api/auth/accounts` → AccountDTO[]；POST `/api/auth/accounts/select`，请求 `{id}` → AccountDTO |
| 课程 | GET `/api/course/filter`、`/api/course/`、`/api/course/:id` |
| 教师 | GET `/api/teacher/filter`、`/api/teacher/`、`/api/teacher/:id`、`/api/teacher/:id/course` |
| 热度 | GET `/api/course/hot`、`/api/teacher/hot`，使用服务端 week/month、period_key，不计算教学周 |
| 点评和统计 | GET `/api/course/:id/review`、`/review/filter`、`/review/trend`；GET `/api/review/:id` |
| 点评写入 | POST `/api/review/`：course_id、semester、rating、content、可选 score；PUT/DELETE `/api/review/:id` |
| 投票 | POST `/api/review/:id/vote`，vote_type=1/-1/0 |
| 关注和屏蔽 | POST `/api/course/:id/notification`，level=0 普通、1 关注、2 屏蔽 |
| 我的课程 | GET `/api/course/enrolled`、`/followed`、`/ignored`；POST `/api/course/enrollment`、PUT/DELETE `/api/course/enrollment/:id` |
| 手动选课 | 保存请求为 section_id，必须选择本站已有课程及教学班，学期与课程由后端验证 |
| 积分 | GET `/api/user/:userID/point`，本人或有权限管理员；结合 has_points_account 判断 |
| 站点内容 | site-copy/site-pages 中的品牌、介绍等常用文案可复用，通用按钮固定；网页主题不强制覆盖 MyCHU 宿主主题 |

分页响应为 `{items,total,page,page_size}`。列表字段可空、评分分布、学期字符串和 Markdown 使用类型模型；不显示教师程序化编码。教师列表隐藏缺失项，详情允许显示暂无数据。

既有业务错误可能为 `{error}`；原生兑换、CSRF、强制改密、封禁和禁言已补稳定 code。客户端区分未登录、待选账号、本站封禁/禁言、权限不足、服务暂不可用及网络错误；不得按中文错误文案匹配业务。安全读取最多恢复后重试一次，写入不自动重放，避免重复点评、投票或积分动作。

## 实施分工与顺序

| 阶段 | ResHub / CHDXK | MyCHU |
| --- | --- | --- |
| 0 契约冻结 | 桥接票据与兑码、身份/撤销、pending、错误 code、环境地址及样例 | 宿主会话能力接口、原生页面范围、模型与fixture |
| 1 授权闭环 | 实现桥接、复用本站账号选择与 Cookie/CSRF、本站处罚验证 | 已有频道登录复用、专用 cookie jar、账号选择、退出与账号切换 |
| 2 浏览 | 稳定课程/教师/点评/热度契约，不复制业务服务 | 应用注册、typed服务、检索/详情/榜单与空态错误态 |
| 3 写入 | 点评权限、投票、关注与手动选课、异步积分语义 | 原生 Markdown 点评、提交确认/草稿、关注与我的课程 |
| 4 课表导入 | 新增批量匹配预览/确认契约与幂等校验 | 从宿主课表能力读取，用户确认后导入 |

课表匹配优先学期与 source_lesson_id/lesson_no；课程码、教师及学院用于辅助与消歧。不凭课程名称自动猜测，不把未匹配课程新建到教学数据。导入必须展示匹配、重复、歧义和未匹配结果，不删除用户已有选课记录。

## 联调验收

| 场景 | 必须确认的结果 |
| --- | --- |
| 未登录/已有频道登录 | 一套频道登录体验；第二服务会话单独建立，不泄露凭据 |
| 首次使用/单绑定/多绑定 | 同一 issuer+subject；自动创建或登录正确账号；多绑定明确选择 |
| 伪造或重放授权 | 过期、重复、错误客户端、错误 verifier、其他身份 account_id 全部拒绝 |
| 退出/切换/重启 | 专用会话和pending清除、旧请求不串号；持久保存策略由宿主统一决定 |
| 权限和处罚 | 管理员身份继承；本站封禁不可登录、禁言不可发点评；资料站账号不受本站处罚 |
| CSRF/服务异常 | 写入校验通过；缺失token拒绝；401/403/503与网络错误可区分且不重复写入 |
| 点评/投票/关注/选课 | 网站与原生数据一致，教学班约束和等级枚举正确 |
| 积分 | 共用资料站余额，点评异步到账可刷新；多绑定账号不生成多个余额 |
| 真实客户端 | 静态检查、模型/会话测试、两站Docker联调、Android/iOS真机或明确可用目标验收分别记录 |

不能将源码检查或 mock 测试宣称为已完成设备体验或生产部署验证。授权协议已在工作分支和隔离联调实例可调用；课表批量导入仍为后续规划。
