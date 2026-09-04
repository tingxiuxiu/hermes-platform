# 02 · 架构

## 1. 现状缺口（以代码为准，README 可能过期）

| 层 | 现状 | 与契约的冲突 |
| --- | --- | --- |
| 插件 | 自研 `step()`；listener 未调 `create_steps`；步骤默认 `passed`；xdist worker **静默 return** | 没有 Allure 共用；步骤不实时；并行丢数 |
| 插件上报路径 | `POST .../items/{case_uid}/steps` | Go 实际是 `POST .../items/{item_id}/steps`，对不上 |
| Go 写侧 | `(case_uid, step_path)` upsert 已存在；`MaxStepDepth=16` | 深度应为 5；ingest step **不发事件** |
| Dashboard | asynq Unique(5m)；事件只有 execution/item | 不能驱动 1 秒步骤 UI |
| Web | `/hermes-platform/api/v1`；Report Sheet 一次性 REST；状态枚举与 Go 不一致 | 前缀、契约、无 live |
| Dashboard 快照 | running 卡片 **无 `build_uid`** | 无法一键打开会话 |

## 2. 目标数据面

```
CI pytest ──Allure lifecycle──► hermes_plugin ──HTTP+ServiceToken──► go-service api
                                                                      │
                                                                      ├─ PG 写侧 (source of truth)
                                                                      ├─ Redis PUBLISH hermes:live:{build_uid}
                                                                      └─ asynq（仅 item/execution 终态 → 统计快照）

浏览器 ──JWT──► GET /live 全量
       ──JWT fetch SSE──► GET /events  （订阅 Redis）
```

三个二进制职责不变：

- **api**：ingest、live 查询、SSE、写心跳。
- **scheduler**： besides 现有 dashboard refresh，增加 **扫描过期 running → aborted**。
- **worker**：仍只投影 dashboard，不碰 SSE。

## 3. 领域变更

### 3.1 步骤深度

`MaxStepDepth = 5` 表示 **path 段数上限**（与现实现一致，改常数）。

- 合法：`0`、`0.1`、`0.1.2.3.4`（5 段 = 5 级 `with step`）。
- 非法：`0.1.2.3.4.5`。领域错误文案改为「不能超过 5 层」。

`Depth()` 仍为 `len(segments)-1`（根为 0）。插件用 **栈长度** 判断：进入第 6 个 `start_step` 时失败。

### 3.2 Execution 心跳

`test_executions` 增加：

- `last_heartbeat_at TIMESTAMPTZ NULL`

创建 execution 时写入 `now()`。心跳接口只更新该列，不改变 `status`。终止态忽略心跳（幂等 200）。

配置（`AutomationConfig`）：

- `HEARTBEAT_TIMEOUT_SECONDS` 默认 `60`
- 插件间隔 10–15s，不配在服务端强制。

### 3.3 领域事件（live，不进 asynq）

新增事件仅用于 Redis fan-out：

| 事件 | 何时 |
| --- | --- |
| `StepUpserted` | 步骤 upsert 成功 |
| `ItemUpdated` | 已有；live 也订阅 |
| `ExecutionUpdated` | 已有；含 aborted |
| `ExecutionHeartbeat` | **不** 推 SSE（避免 10s 刷屏） |

`bootstrap.EventPublisher`：execution/item **终态** 继续入 asynq；`StepUpserted` **只** 走 Redis。允许 api 进程内同时实现「asynq 桥 + live hub」，automation 用例层仍只依赖 `EventPublisher` 或拆一个 `LiveHub` 端口——推荐 **拆 `LivePublisher` 端口**，避免 asynq 桥误吞 step 事件。

推荐端口：

```go
type LivePublisher interface {
    Publish(ctx context.Context, buildUID uuid.UUID, msg LiveMessage) error
}
```

ingest 在写库成功后调 `LivePublisher`；heartbeat **不调**。

## 4. Redis 频道

- 频道名：`hermes:live:{build_uid}`
- 消息：JSON，`type` + 实体 delta（见 [04-api.md](./04-api.md)）
- api 实例：SSE handler `SUBSCRIBE` 该频道，把 payload 写成 `event: <type>\ndata: <json>\n\n`
- 无订阅者时 PUBLISH 仍成功，允许浪费（步骤频率可接受）

## 5. 心跳超时扫描

scheduler 注册第二条周期任务，例如 `@every 15s`：

- `automation:abort-stale-executions`
- worker **或** scheduler 内直接跑（推荐 **worker 消费**，scheduler 只投递，与现有 refresh:all 一致）
- SQL：`status='running' AND last_heartbeat_at < now() - interval '60 seconds'`
- 对每条走领域 `Finish(aborted)`（与插件结束同一条路径，计回填、pipeline 同步、live 事件、asynq 统计）

并发：`Unique` 短窗口（如 10s）防止扫描重叠。

## 6. 步骤上报身份

插件只有 `case_uid`，没有数字 `item_id`。新增以 **case_uid** 为路径的 upsert，作为 live 主路径。旧 `.../items/{item_id}/steps` 可保留兼容，内部同样走 `IngestStep`。

Live upsert **单步** 为主（enter 一次、exit 一次），禁止再等 case 结束批量。

## 7. 查询：live 快照

`GET .../executions/{build_uid}/live` 读写侧：

- execution（含 `last_heartbeat_at`、`status`）
- 全部 items 摘要（无附件）
- **当前 running item**（v1 至多一条）的 **完整步骤树**
- 选中的已结束 item 不在该接口展开；点选后再 `GET .../items/{case_uid}`（可复用/补齐现有按 case_uid 查询）

列表页、历史统计继续现有 list + dashboard overview。

## 8. 与 dashboard 快照的衔接

为让 Jobs / Dashboard running 卡跳进会话页，快照 DTO **必须带 `build_uid`**。

实现：投影时从 `test_executions.build_uid` 写入快照新列，或 overview 查询 join 源表。推荐 **快照加 `build_uid` 列**，避免 live 导航依赖二次查询。这不把步骤树写入快照。

## 9. 鉴权

| 接口 | 鉴权 |
| --- | --- |
| ingest / heartbeat / steps | `X-Service-Token` |
| live GET / SSE / 列表 / 会话 | 用户 JWT `Authorization: Bearer` |

SSE：`fetch` 带同一 Bearer；`WriteTimeout` 需允许长连接（或该路由不走短 WriteTimeout）。配置 `HTTP_WRITE_TIMEOUT` 若过短会掐 SSE——api 对 SSE 使用独立超时或禁用该连接的 write deadline。

## 10. 失败与背压

插件已有界队列。Live 要求 **步骤事件不可静默丢弃到「用户以为在跑」**：队列满时对 **step enter/exit** 应阻塞短超时或同步重试，而不是 drop。Execution/item 仍可异步。详见 [03-plugin.md](./03-plugin.md)。
