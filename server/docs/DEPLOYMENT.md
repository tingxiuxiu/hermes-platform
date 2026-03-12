# 部署文档

## 1. 项目概述

Hermes Platform 是一个测试管理平台，提供测试任务、测试详情和测试记录的管理功能，同时包含用户认证和权限管理系统。

## 2. 技术栈

- **语言**: Go 1.20+
- **Web 框架**: Gin
- **数据库**: PostgreSQL
- **缓存**: Redis
- **认证**: JWT

## 3. 环境要求

- Go 1.20 或更高版本
- PostgreSQL 12.0 或更高版本
- Redis 6.0 或更高版本（可选，用于缓存）

## 4. 配置文件

项目使用 YAML 格式的配置文件，位于 `internal/config/` 目录：

- `config.development.yaml`: 开发环境配置
- `config.production.yaml`: 生产环境配置
- `config.test.yaml`: 测试环境配置

### 配置文件示例

```yaml
# 服务器配置
server:
  host: "0.0.0.0"
  port: "8080"

# 数据库配置
database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "password"
  dbname: "hermes_platform"
  sslmode: "disable"

# Redis 配置
redis:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0

# JWT 配置
jwt:
  secret: "your-secret-key"
  expireHours: 24

# 环境配置
env: "development"
```

## 5. 部署步骤

### 5.1 准备环境

1. **安装 Go**
   - 从 [Go 官网](https://golang.org/dl/) 下载并安装 Go 1.20+ 版本
   - 验证安装：`go version`

2. **安装 PostgreSQL**
   - 从 [PostgreSQL 官网](https://www.postgresql.org/download/) 下载并安装
   - 创建数据库：`CREATE DATABASE hermes_platform;`

3. **安装 Redis** (可选)
   - 从 [Redis 官网](https://redis.io/download/) 下载并安装
   - 启动 Redis 服务

### 5.2 部署流程

1. **克隆代码**
   ```bash
   git clone <repository-url>
   cd hermes-platform/server
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **构建项目**
   ```bash
   go build -o server.exe ./cmd/server
   ```

4. **配置环境变量**
   - 生产环境：设置 `CONFIG_PATH` 环境变量指向生产配置文件
   - 开发环境：默认使用 `config.development.yaml`

5. **启动服务**
   ```bash
   # 开发环境
   go run ./cmd/server
   
   # 生产环境
   ./server.exe
   ```

### 5.3 容器化部署 (Docker)

**Dockerfile 示例**:

```dockerfile
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY internal/config/config.production.yaml ./config.yaml

EXPOSE 8080

CMD ["./server"]
```

**Docker Compose 示例**:

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - db
      - redis
    environment:
      - CONFIG_PATH=/app/config.yaml

  db:
    image: postgres:13-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=hermes_platform
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:6-alpine

volumes:
  postgres_data:
```

## 6. 系统初始化

服务启动时会自动执行以下操作：

1. 加载配置文件
2. 初始化 JWT 配置
3. 连接数据库并自动迁移模型
4. 初始化默认数据（包括管理员账户）
5. 尝试连接 Redis（失败不影响服务启动）
6. 注册 API 路由和中间件
7. 启动服务器

## 7. 健康检查

服务提供健康检查接口：

```bash
GET /health
```

响应示例：
```json
{
  "status": "ok"
}
```

## 8. 监控与日志

- 服务会记录请求日志到标准输出
- 错误信息会被捕获并返回标准化的错误响应
- 可以配置外部监控工具（如 Prometheus、Grafana）监控服务状态

## 9. 常见问题

### 9.1 数据库连接失败
- 检查数据库服务是否运行
- 验证数据库配置是否正确
- 确保数据库用户有足够的权限

### 9.2 Redis 连接失败
- Redis 连接失败不会阻止服务启动
- 若需要 Redis 功能，请检查 Redis 服务是否运行

### 9.3 端口被占用
- 修改配置文件中的 `server.port` 为其他可用端口

### 9.4 JWT 令牌无效
- 确保 JWT 密钥配置正确
- 检查令牌是否过期

## 10. 升级与维护

- 备份数据库定期备份
- 监控系统性能和错误日志
- 定期更新依赖包和安全补丁
- 测试环境验证后再部署到生产环境