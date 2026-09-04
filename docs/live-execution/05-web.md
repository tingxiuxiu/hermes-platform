# 05 · Web 会话页

## 1. 路由

| 路径 | 用途 |
| --- | --- |
| `/automation/jobs` | Pipeline 列表；running 且仅一次 build → 链到会话 |
| `/automation/executions` | Execution 列表 |
| `/automation/executions/$buildUid` | **会话页**（running 与 history 共用） |
| Dashboard | 统计；running 卡点进会话页（需 `build_uid`） |

去掉以数字 `execution.id` 为主键的详情依赖。内部表格可仍显示 id，跳转用 `build_uid`。

## 2. 会话页结构

后台壳体（侧栏、顶栏）不变。主区：

```
┌─────────────┬──────────────────────────────┐
│ Case 列表    │ 主卡片：当前或选中的 case      │
│ 摘要/状态    │ 步骤树，当前 running 节点高亮   │
│ 进行中置顶   │ 计时 / 错误信息                 │
└─────────────┴──────────────────────────────┘
```

- `status=running` 的 execution：挂载 SSE；`current_item` 默认选中。
- 已结束：不订 SSE；默认选中最后结束的 case 或 URL `?case=`。
- 步骤树复用并改造现有 `StepTree`：支持 `running`、按 `step_path` 应用 delta，**不要**每次 SSE 整树重挂丢折叠状态。

## 3. 数据流

1. 进入页：`GET /live`
2. 若 `execution.status=running`：`fetch` SSE `/events`
3. `step.upserted`：在对应 case 的树里 upsert 节点（无则插入，有则改状态）
4. `item.updated`：更新列表行；若变为终态且无其它 running，主卡片保持该 case
5. `execution.updated` 终止：关闭 SSE，再 GET `/live`
6. 断线：指数退避后 **再 GET /live**，再订 SSE

点选已结束 case：`GET /items/{case_uid}` 填主卡片（不依赖 SSE）。

## 4. API 客户端

- `VITE_API_BASE_URL` 默认 `/tap/api/v1`（或 axios `baseURL=/tap/api/v1`）。
- 删除 `/hermes-platform/api/v1`。
- Vite proxy：`/tap/api` → `http://localhost:8080`。
- schema 与 Go DTO 对齐（`build_uid`、`passed`/`failed` 等）。

## 5. Jobs 一键打开

`GET /pipelines` 若能提供 `last_build_uid` + 最近 execution 是否 running：有则 Link 到 `/automation/executions/{last_build_uid}`。否则链到 executions 列表 `?jobName=`。

Pipeline 行若没有 running 语义，会话入口以 **executions 列表 status=running** 和 **dashboard running 卡** 为主。Jobs 页文案写明：运行状态以插件上报为准。

## 6. 视觉（Apple 作法则，不当营销页）

参考 `.agents/skills/apple-design/DESIGN.md` 的约束，**密度按状态板**：

- 交互色只用 Action Blue `#0066cc`（暗面链接可用 `#2997ff`）。
- 卡片：白底、`18px` 圆角、`1px` hairline `#e0e0e0`，**无阴影**。
- 主按钮：蓝胶囊；按压缩放 `scale(0.95)`。
- 不要全幅 80px tile、不要摄影区、不要第二品牌色、不要卡片阴影堆层级。
- 正文约 17px；标题用系统 UI / Inter，略收 tracking。

shadcn 组件可留，用 token 覆盖，不要新开一套组件库。

## 7. 文案

会话页顶栏：Job 名 + `build_uid` 短号 + status。Running 时副文案：「以插件心跳为准；进程被杀约 60 秒后标记中止。」
