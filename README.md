<div align="center">

# 🔱 Hermes Platform

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Gin](https://img.shields.io/badge/Gin-1.12-00ADD8?style=flat-square&logo=go)](https://gin-gonic.com)
[![React Version](https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react)](https://reactjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

**中文** | [**English**](#english)

**🚀 自动化测试执行追踪与团队管理平台 | Automation Test Execution Tracking & Team Management Platform**

<p align="center">
  <a href="#-功能特性">功能特性</a> •
  <a href="#-技术栈">技术栈</a> •
  <a href="#-快速开始">快速开始</a> •
  <a href="#-api-文档">API 文档</a> •
  <a href="#-项目架构">项目架构</a>
</p>

</div>

---

<details open>
<summary><h2>🇨🇳 中文</h2></summary>

## ✨ 功能特性

<table>
<tr>
<td width="50%">

### 🔐 双重鉴权体系

- 用户 JWT 无状态登录鉴权
- 自动化上报接口使用独立 `X-Service-Token` 服务令牌，与用户登录态完全区分
- 基于角色的权限控制 (RBAC)，权限树管理
- 密码使用 Argon2id 哈希存储

### 🧪 自动化测试执行追踪

- Execution → Item → Step → Attachment 分层数据模型
- 用例重试历史保留，支持失败重跑追溯
- 供 pytest / CI Runner 上报执行结果的专用接口
- 面向前端 Dashboard 的查询类接口 (执行列表、用例详情、历史记录)

</td>
<td width="50%">

### 🛠️ 开发者友好

- Go + Gin，六边形分层（`domain` / `application` / `adapter` / `bootstrap`）
- 模块化单体 + 多二进制（`api` / `worker` / `scheduler` / `migrate`）
- pgx + golang-migrate 数据库迁移
- 完整的 TypeScript 类型支持，前后端契约文档齐全

### 📈 可观测性

- OpenTelemetry 全链路追踪 (Gin / pgx / Redis)
- slog 结构化日志
- `/healthz` 存活探针与 `/readyz` 就绪探针

</td>
</tr>
</table>

## 🏗️ 技术栈

### 后端 (`service/`)

| 技术                                        | 用途           | 版本    |
| ------------------------------------------- | -------------- | ------- |
| [Go](https://go.dev)                        | 主要语言       | 1.26+   |
| [Gin](https://gin-gonic.com)                | Web 框架       | 1.12    |
| [pgx](https://github.com/jackc/pgx)         | PostgreSQL 驱动 | v5     |
| [golang-migrate](https://github.com/golang-migrate/migrate) | 数据库迁移 | v4 |
| [PostgreSQL](https://postgresql.org)         | 主数据库       | 14+     |
| [Redis](https://redis.io)                    | 缓存 / 任务队列 | 6+      |
| [golang-jwt](https://github.com/golang-jwt/jwt) | 用户身份认证 | v5 |
| [Argon2id](https://pkg.go.dev/golang.org/x/crypto/argon2) | 密码哈希 | latest |
| [asynq](https://github.com/hibiken/asynq)    | 异步任务 / 定时调度 | 0.25+ |
| [OpenTelemetry](https://opentelemetry.io)    | 可观测性/链路追踪 | 1.46+ |
| [slog](https://pkg.go.dev/log/slog)          | 结构化日志     | 标准库  |

### 前端 (`web/`)

| 技术                                            | 用途             | 版本   |
| ----------------------------------------------- | ---------------- | ------ |
| [React](https://react.dev)                      | UI 框架          | 19     |
| [TypeScript](https://typescriptlang.org)        | 类型安全         | 5.9    |
| [Vite](https://vitejs.dev)                      | 构建工具         | 8      |
| [TanStack Router](https://tanstack.com/router)  | 路由 (文件路由)  | v1     |
| [TanStack Query](https://tanstack.com/query)    | 服务端状态管理   | v5     |
| [TanStack Table](https://tanstack.com/table)    | 表格             | v8     |
| [Tailwind CSS](https://tailwindcss.com)         | 样式框架         | 4      |
| [shadcn/ui](https://ui.shadcn.com) + Radix UI   | 组件库           | latest |
| [Zustand](https://zustand-demo.pmnd.rs)         | 客户端状态管理   | 5      |
| [Clerk](https://clerk.com)                      | 第三方登录 (可选)| v6     |
| [Axios](https://axios-http.com)                 | HTTP 客户端      | v1     |
| [Zod](https://zod.dev) + React Hook Form        | 表单校验         | latest |
| [Vitest](https://vitest.dev) + [Playwright](https://playwright.dev) | 单测/浏览器测试 | v4 |
| [pnpm](https://pnpm.io)                         | 包管理           | latest |

## 🚀 快速开始

### 前置要求

- ✅ Go 1.26+
- ✅ Node.js 20+ 与 [pnpm](https://pnpm.io)
- ✅ PostgreSQL 14+
- ✅ Redis 6+

### 1️⃣ 克隆项目

```bash
git clone https://github.com/tingxiuxiu/hermes-platform.git
cd hermes-platform
```

### 2️⃣ 配置环境变量

在仓库根目录创建 `.env`（`service/` 会向上查找并读取），至少需要包含：

```bash
PROJECT_NAME=Hermes Platform
ENVIRONMENT=local

# HTTP（本地开发建议 8080，与前端 Vite 代理一致；默认 80）
SERVICE_PORT=8080

# PostgreSQL
POSTGRES_SERVER=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=hermes_dev

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# 首个超级管理员账号
FIRST_SUPERUSER=admin@example.com
FIRST_SUPERUSER_PASSWORD=change-this-password

# JWT 签名密钥（生产环境必须修改，否则非 local 环境启动会直接报错）
SECRET_KEY=change-this-secret-key

# 自动化用例上报接口专用服务令牌，与用户登录 JWT 完全区分
AUTOMATION_SERVICE_TOKEN=change-this-service-token
```

### 3️⃣ 启动后端 (`service/`)

```bash
cd service

# 执行数据库迁移
make migrate-up

# 启动 HTTP API
make run-api
```

服务将在 `http://localhost:8080` 启动 🎉
存活探针：`http://localhost:8080/healthz`
就绪探针：`http://localhost:8080/readyz`

可选：启动 asynq worker / scheduler（用于异步刷新 dashboard 快照）：

```bash
make run-worker
make run-scheduler   # 务必单实例
```

Windows 无 make 时可用 `./scripts/tasks.sh`（见 [`service/README.md`](service/README.md)）。

### 4️⃣ 启动前端 (`web/`)

```bash
cd web

# 安装依赖
pnpm install

# 如需 Clerk 登录，配置 VITE_CLERK_PUBLISHABLE_KEY（可选）
cp .env.example .env.local

# 启动开发服务器
pnpm dev
```

应用将在 `http://localhost:5173` 启动 🚀（已配置 `/tap/api` 反向代理到后端 8080 端口）

## 📚 API 文档

- **后端实现**：[`service/`](service/README.md)（Gin，前缀 `/tap/api/v1`）
- **设计文档**：[docs/service/](docs/service/README.md)
- **实时会话设计**：[docs/live-execution/](docs/live-execution/README.md)
- **探针**：`/healthz`（存活）、`/readyz`（就绪，检查 PostgreSQL + Redis）
- **契约说明**：[docs/live-execution/04-api.md](docs/live-execution/04-api.md)

### 认证接口

```http
POST   /tap/api/v1/login               # 用户登录 (JWT)
POST   /tap/api/v1/login/access-token  # OAuth2 兼容登录
POST   /tap/api/v1/login/refresh       # 刷新令牌
POST   /tap/api/v1/register            # 用户注册
POST   /tap/api/v1/logout              # 用户登出
```

### 自动化测试执行追踪

```http
# 查询类 (前端 Dashboard / 会话页，用户 JWT)
GET    /tap/api/v1/automation/dashboard/overview
GET    /tap/api/v1/automation/dashboard/executions/{execution_id}/cases
GET    /tap/api/v1/automation/pipelines
GET    /tap/api/v1/automation/executions
GET    /tap/api/v1/automation/executions/{build_uid}
GET    /tap/api/v1/automation/executions/{build_uid}/items
GET    /tap/api/v1/automation/executions/{build_uid}/live
GET    /tap/api/v1/automation/executions/{build_uid}/events
GET    /tap/api/v1/automation/items/{case_uid}

# 上报类 (pytest / CI，X-Service-Token)
POST   /tap/api/v1/automation/executions
PATCH  /tap/api/v1/automation/executions/{build_uid}
POST   /tap/api/v1/automation/executions/{build_uid}/heartbeat
POST   /tap/api/v1/automation/executions/{build_uid}/items
PATCH  /tap/api/v1/automation/executions/{build_uid}/items/{case_uid}
POST   /tap/api/v1/automation/items/{case_uid}/steps
```

v1 附件只写 Allure，不上报 Hermes。完整契约见 [docs/live-execution/04-api.md](docs/live-execution/04-api.md)。

## 🏛️ 项目架构

```
hermes-platform/
├── ⚙️ service/                 # Go 后端 (Gin + asynq)
├── 🎨 web/                     # React 前端 (TanStack Router + shadcn/ui)
├── 🔌 hermes_plugin/           # pytest 实时上报插件
├── 📖 docs/service/            # Go 服务设计 / ADR
├── 📖 docs/live-execution/     # 实时会话设计
├── 🐳 docker/                  # Dockerfile
├── 📡 etc/otel/                # OpenTelemetry Collector 配置
├── docker-compose.yml
└── 📖 README.md
```

## 🔒 安全特性

| 特性                  | 说明                                                         |
| --------------------- | ------------------------------------------------------------ |
| 🎫 **JWT 用户认证**   | 无状态身份验证，供前端登录态使用                              |
| 🔑 **服务令牌鉴权**   | 自动化上报接口使用独立的 `X-Service-Token`，与用户登录态区分  |
| 🛡️ **密码哈希**       | Argon2id 加密存储                                             |
| 👥 **RBAC**           | 基于角色的访问控制，支持权限树管理                            |
| ⚠️ **默认密钥保护**   | `SECRET_KEY` / `AUTOMATION_SERVICE_TOKEN` 等若在非 local 环境仍为默认值，启动时直接报错拒绝运行 |

## 🧪 测试

```bash
# 后端测试 (service/)
cd service
make test          # 全量测试（真实 PostgreSQL + Redis）
make lint          # golangci-lint

# 前端测试 (web/)
cd web
pnpm lint
pnpm test
pnpm build
```

## 🤝 贡献

1. 🍴 Fork 本仓库
2. 🌿 创建分支 (`git checkout -b feature/AmazingFeature`)
3. 💾 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 📤 推送分支 (`git push origin feature/AmazingFeature`)
5. 🔃 创建 Pull Request

</details>

---

<details>
<summary><h2 id="english">🇺🇸 English</h2></summary>

## ✨ Features

<table>
<tr>
<td width="50%">

### 🔐 Dual Authentication

- Stateless JWT authentication for user login
- Automation reporting endpoints use an independent `X-Service-Token`, fully decoupled from user login state
- Role-based access control (RBAC) with permission tree management
- Passwords hashed with Argon2id

### 🧪 Automation Test Execution Tracking

- Layered data model: Execution → Item → Step → Attachment
- Retry history preserved for traceable re-runs of failed cases
- Dedicated reporting endpoints for pytest / CI runners
- Query endpoints for the frontend dashboard (execution list, case detail, history)

</td>
<td width="50%">

### 🛠️ Developer Friendly

- Go + Gin with hexagonal layering (`domain` / `application` / `adapter` / `bootstrap`)
- Modular monolith with multiple binaries (`api` / `worker` / `scheduler` / `migrate`)
- pgx + golang-migrate for schema migrations
- Full TypeScript typing with a dedicated frontend integration doc

### 📈 Observability

- OpenTelemetry tracing across Gin / pgx / Redis
- Structured logging via slog
- `/healthz` liveness and `/readyz` readiness probes

</td>
</tr>
</table>

## 🏗️ Tech Stack

### Backend (`service/`)

| Technology                                   | Purpose                        | Version |
| --------------------------------------------- | ------------------------------- | ------- |
| [Go](https://go.dev)                          | Main Language                  | 1.26+   |
| [Gin](https://gin-gonic.com)                  | Web Framework                  | 1.12    |
| [pgx](https://github.com/jackc/pgx)           | PostgreSQL Driver              | v5      |
| [golang-migrate](https://github.com/golang-migrate/migrate) | Database Migrations | v4 |
| [PostgreSQL](https://postgresql.org)           | Database                       | 14+     |
| [Redis](https://redis.io)                      | Cache / Task Broker            | 6+      |
| [golang-jwt](https://github.com/golang-jwt/jwt) | User Authentication          | v5      |
| [Argon2id](https://pkg.go.dev/golang.org/x/crypto/argon2) | Password Hashing    | latest  |
| [asynq](https://github.com/hibiken/asynq)      | Async Jobs / Scheduling        | 0.25+   |
| [OpenTelemetry](https://opentelemetry.io)      | Observability / Tracing        | 1.46+   |
| [slog](https://pkg.go.dev/log/slog)            | Structured Logging             | stdlib  |

### Frontend (`web/`)

| Technology                                       | Purpose                  | Version |
| ------------------------------------------------- | ------------------------- | ------- |
| [React](https://react.dev)                         | UI Framework              | 19      |
| [TypeScript](https://typescriptlang.org)           | Type Safety               | 5.9     |
| [Vite](https://vitejs.dev)                          | Build Tool                | 8       |
| [TanStack Router](https://tanstack.com/router)      | File-based Routing        | v1      |
| [TanStack Query](https://tanstack.com/query)        | Server State Management   | v5      |
| [TanStack Table](https://tanstack.com/table)        | Tables                     | v8      |
| [Tailwind CSS](https://tailwindcss.com)             | Styling                    | 4       |
| [shadcn/ui](https://ui.shadcn.com) + Radix UI       | Component Library          | latest  |
| [Zustand](https://zustand-demo.pmnd.rs)             | Client State Management    | 5       |
| [Clerk](https://clerk.com)                           | Third-party Auth (optional)| v6      |
| [Axios](https://axios-http.com)                      | HTTP Client                 | v1      |
| [Zod](https://zod.dev) + React Hook Form             | Form Validation             | latest  |
| [Vitest](https://vitest.dev) + [Playwright](https://playwright.dev) | Unit/Browser Testing | v4 |
| [pnpm](https://pnpm.io)                              | Package Management          | latest  |

## 🚀 Quick Start

### Prerequisites

- ✅ Go 1.26+
- ✅ Node.js 20+ with [pnpm](https://pnpm.io)
- ✅ PostgreSQL 14+
- ✅ Redis 6+

### 1️⃣ Clone the Project

```bash
git clone https://github.com/tingxiuxiu/hermes-platform.git
cd hermes-platform
```

### 2️⃣ Configure Environment Variables

Create a `.env` file at the repository root (`service/` walks upward to load it), with at least:

```bash
PROJECT_NAME=Hermes Platform
ENVIRONMENT=local

# HTTP (use 8080 locally to match the Vite proxy; default is 80)
SERVICE_PORT=8080

# PostgreSQL
POSTGRES_SERVER=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=hermes_dev

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# First superuser account
FIRST_SUPERUSER=admin@example.com
FIRST_SUPERUSER_PASSWORD=change-this-password

# JWT signing key (must be changed for non-local environments, or startup fails)
SECRET_KEY=change-this-secret-key

# Dedicated service token for automation reporting endpoints, decoupled from user JWT
AUTOMATION_SERVICE_TOKEN=change-this-service-token
```

### 3️⃣ Start the Backend (`service/`)

```bash
cd service

# Run database migrations
make migrate-up

# Start the HTTP API
make run-api
```

Server will start at `http://localhost:8080` 🎉
Liveness: `http://localhost:8080/healthz`
Readiness: `http://localhost:8080/readyz`

Optional: start the asynq worker / scheduler (for async dashboard snapshot refresh):

```bash
make run-worker
make run-scheduler   # single instance only
```

On Windows without make, use `./scripts/tasks.sh` (see [`service/README.md`](service/README.md)).

### 4️⃣ Start the Frontend (`web/`)

```bash
cd web

# Install dependencies
pnpm install

# Optional: configure VITE_CLERK_PUBLISHABLE_KEY if using Clerk login
cp .env.example .env.local

# Start the dev server
pnpm dev
```

App will start at `http://localhost:5173` 🚀 (already proxies `/tap/api` to the backend on port 8080)

## 📚 API Documentation

- **Backend**: [`service/`](service/README.md) (Gin, prefix `/tap/api/v1`)
- **Design docs**: [docs/service/](docs/service/README.md)
- **Live session design**: [docs/live-execution/](docs/live-execution/README.md)
- **Probes**: `/healthz` (liveness), `/readyz` (readiness, checks PostgreSQL + Redis)
- **Contract**: [docs/live-execution/04-api.md](docs/live-execution/04-api.md)

### Authentication

```http
POST   /tap/api/v1/login               # User login (JWT)
POST   /tap/api/v1/login/access-token  # OAuth2-compatible login
POST   /tap/api/v1/login/refresh       # Refresh token
POST   /tap/api/v1/register            # User registration
POST   /tap/api/v1/logout              # User logout
```

### Automation Execution Tracking

```http
# Query (dashboard / session page, JWT)
GET    /tap/api/v1/automation/dashboard/overview
GET    /tap/api/v1/automation/dashboard/executions/{execution_id}/cases
GET    /tap/api/v1/automation/pipelines
GET    /tap/api/v1/automation/executions
GET    /tap/api/v1/automation/executions/{build_uid}
GET    /tap/api/v1/automation/executions/{build_uid}/items
GET    /tap/api/v1/automation/executions/{build_uid}/live
GET    /tap/api/v1/automation/executions/{build_uid}/events
GET    /tap/api/v1/automation/items/{case_uid}

# Reporting (pytest / CI, X-Service-Token)
POST   /tap/api/v1/automation/executions
PATCH  /tap/api/v1/automation/executions/{build_uid}
POST   /tap/api/v1/automation/executions/{build_uid}/heartbeat
POST   /tap/api/v1/automation/executions/{build_uid}/items
PATCH  /tap/api/v1/automation/executions/{build_uid}/items/{case_uid}
POST   /tap/api/v1/automation/items/{case_uid}/steps
```

v1 attachments go to Allure only, not Hermes. Full contract: [docs/live-execution/04-api.md](docs/live-execution/04-api.md).

## 🏛️ Project Structure

```
hermes-platform/
├── ⚙️ service/                 # Go backend (Gin + asynq)
├── 🎨 web/                     # React frontend (TanStack Router + shadcn/ui)
├── 🔌 hermes_plugin/           # pytest live-reporting plugin
├── 📖 docs/service/            # Go service design / ADRs
├── 📖 docs/live-execution/     # Live session design
├── 🐳 docker/                  # Dockerfile
├── 📡 etc/otel/                # OpenTelemetry Collector config
├── docker-compose.yml
└── 📖 README.md
```

## 🔒 Security Features

| Feature                     | Description                                                              |
| ---------------------------- | -------------------------------------------------------------------------- |
| 🎫 **JWT User Auth**         | Stateless authentication for the frontend login session                  |
| 🔑 **Service Token Auth**    | Automation reporting endpoints use an independent `X-Service-Token`, decoupled from user login |
| 🛡️ **Password Hashing**      | Argon2id encrypted storage                                                |
| 👥 **RBAC**                  | Role-based access control with permission tree management                |
| ⚠️ **Default Secret Guard**  | Startup fails if `SECRET_KEY` / `AUTOMATION_SERVICE_TOKEN`, etc. are left at defaults in non-local environments |

## 🧪 Testing

```bash
# Backend tests (service/)
cd service
make test          # full suite (real PostgreSQL + Redis)
make lint          # golangci-lint

# Frontend tests (web/)
cd web
pnpm lint
pnpm test
pnpm build
```

## 🤝 Contributing

1. 🍴 Fork this repository
2. 🌿 Create branch (`git checkout -b feature/AmazingFeature`)
3. 💾 Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. 📤 Push to branch (`git push origin feature/AmazingFeature`)
5. 🔃 Create Pull Request

</details>

---

<div align="center">

## 📄 License

This project is licensed under the [MIT License](LICENSE)

**[⬆ Back to Top](#-hermes-platform)**

Made with ❤️ by [tingxiuxiu](https://github.com/tingxiuxiu)

</div>
