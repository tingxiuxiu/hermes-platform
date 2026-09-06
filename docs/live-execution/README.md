# Live Execution · 实时会话

本目录是 **pytest 运行期实时会话** 的设计源：用户点开一次正在跑的 Execution，用卡片看到当前 case 精确到步骤（进入/退出约 1 秒可见）。

本设计只约束 `service`、`web`、`hermes_plugin`。

| 文档 | 内容 |
| --- | --- |
| [01-overview.md](./01-overview.md) | 产品契约、锁定决策、明确不做 |
| [02-architecture.md](./02-architecture.md) | 写侧 / live 热路径 / 统计快照、心跳、SSE |
| [03-plugin.md](./03-plugin.md) | Allure 为埋点源、插件监听与上报时机 |
| [04-api.md](./04-api.md) | 新增 HTTP 契约（心跳、步骤 upsert、live 快照、SSE） |
| [05-web.md](./05-web.md) | 会话页、导航、鉴权、视觉约束 |
| [TASKS.md](./TASKS.md) | 切片路线、验收、测试接缝 |

相关既有文档：[docs/go-service](../service/README.md)（六边形分层、双令牌、asynq 范围）。本特性 **不** 把 live 步骤树挂进 dashboard 快照；asynq `Unique(5m)` 保持给统计用。
