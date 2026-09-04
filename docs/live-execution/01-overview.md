# 01 · 产品契约与锁定决策

## 1. 要解决什么

自动化测试平台是 **观察者**：CI / 本地调用 `pytest --enable-hermes-plugin`，平台不拉代码、不起 agent、不触发 Jenkins。

用户要能：

1. 在 Web 上看到 **当前** 测试运行状态（以插件上报为准，没有「排队等待启动」）。
2. 看到 **历史** 运行统计（沿用现有 Go dashboard 快照：摘要、近 7 天趋势、running 数）。
3. **点开一次正在跑的 job/execution** 后，会话页用卡片展示当前 case，**精确到步骤**。

## 2. 实时契约（不可再议）

| 项 | 约定 |
| --- | --- |
| 新鲜度 | 步骤 **进入 / 退出** 约 **1 秒** 内出现在已打开的会话页 |
| 不是 | 步骤内部日志流、截图流、毫秒级推送 |
| 会话身份 | 一次 pytest = 一次 `TestExecution`，键为 **`build_uid`** |
| Job | 目录。若该 job **恰好一个** running execution → 一键打开；多个 → 先选 execution |
| 当前 case | v1 **一张** 主卡片（禁止 xdist）。已结束 case 是侧栏列表，不是 live 卡片 |
| 步骤深度 | 路径段数 **≤ 5**（`0` … `0.1.2.3.4`）。第 6 层在 **测试进程失败**，后端拒写 |
| 并行 | v1 检测到 pytest-xdist **直接失败退出**，提示去掉 `-n` |
| 附件 | v1 Hermes **不存、不展示** 附件内容；截图/日志只在 Allure 报告 |
| 报告 | Allure 由 CI 本地 `--alluredir` + `allure generate` 生成；服务端不导出 Allure zip |

## 3. 埋点

- 用户 API：**Allure**（`allure.step` / `allure.attach` / `@allure.step`）。
- `hermes_plugin.step` / `attachment` 仅为 **别名**，不得再维护第二棵步骤树。
- 插件 **硬依赖 `allure-pytest`**。未启用 Allure lifecycle 则无法上报步骤。
- Hermes 订阅 Allure lifecycle：`start_step` → 上报 `running`；`stop_step` → 上报终态。

## 4. 读模型

| 路径 | 数据源 | 刷新 |
| --- | --- | --- |
| Live 会话页 | automation **写侧**（execution / item / step） | REST 全量 + SSE delta |
| Dashboard 统计 | dashboard **快照表** | asynq，`Unique(5m)` 可保留 |

Live **禁止** 走 Unique(5m) 投影。Step 事件 **不** 触发全量 dashboard 重算；item/execution 终态仍可发既有 `ItemUpdated` / `ExecutionUpdated`。

## 5. 存活定义（观察者）

Running = 插件已 `POST /executions` 且尚未进入终止态。

进程级心跳 **10–15s**；scheduler 发现 **60s** 无心跳 → `aborted`，SSE 推终态。**不用**「最后一步的时间」判断存活。Web 人工中止为后续补充，不进 v1。

## 6. 技术边界

- **只做 Go。** Web 自动化与 dashboard 打 `/tap/api/v1`。
- SSE 客户端用 **`fetch` + `ReadableStream`** 带 Bearer；不用原生 `EventSource`（无法带头）。
- 断线：先 REST 全量，再订 SSE；不依赖 Last-Event-ID。
- Fan-out：**Redis Pub/Sub** 频道按 `build_uid`，不走 asynq。
- 视觉：后台壳体保留；会话页单独卡片语法（见 [05-web.md](./05-web.md)）。running 与 history **同一路由**，只差是否订 SSE。

## 7. 明确不做（本期）

- 调度 pytest / 触发 Jenkins / agent 池
- pytest-xdist
- 对象存储、附件二进制、会话页附件标记
- 服务端生成 Allure zip
- 插件 CLI 改名（保留 `--tap-*` / `--jenkins-*`）
- 删除 `service/` 目录
- 修复既有 `is_latest` 不翻转等已知缺陷（除非挡住 live「当前 case」）
