# 小内存 Docker 部署

目标是低负载常驻占用尽量控制在约 0.5 GB，CHDXK 五个容器的内存硬上限合计 928 MiB（约 973 MB），低于 1 GB。预算不包含宿主 Linux、Docker daemon、外部反向代理、ResHub 或源码构建。

## 默认预算

| 容器 | 内存上限 |
| --- | ---: |
| API（含 Node、Chromium） | 560 MiB |
| PostgreSQL | 192 MiB |
| worker | 96 MiB |
| Redis | 48 MiB |
| 前端 Nginx | 32 MiB |
| 合计 | 928 MiB |

Compose 同时设置 mem_limit 和相同的 memswap_limit，避免靠交换空间掩盖进程超额。API 和 worker 的 Go 使用软预算 48 MiB、单核执行；Node 堆预算 96 MiB。这些软预算不能替代容器硬限制。达到硬限制时，进程可能被 OOM 终止，任务会失败，不能保证任何数据量都能在此额度内完成。

## 排队与任务

管理员开课同步和个人课程导入共用一个先进先出队列，一次只运行一个浏览器任务；队列最多 32 项（含正在运行的一项）。排队中没有浏览器、二维码或学校登录态，页面显示前方任务数，可以取消。轮到后才生成二维码，请此时扫码。

排队最多 30 分钟，过期需重新提交；个人任务运行最多 15 分钟。连接后前 5 分钟仍未开始同步会释放连接，避免无人扫码或不选择学期占用队列；管理员已开始的批量同步保留原有长任务期限。结束、取消、失败或超时均关闭连接；在旧进程退出前不会运行下一个。Linux 清理采集进程组，防止遗留 Chromium。重启不会恢复排队或学校登录态，需要重新提交。

课程详情采集和 worker 并发均为 1。大量任务会延迟，不通过增加浏览器并发提速。该部署面向单 API 实例，多实例不保证全站单浏览器。

## 数据库和缓存

PostgreSQL shared_buffers=32MB、work_mem=1MB、maintenance_work_mem=16MB、max_connections=20，关闭单查询并行 worker。每个应用进程的连接池最多 6 条、空闲 2 条。

Redis maxmemory=24MB、noeviction，不淘汰会话与任务，满额后写入报错，需要处理积压或增加资源。AOF 保留，关闭定时 RDB 快照以降低常规 fork 频率；仍可能发生 AOF 重写。

## 部署与观察

使用包含本次更改的源码构建；原 v0.1.0 标签不会自动获得这些设置。在 docker/release 下执行：

```bash
docker compose --env-file .env -f compose.yaml build api taskworker frontend
docker compose --env-file .env -f compose.yaml up -d
docker compose --env-file .env -f compose.yaml stats
```

构建阶段不受运行时额度限制，小内存服务器可在其他机器构建镜像后 docker save/load，或使用镜像仓库分发。

观测时区分空闲、扫码等待、个人全学期导入和管理员批量同步，并检查 OOMKilled 和重启次数。约 0.5 GB 是目标，928 MiB 是容器额度总和，不是所有真实任务峰值均低于 0.5 GB 的承诺。

## 本地验证记录（2026-10-08）

在隔离的 Linux Docker 测试栈中，五个容器空闲合计约 56 MiB；生成真实学校登录二维码、验证个人任务排队与取消，再运行 Chromium 渲染 5000 行模拟页面，约两秒间隔采样得到合计峰值约 206 MiB。实际容器额度合计 928 MiB，未观察到 OOM 或重启。这是空数据库与模拟页面测试，尚未验证真实扫码后的全学期个人导入、完整开课同步或多用户网站负载，不能作为生产峰值保证。

API 正常 Docker 构建因 Debian 依赖下载返回 HTTP 500 中断；本次运行验证使用现有 v0.1.0 运行环境，替换为最新 Linux API 二进制和采集器产物。worker 与前端使用当前源码正常构建。完整 API 镜像构建仍需在依赖下载恢复后验证。
