# Go Service

Go 服务端设计文档（六边形分层、契约、迁移与测试）。实现代码见仓库 [`go-service/`](../../go-service/README.md)。

| 文档 | 内容 |
| --- | --- |
| [01-architecture.md](./01-architecture.md) | 架构总览 |
| [02-domain-model.md](./02-domain-model.md) | 领域模型 |
| [03-api-contract.md](./03-api-contract.md) | API 契约 |
| [04-migration-and-testing.md](./04-migration-and-testing.md) | 迁移与测试 |
| [05-known-defects.md](./05-known-defects.md) | 已知缺陷 |
| [glossary.md](./glossary.md) | 术语 |
| [TASKS.md](./TASKS.md) | 切片任务 |

## ADR

| 编号 | 决策 |
| --- | --- |
| [ADR-0001](./adr/ADR-0001-contract-baseline.md) | 契约基线 |
| [ADR-0002](./adr/ADR-0002-modular-monolith-multi-binary.md) | 模块化单体 + 多二进制 |
| [ADR-0003](./adr/ADR-0003-schema-baseline.md) | Schema 基线 |
| [ADR-0004](./adr/ADR-0004-hexagonal-layering.md) | 六边形分层 |
| [ADR-0005](./adr/ADR-0005-auth-dual-token.md) | 双令牌鉴权 |
| [ADR-0006](./adr/ADR-0006-asynq-scope.md) | asynq 范围 |
| [ADR-0007](./adr/ADR-0007-testing-real-infra.md) | 真实基础设施测试 |
| [ADR-0008](./adr/ADR-0008-golang-migrate.md) | golang-migrate |
| [ADR-0009](./adr/ADR-0009-defect-handling.md) | 缺陷处理 |

实时会话（心跳、步骤 upsert、SSE、`GET /live`、Web 会话页）见 [live-execution](../live-execution/README.md)。
