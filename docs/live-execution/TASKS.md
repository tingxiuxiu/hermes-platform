# 开发任务与切片路线

按垂直切片推进：每一刀用户可感知的行为变完整一块，而不是先铺完所有层再接线。

测试接缝写在切片里。未列出的内部函数不要单测。

---

## 路线总览

```
S1 领域深度 5
 → S2 心跳列 + 心跳 API + 超时 aborted
 → S3 步骤 ingest 发 live 事件形状（进程内 hub，测试可注入）
 → S4 case_uid 步骤 API + 深度 422 + 忽略附件
 → S5 Redis Pub/Sub + SSE HTTP
 → S6 GET /live 快照 + 修正 GET execution
 → S7 插件 Allure + xdist 失败 + 步骤 enter/exit + 心跳
 → S8 Web 契约对齐 + 会话页 + SSE 客户端
 → S9 Dashboard/Pipeline 带 build_uid + Jobs 入口
 → S10 文档/代理/契约 diff allowlist 收尾
```

S1–S6 可在无真实 pytest 的情况下用 Go 集成测试打通「上报步骤 → SSE」。S7 接真实插件。S8 才是「点开卡片」。

---

## S1 · 步骤深度上限 5（领域）

**行为：** `ParseStepPath("0.1.2.3.4")` 成功；六段路径失败，错误为 5 层而非 16 层。

**接缝：** `automation.ParseStepPath` / `NewStep`（`go-service/internal/domain/automation`）。

**验收：** 现有 `TestParseStepPathValid` 增加五段合法例；非法集包含 `0.1.2.3.4.5`；`ErrStepPathTooDeep` 文案含 `5`。

**依赖：** 无。

---

## S2 · 心跳与僵尸 aborted

**行为：** running execution 可心跳续命；超过 60s 无心跳被标 `aborted`，与插件 `Finish` 同一领域规则。

**接缝：**

- HTTP：`POST /automation/executions/{build_uid}/heartbeat`（集成测试 + 真实 PG）
- 应用：扫描 stale 并 `Finish(aborted)` 的用例函数（集成测试用插入过期行，不 mock 时钟则可注入 Clock）

**工作：**

- 迁移 `000004_execution_heartbeat.up.sql`：`last_heartbeat_at`
- 创建 execution 时写入 now；仓储 UpdateHeartbeat / ListStaleRunning
- 配置 `HEARTBEAT_TIMEOUT_SECONDS=60`
- scheduler 注册 `automation:abort-stale`；worker 消费
- 心跳不发 SSE、不发 asynq

**验收：** 集成测试 create → heartbeat 200；把 `last_heartbeat_at` 打到超时前 → abort 后 GET 为 `aborted`。

---

## S3 · LivePublisher 端口 + StepUpserted

**行为：** `IngestStep` 成功后调用 `LivePublisher`（测试用记录型 fake）；不入 asynq Unique。

**接缝：** `IngestUseCase.IngestStep` 对 `LivePublisher` 的可观察调用（通过 fake 的公开记录，不 mock 仓储）。可用现有集成测试 upsert 一步后断言 fake 收到的 `type/build_uid/step_path`。

**依赖：** S1。S2 可并行，但事件里暂不强制带 heartbeat。

---

## S4 · `POST /items/{case_uid}/steps`

**行为：** 服务令牌 upsert 单步；第 6 层 422；attachments 字段忽略。

**接缝：** HTTP 集成测试（与 `automation_test.go` 同风格）。

**依赖：** S1、S3。

---

## S5 · Redis SSE

**行为：** 两进程语义：ingest 后订阅者收到 `step.upserted`。单测 hub 可用 miniredis 或测试 DB15（项目已有真实 Redis，优先真实 Redis）。

**接缝：**

- `GET /automation/executions/{build_uid}/events` + 并行 `POST` 步骤，读 SSE 第一帧
- JWT 鉴权：无 token 401

**注意：** api `WriteTimeout` 对 SSE 连接放行。

**依赖：** S3、S4。

---

## S6 · `GET /live` + 修正 execution GET

**行为：** `/live` 返回 execution + items 摘要 + `current_item` 步骤树；`GET /executions/{build_uid}` 只返回 execution 行；`GET .../items` 分页摘要。

**接缝：** HTTP 集成测试：跑中 case + 两步树。

**依赖：** S4。可与 S5 并行。

---

## S7 · 插件切换 Allure

**行为：** 示例测试用 `allure.step`；Hermes 在 enter 打 running、exit 打终态；6 层失败；xdist 失败；心跳。

**接缝：** `hermes_plugin/tests` + httpx mock / 本地 httptest（Go 未必须起来时用 respx/httpx MockTransport）。

**依赖：** S4、S2 的心跳路径。可先 mock HTTP 再连真实 api。

---

## S8 · Web 会话页

**行为：** `/automation/executions/$buildUid` 展示列表 + 主卡片步骤树；running 时 fetch SSE 应用 delta；前缀 `/tap/api/v1`。

**接缝：** 组件/路由测试优先用 msw 或现有 vitest；有浏览器工具时再手点。

**依赖：** S5、S6。视觉按 05-web.md。

---

## S9 · 入口 build_uid

**行为：** dashboard `running_executions[].build_uid`；Jobs/Dashboard 点进会话页。

**接缝：** dashboard 集成测试 overview JSON 含 uuid。

**依赖：** S6、S8。

---

## S10 · 收尾

- Vite proxy `/tap/api`
- `contract-diff` allowlist 新路由
- 插件 README 按 03-plugin 重写（旧 README 作废）
- `docs/go-service/README.md` 链到 `docs/live-execution/`

---

## 建议开工顺序（本会话）

已开始 **S1**。S1 合入后继续 S2（心跳是观察者正确性，不依赖 SSE）。
