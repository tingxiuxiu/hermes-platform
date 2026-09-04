# 多阶段构建：Hermes Platform · go-service
#
# 产物：api / worker / scheduler / migrate 4 个二进制 + migrations 目录。
# 用法：docker build -f go-service/Dockerfile -t hermes-go:latest .
#       或用 docker-compose（见 T-10.4 编排改造）。
# =============================================================================

# ---------- 构建阶段 ----------
FROM golang:1.26 AS builder

WORKDIR /src

# 先拷贝依赖清单，利用缓存层加速
COPY go-service/go.mod go-service/go.sum ./
RUN go mod download

# 拷贝源码
COPY go-service/. .

# 编译 4 个二进制（禁用 CGO 便于静态链接，减少运行镜像体积）
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api       ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker    ./cmd/worker \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/scheduler ./cmd/scheduler \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate   ./cmd/migrate

# ---------- 运行阶段 ----------
FROM alpine:3.20

# CA 证书用于 HTTPS 出站（OTLP 上报、外部回调）
RUN apk add --no-cache ca-certificates tzdata \
 && addgroup -S hermes && adduser -S hermes -G hermes

WORKDIR /app

# 拷贝二进制与迁移文件
COPY --from=builder /out/ /usr/local/bin/
COPY --from=builder /src/migrations ./migrations

# 非 root 运行
USER hermes

# 健康检查探针（api 进程）
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:${HTTP_PORT:-8080}/healthz || exit 1

EXPOSE 8080

# 默认启动 api（compose 里用 command 覆盖启动 worker/scheduler/migrate）
CMD ["api"]
