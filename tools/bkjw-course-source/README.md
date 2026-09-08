# 长大教务全校开课数据源

这个独立工具从 `http://bkjw.chd.edu.cn/eams/stdSyllabus` 读取全校开课查询结果，按学期保存为本地 JSON。它不依赖当前学期的固定 ID，也不修改 jcourse 的数据库或 CSV 导入器。

## 安装

```powershell
pnpm install
pnpm exec playwright install chromium
```

如果本机已有兼容的 Chromium，也可以将 Playwright 的浏览器安装目录配置为已有浏览器；不要把浏览器配置目录提交到 Git。

## 首次登录

```powershell
pnpm start login
```

命令会打开可见浏览器，并在终端提示输入学工号和密码，自动填写长大统一身份认证表单，同时勾选页面的“7天免登录”选项。程序会调用统一认证的验证码检查接口；需要验证码时会使用与 CHUAuthSDK 相同的 `getCaptcha.htl` 图片接口，在浏览器中显示验证码并打开一个临时图片副本，用户手动输入 4 位字母或数字后继续提交。账号、密码和验证码只在本次进程内使用，不会打印、写入参数或导出 Cookie。若页面结构无法识别，可在浏览器中手动完成登录，再回终端按回车。

登录完成后程序会新开一个教务入口页验证会话是否真正可复用；只有验证成功才会打印“登录状态已验证并保存”。由于 Chromium 默认会在关闭进程时丢弃会话级 Cookie，程序会在内存中将长大教务和统一认证域名的会话项转为 7 天的本地浏览器配置项，然后关闭浏览器；不会导出、打印或写入 Cookie 值。会话保存在 `data/bkjw/browser-profile`，采集和详情请求始终只允许 `bkjw.chd.edu.cn/eams/`。

## 常用命令

```powershell
# 查看站点当前可发现的学期（规范化名称、站点 ID、原始名称）
pnpm start semesters

# 采集一个学期，可使用 2025-2026-1、原始中文名称或站点 ID
pnpm start fetch --semester 2025-2026-1

# 采集所有可发现学期
pnpm start fetch --all

# 从上次中断的位置继续；详情失败项会重新尝试
pnpm start fetch --semester 2025-2026-1 --resume
```

可选参数：`--output <目录>`（默认 `data/bkjw`）、`--profile <目录>`、`--headless`（采集默认可见）、`--page-size <数量>`（默认 1000）、`--detail-concurrency <数量>`（默认 3，最多 8）。页大小会从站点实际提供的选项中选择不超过请求值的最大值；站点新增更大选项时无需修改程序。

详情页使用同一 Playwright 登录会话的服务端 HTML 请求，避免为每条课程反复启动浏览器页面；仍保留有限并发、请求间隔、超时重试和断点保存。

## 输出

每个学期目录包含：

* `semester.json`：列表字段和详情页字段组成的课程源。`lessons` 以站点 `lesson.id` 为稳定键，保留课程摘要、基本信息、排课信息、限制条件、链接和解析警告。
* `checkpoint.json`：列表进度、已完成详情 ID、累计重试次数和失败项。
* `report.json`：完成状态、数量一致性、重复 ID、缺失详情和页面结构签名。

文件采用临时文件写入后替换。中断或详情失败时仍会保存当前数据，但 `report.json.completed` 为 `false`，必须使用 `--resume` 完成后才会标记为 `true`。课程教师只保存网页展示的姓名，不推断不存在的教师工号、职称或主讲关系。

JSON 文件使用稳定的 snake_case 键名（例如 `schema_version`、`fetched_at`、`expected_count`、`completed_lesson_ids`）；程序读取断点时也兼容此前的 camelCase 文件。

## 开发检查

```powershell
pnpm typecheck
pnpm lint
pnpm test
pnpm build
```

测试使用脱敏 HTML fixture 覆盖学期日历、分页列表、详情字段、限制条件和断点恢复；在线运行仍需要用户自己的教务登录会话。
