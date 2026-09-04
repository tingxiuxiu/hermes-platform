# 04 · API 契约（Go 为源）

前缀：`/tap/api/v1`。Python 路由表不再作为新接口基准。`contract-diff` 对 **本节新增路径** 使用 allowlist 或改为只回归旧路径。

JSON 时间一律 RFC3339 UTC。列表分页字段保持现有 `rows`（automation 写侧惯例）。

## 1. 上报（`X-Service-Token`）

### 1.1 已有（保持）

- `POST /automation/executions`
- `PATCH /automation/executions/{build_uid}`
- `POST /automation/executions/{build_uid}/items`
- `PATCH /automation/executions/{build_uid}/items/{case_uid}`

### 1.2 心跳（新）

`POST /automation/executions/{build_uid}/heartbeat`

请求体可空 `{}`。

行为：

- execution 不存在 → 404
- `status=running` → 更新 `last_heartbeat_at=now()`，200
- 已终止 → 200 空操作（插件在 sessionfinish 竞态下仍可能打一拍）

不返回大对象，`data` 可 `{ "last_heartbeat_at": "..." }`。

### 1.3 步骤 upsert（新，live 主路径）

`POST /automation/items/{case_uid}/steps`

```json
{
  "step_path": "0.1",
  "step_name": "输入账号",
  "status": "running",
  "start_time": "2026-09-04T14:00:00Z",
  "end_time": null,
  "duration": null
}
```

- `step_path` 段数 > 5 → 422，领域 `ErrStepPathTooDeep`
- `status=running` 时允许无 `end_time`
- 幂等键 `(case_uid, step_path)`，覆盖可变字段（与现 `ApplyUpdate` 一致）
- case 不存在 → 404
- **忽略 `attachments`**（即使客户端误传）
- 成功后 Redis 推 `step.upserted`

旧路径 `POST /automation/executions/{build_uid}/items/{item_id}/steps` 可保留，改为内部转 `IngestStep` 并同样 fan-out；插件新代码只打 case_uid 路径。

## 2. 查询（JWT）

### 2.1 Live 快照（新）

`GET /automation/executions/{build_uid}/live`

```json
{
  "execution": {
    "build_uid": "...",
    "job_name": "...",
    "status": "running",
    "start_time": "...",
    "end_time": null,
    "planned_cases_count": 10,
    "pass_count": 2,
    "failure_count": 0,
    "skipped_count": 0,
    "last_heartbeat_at": "..."
  },
  "items": [
    {
      "case_uid": "...",
      "case_key": "...",
      "case_name": "...",
      "attempt_number": 1,
      "status": "passed",
      "start_time": "...",
      "end_time": "...",
      "duration": 1.2
    }
  ],
  "current_item": {
    "case_uid": "...",
    "case_key": "...",
    "case_name": "...",
    "status": "running",
    "steps": [ { "step_path": "0", "step_name": "...", "status": "running", "children": [] } ]
  }
}
```

`current_item`：`status=running` 的最新 item；没有则为 `null`（全部结束或尚未开始 case）。v1 若出现多条 running（不应发生），取 `start_time` 最新一条并打日志。

`items` **不带** 步骤树，避免全量打爆。点选已结束 case 用：

`GET /automation/items/{case_uid}` → item + 步骤树（补齐若尚无此查询路径）。

### 2.2 SSE（新）

`GET /automation/executions/{build_uid}/events`

- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`
- 鉴权：Bearer（fetch）
- 可选：连接后先不推全量（全量走 REST）
- 事件名与 `data.type` 一致

`data` 示例：

```json
{
  "type": "step.upserted",
  "build_uid": "...",
  "case_uid": "...",
  "step_path": "0.1",
  "step_name": "输入账号",
  "status": "running",
  "start_time": "...",
  "end_time": null,
  "duration": null
}
```

```json
{
  "type": "item.updated",
  "build_uid": "...",
  "case_uid": "...",
  "status": "passed",
  "end_time": "...",
  "duration": 3.2,
  "error_message": ""
}
```

```json
{
  "type": "execution.updated",
  "build_uid": "...",
  "status": "aborted"
}
```

浏览器收到 `execution.updated` 且 `status` 为终止态 → 关 SSE，再拉一次 live 快照。

### 2.3 既有列表对齐 Web

Web 必须改用 Go 字段，不再发明 Python Sheet 那套 `job_uid` / `success`：

| Go | 含义 |
| --- | --- |
| `GET /automation/pipelines` | Job 列表（web 现叫 jobs） |
| `GET /automation/executions` | 过滤 `job_name` / `status` |
| `GET /automation/executions/{build_uid}` | 需修正：今日 handler 只回 items；会话页用 `/live`。列表「打开」走 `/live` 或先 GET execution row |

**修正 GET execution**：现 `GetExecution` 把 items 当 data 返回，丢掉 execution 本体。本期改为：

- `GET /automation/executions/{build_uid}` → `ExecutionRow`（含 heartbeat）
- items 用 `GET /automation/executions/{build_uid}/items`（新，分页，摘要、默认不带步骤）

这是对前端的破坏性对齐，Python 已弃用，允许改。

### 2.4 Dashboard

`GET /automation/dashboard/overview` 的 `running_executions[]` **增加 `build_uid`**。快照表加列或 join。

## 3. 状态枚举（前后端统一为 Go）

Execution：`running` | `completed` | `failed` | `aborted` | `calculated`

Case / Step：`running` | `passed` | `failed` | `skipped` | `broken`

Web 现有 `success` / `failure` / `pending` **废弃**，会话页与表格映射到上表。
