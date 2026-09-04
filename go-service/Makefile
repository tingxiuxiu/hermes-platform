.PHONY: help build run run-api run-worker run-scheduler \
        fmt vet lint test test-race cover \
        migrate-up migrate-down migrate-version migrate-baseline migrate-force migrate-verify \
        sqlc-gen contract-diff testdb-reset clean

GO       := go
GOBIN    := $(shell pwd)/bin
MIGRATE  := $(GO) run ./cmd/migrate

help: ## 显示帮助
	@echo "Hermes Platform · go-service"
	@echo ""
	@echo "  构建与运行"
	@echo "    make build            编译 4 个二进制到 bin/"
	@echo "    make run-api          启动 HTTP API"
	@echo "    make run-worker       启动 asynq worker"
	@echo "    make run-scheduler    启动 asynq scheduler（务必单实例）"
	@echo ""
	@echo "  质量门禁"
	@echo "    make fmt              格式化"
	@echo "    make vet              go vet"
	@echo "    make lint             golangci-lint（含 depguard 依赖方向校验）"
	@echo "    make test             全量测试（真实 PG/Redis）"
	@echo "    make test-race        带竞态检测的全量测试"
	@echo "    make cover            覆盖率报告"
	@echo ""
	@echo "  数据库"
	@echo "    make migrate-up       应用迁移"
	@echo "    make migrate-down     回滚 1 步"
	@echo "    make migrate-version  查看版本"
	@echo "    make migrate-baseline 存量库打基线（跳过 000001）"
	@echo "    make migrate-verify   临时库上验证 up→down→up"
	@echo "    make testdb-reset     重建 hermes_test 测试库"
	@echo ""
	@echo "  其他"
	@echo "    make sqlc-gen         生成 sqlc 代码"
	@echo "    make contract-diff    比对 Go 路由表与 Python router.py"
	@echo "    make clean            清理产物"

# ============================================================================
# 构建与运行
# ============================================================================

build: ## 编译全部二进制
	mkdir -p $(GOBIN)
	$(GO) build -o $(GOBIN)/ ./cmd/...

run-api: ## 启动 HTTP API
	$(GO) run ./cmd/api

run-worker: ## 启动 asynq worker
	$(GO) run ./cmd/worker

run-scheduler: ## 启动 asynq scheduler
	$(GO) run ./cmd/scheduler

# ============================================================================
# 质量门禁
# ============================================================================

fmt: ## 格式化
	$(GO) fmt ./...

vet: ## go vet
	$(GO) vet ./...

lint: ## golangci-lint（未安装时给出提示而非失败）
	@command -v golangci-lint >/dev/null 2>&1 || \
	  { echo "golangci-lint not installed, see https://golangci-lint.run/welcome/install/"; exit 1; }
	golangci-lint run ./...

test: ## 全量测试（真实 PG + 真实 Redis）
	$(GO) test ./... -count=1

test-race: ## 带竞态检测（需要 cgo，Windows 上需先安装 gcc/mingw-w64）
	@command -v gcc >/dev/null 2>&1 || \
	  { echo "-race requires cgo and gcc is not in PATH."; \
	    echo "  Windows: install mingw-w64 (https://www.mingw-w64.org) or run in WSL."; \
	    echo "  CI on Linux has gcc preinstalled, so 'make test-race' works there."; \
	    exit 1; }
	CGO_ENABLED=1 $(GO) test -race ./... -count=1

cover: ## 覆盖率报告
	$(GO) test -covermode=atomic -coverprofile=coverage.out ./... -count=1
	$(GO) tool cover -func=coverage.out | tail -1
	$(GO) tool cover -html=coverage.out -o coverage.html

# ============================================================================
# 数据库迁移
# ============================================================================

migrate-up: ## 应用全部迁移
	$(MIGRATE) -cmd up

migrate-down: ## 回滚 1 步
	$(MIGRATE) -cmd down -steps 1

migrate-version: ## 查看当前版本
	$(MIGRATE) -cmd version

migrate-baseline: ## 存量库打基线（跳过 000001 建表）
	$(MIGRATE) -cmd baseline

migrate-force: ## 强制设置版本号：make migrate-force V=2
	$(MIGRATE) -cmd force -version $(V)

migrate-verify: ## 临时库验证 up→down→up 可逆
	$(MIGRATE) -cmd verify

testdb-reset: ## 重建 hermes_test 测试库
	$(GO) run ./cmd/migrate -cmd verify >/dev/null
	$(GO) test ./test/harness/... -run TestResetTestDatabase -count=1 -v

# ============================================================================
# 代码生成与契约校验
# ============================================================================

sqlc-gen: ## 生成 sqlc 代码
	@command -v sqlc >/dev/null 2>&1 || \
	  { echo "sqlc not installed, see https://docs.sqlc.dev/en/latest/overview/install.html"; exit 1; }
	sqlc generate

contract-diff: ## 比对 Go 与 Python 路由表
	$(GO) run ./tools/contract-diff -python $(abspath ../service/src/app) -go $(abspath internal/adapter/http)

clean: ## 清理产物
	rm -rf $(GOBIN) coverage.out coverage.html
