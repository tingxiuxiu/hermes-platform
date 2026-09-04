# go-service · Hermes Platform 服务端（Go）

## 目录结构

```
cmd/
  api/         HTTP API 主进程（优雅停机 + 首启 seed）
  worker/      asynq 消费端（dashboard 快照投影）
  scheduler/   asynq 定时投递端（refresh:all，⚠️ 单实例）
  migrate/     golang-migrate CLI 包装（up/down/version/baseline/verify）
internal/
  domain/      领域模型（automation / dashboard / identity / system）
  application/ 应用用例（用例编排、事件发布）
  adapter/     HTTP 处理器、持久化仓储、asynq worker、缓存
  platform/    基础设施（database / redis / queue / config / telemetry ...）
  bootstrap/   组合根：唯一 import 所有层的包（ADR-0004）
migrations/    SQL 迁移文件（golang-migrate 格式）
tools/         contract-diff 契约一致性校验工具
scripts/       Windows 友好任务脚本（无 make/gcc 环境）
test/          集成测试 harness 与端到端用例
```

## 架构原则（ADR）

- **分层依赖方向**（ADR-0004）：`domain ← application ← adapter ← bootstrap`。
  `bootstrap` 是唯一 import 所有层的组合根。
- **契约一致**（ADR-0001）：路径/方法/鉴权/响应信封与 Python 版一致。
  `tools/contract-diff` 校验 Go 路由表与 Python `router.py`。
- **真实测试**（ADR-0007）：集成测试连真实 PG + Redis，禁止 mock 替代。

## 实时会话

心跳、步骤 upsert、SSE、`GET /live` 与 Web 会话页见 [docs/live-execution](../docs/live-execution/README.md)。

## 本地启动

依赖：Go 1.26+、本地 PostgreSQL、本地 Redis（或 Docker）。

### 1. 配置环境变量

从仓库根目录 `.env` 或进程环境变量读取（Go 进程工作目录在 `go-service/` 时，也可把变量写进 shell）：

```bash
export POSTGRES_SERVER=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=hermes
export POSTGRES_PASSWORD=hermes123
export POSTGRES_DB=hermes
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_DB=0
export AUTOMATION_SERVICE_TOKEN=change-this-service-token
export FIRST_SUPERUSER=admin@example.com
export FIRST_SUPERUSER_PASSWORD=change-this-admin-password
export SECRET_KEY=change-this-secret-key
```

### 2. 数据库迁移

```bash
make migrate-up            # 新库：应用全部迁移
make migrate-baseline      # 存量库（已由 alembic 建表）：打基线后补迁移
make migrate-verify        # 临时库验证 up→down→up 可逆
```

### 3. 启动进程

```bash
make run-api               # HTTP API（8080）
make run-worker            # asynq worker
make run-scheduler         # asynq scheduler（单实例！）
```

或编译全部二进制：

```bash
make build                 # 输出到 bin/
```

## 测试

需要测试库与独立 Redis DB（包之间互不共用，避免并行 `go test ./...` 互相 TRUNCATE / FLUSHDB）：

- `test/harness`：默认 `hermes_test` + Redis db15
- `test/integration`：`hermes_integration` + Redis db14

```bash
make testdb-reset          # 重建测试库
make test                  # 全量测试（真实 PG + Redis）
make test-race             # 带竞态检测（需 gcc，Windows 用 WSL/mingw）
make cover                 # 覆盖率报告
```

> Windows 无 gcc 时 `-race` 不可用（Makefile 前置检测并提示），
> Linux CI 自带 gcc 正常生效。Windows 也可用 `./scripts/tasks.sh test`。

## 质量门禁

```bash
make fmt         # gofmt
make vet         # go vet
make lint        # golangci-lint（含 depguard 依赖方向校验）
make contract-diff  # 比对 Go 路由表与 Python router.py
```

## 容器化

```bash
# 构建镜像（4 个二进制 + migrations）
docker build -f Dockerfile -t hermes-go:latest .

# 用 docker-compose 编排（api/worker/scheduler/migrate + postgres/redis）
docker compose up -d migrate api worker scheduler
```

详见 `docker-compose.yml` 的 Go 服务段。

## 关键环境变量

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `HTTP_PORT` | 8080 | API 监听端口 |
| `POSTGRES_SERVER` / `POSTGRES_PORT` / `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | - | PostgreSQL 连接 |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_DB` | - | Redis（asynq broker） |
| `AUTOMATION_SERVICE_TOKEN` | - | 上报接口服务令牌 |
| `FIRST_SUPERUSER` / `FIRST_SUPERUSER_PASSWORD` | - | 首启超级管理员（幂等） |
| `SECRET_KEY` | - | JWT 签名密钥 |
| `AUTOMATION_DASHBOARD_REFRESH_SECONDS` | 60 | scheduler 投递 refresh:all 周期 |
| `AUTOMATION_DASHBOARD_TREND_DAYS` | 7 | dashboard 趋势窗口 |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | 空 | OpenTelemetry 端点；空则 no-op（`-tags otel` 启用完整实现） |

## OpenTelemetry

默认构建走 no-op 降级（`OTEL_EXPORTER_OTLP_ENDPOINT` 为空时不阻塞启动）。
启用完整 trace：

```bash
go build -tags otel -o bin/api ./cmd/api
```
