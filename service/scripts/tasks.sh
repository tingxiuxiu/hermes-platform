#!/usr/bin/env bash
# 本地开发任务脚本。
#
# 存在的意义：Windows 环境（如本机）没有 make 与 gcc，
# Makefile 只能在 Linux CI / WSL 下使用。本脚本只依赖 go 与 bash，
# 保证任何环境都能跑同一套命令。
#
# 用法：
#   ./scripts/tasks.sh test
#   ./scripts/tasks.sh migrate-verify
#   ./scripts/tasks.sh help

set -euo pipefail
cd "$(dirname "$0")/.."

cmd="${1:-help}"

case "$cmd" in
help | -h | --help)
  cat <<'EOF'
可用命令：
  build             编译全部二进制到 bin/
  fmt               格式化
  vet               go vet
  test              全量测试（真实 PG + 真实 Redis）
  test-race         带竞态检测（需要 gcc/cgo，Windows 需 mingw-w64 或 WSL）
  cover             生成覆盖率报告 coverage.html
  migrate-up        应用迁移
  migrate-down      回滚 1 步
  migrate-version   查看迁移版本
  migrate-baseline  存量库打基线（跳过 000001）
  migrate-verify    临时库验证 up→down→up
  testdb-reset      重建 hermes_test 测试库
  contract-diff     比对 Go 路由表与 Python router.py
  help              显示本帮助
EOF
  ;;

build)
  mkdir -p bin
  go build -o ./bin/ ./cmd/...
  echo "built into ./bin"
  ;;

fmt) go fmt ./... ;;
vet) go vet ./... ;;
lint)
  command -v golangci-lint >/dev/null 2>&1 || {
    echo "golangci-lint not installed: https://golangci-lint.run/welcome/install/"
    exit 1
  }
  golangci-lint run ./...
  ;;

test) go test ./... -count=1 ;;

test-race)
  if ! command -v gcc >/dev/null 2>&1; then
    echo "-race requires cgo and gcc is not in PATH."
    echo "  Windows: install mingw-w64, or run under WSL."
    echo "  Linux CI: gcc is preinstalled."
    exit 1
  fi
  CGO_ENABLED=1 go test -race ./... -count=1
  ;;

cover)
  go test -covermode=atomic -coverprofile=coverage.out ./... -count=1
  go tool cover -func=coverage.out | tail -1
  go tool cover -html=coverage.out -o coverage.html
  echo "report: coverage.html"
  ;;

migrate-up) go run ./cmd/migrate -cmd up ;;
migrate-down) go run ./cmd/migrate -cmd down -steps 1 ;;
migrate-version) go run ./cmd/migrate -cmd version ;;
migrate-baseline) go run ./cmd/migrate -cmd baseline ;;
migrate-verify) go run ./cmd/migrate -cmd verify ;;

testdb-reset)
  go test ./test/harness/... -run TestResetTestDatabase -count=1 -v
  ;;

contract-diff)
  # cygpath -w 把 Git Bash 的 /d/... 转成 Windows 的 D:\...（Go 的 filepath.Walk 需要）
  if command -v cygpath >/dev/null 2>&1; then
    PY_DIR="$(cygpath -w "$(cd "$(dirname "$0")/../.." && pwd)/service/src/app")"
    GO_DIR="$(cygpath -w "$(cd "$(dirname "$0")/.." && pwd)/internal/adapter/http")"
  else
    PY_DIR="$(cd "$(dirname "$0")/../.." && pwd)/service/src/app"
    GO_DIR="$(cd "$(dirname "$0")/.." && pwd)/internal/adapter/http"
  fi
  go run ./tools/contract-diff -python "$PY_DIR" -go "$GO_DIR"
  ;;

*)
  echo "unknown command: $cmd"
  echo "run '$0 help' for usage"
  exit 1
  ;;
esac
