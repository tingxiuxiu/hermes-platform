<div align="center">

# 🔱 Hermes Platform

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![React Version](https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react)](https://reactjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

**中文** | [**English**](#english)

**🚀 现代化的全栈测试任务管理平台 | Modern Full-Stack Test Task Management Platform**

<p align="center">
  <a href="#功能特性--features">功能特性</a> •
  <a href="#技术栈--tech-stack">技术栈</a> •
  <a href="#快速开始--quick-start">快速开始</a> •
  <a href="#api-文档--api-docs">API 文档</a>
</p>

</div>

---

<details open>
<summary><h2>🇨🇳 中文</h2></summary>

## ✨ 功能特性

<table>
<tr>
<td width="50%">

### 🔐 安全认证
- RSA 加密登录传输
- JWT 无状态认证
- API Token 长效访问 (365天)
- 基于角色的权限控制 (RBAC)

### 📊 测试管理
- 测试任务 CRUD 操作
- 测试用例详情管理
- 执行历史记录追踪
- 数据统计仪表盘

</td>
<td width="50%">

### 🛠️ 开发者友好
- RESTful API 设计
- 完整的 TypeScript 类型支持
- 热重载开发模式
- 详细的 API 文档

### 🌍 国际化
- 中英文切换支持
- 响应式布局设计
- 现代化 UI (shadcn/ui)
- 流畅的用户体验

</td>
</tr>
</table>

## 🏗️ 技术栈

### 后端
| 技术 | 用途 | 版本 |
|------|------|------|
| [Go](https://golang.org) | 主要语言 | 1.25+ |
| [Gin](https://gin-gonic.com) | Web 框架 | v1.12 |
| [GORM](https://gorm.io) | ORM 框架 | v1.31 |
| [PostgreSQL](https://postgresql.org) | 主数据库 | 14+ |
| [Redis](https://redis.io) | 缓存 | 6+ |
| [JWT](https://jwt.io) | 身份认证 | v5 |

### 前端
| 技术 | 用途 | 版本 |
|------|------|------|
| [React](https://react.dev) | UI 框架 | 19 |
| [TypeScript](https://typescriptlang.org) | 类型安全 | 5.9 |
| [Vite](https://vitejs.dev) | 构建工具 | 7 |
| [Tailwind CSS](https://tailwindcss.com) | 样式框架 | 4 |
| [shadcn/ui](https://ui.shadcn.com) | 组件库 | latest |
| [Zustand](https://zustand-demo.pmnd.rs) | 状态管理 | 5 |

## 🚀 快速开始

### 前置要求

- ✅ Go 1.25+
- ✅ Node.js 18+
- ✅ PostgreSQL 14+
- ✅ Redis 6+

### 1️⃣ 克隆项目

```bash
git clone https://github.com/tingxiuxiu/hermes-platform.git
cd hermes-platform
```

### 2️⃣ 启动后端

```bash
cd server

# 安装依赖
go mod download

# 创建数据库
psql -U postgres -c "CREATE DATABASE hermes_dev;"
psql -U postgres -d hermes_dev -f init_database.sql

# 生成 RSA 密钥对
openssl genrsa -out rsa_private_key.pem 2048
openssl rsa -in rsa_private_key.pem -pubout -out rsa_public_key.pem

# 启动服务
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动 🎉

### 3️⃣ 启动前端

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

应用将在 `http://localhost:5173` 启动 🚀

## 📚 API 文档

完整的 API 文档请查看 [server/docs/API.md](server/docs/API.md)

### 认证接口

```http
POST   /api/auth/register      # 用户注册
POST   /api/auth/login         # 用户登录 (RSA 加密)
POST   /api/auth/logout        # 用户登出
GET    /api/auth/public-key    # 获取 RSA 公钥
```

### 核心资源

```http
# 测试任务
GET    /api/test-tasks              # 列表
POST   /api/test-tasks              # 创建
POST   /api/test-tasks/:id/update   # 更新
POST   /api/test-tasks/:id/delete   # 删除

# API Token 管理
GET    /api/tokens                  # 列表
POST   /api/tokens                  # 创建 (365天有效期)
POST   /api/tokens/:id/revoke       # 撤销
```

## 🏛️ 项目架构

```
hermes-platform/
├── 🎨 frontend/              # React 前端
│   ├── src/
│   │   ├── components/      # UI 组件
│   │   ├── pages/           # 页面
│   │   ├── services/        # API 服务
│   │   └── stores/          # 状态管理
│   └── package.json
│
├── ⚙️ server/                # Go 后端
│   ├── cmd/server/          # 入口
│   ├── internal/
│   │   ├── api/             # HTTP 处理器
│   │   ├── auth/            # 认证中间件
│   │   ├── crypto/          # RSA 加密
│   │   ├── models/          # 数据模型
│   │   ├── repository/      # 数据访问
│   │   └── services/        # 业务逻辑
│   └── docs/                # 文档
│
└── 📖 README.md
```

## 🔒 安全特性

| 特性 | 说明 |
|------|------|
| 🔐 **RSA 加密** | 登录密码使用 RSA 公钥加密传输 |
| 🎫 **JWT 认证** | 无状态身份验证，支持 Token 过期 |
| 🔑 **API Token** | 365 天有效期，支持撤销 |
| 🛡️ **密码哈希** | bcrypt 加密存储 |
| 👥 **RBAC** | 基于角色的访问控制 |
| ⚡ **速率限制** | API 请求频率限制 |

## 🧪 测试

```bash
# 后端测试
cd server
go test ./... -v

# 前端测试
cd frontend
npm run lint
npm run build
```

## 📝 配置

```yaml
# server/internal/config/config.development.yaml
server:
  port: "8080"
  host: "0.0.0.0"

database:
  host: "127.0.0.1"
  port: "5432"
  user: "postgres"
  password: "postgres"
  dbname: "hermes_dev"

jwt:
  secret: "your-secret-key"
  expiration: 24

redis:
  host: "127.0.0.1"
  port: "6379"
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

### 🔐 Security
- RSA encrypted login transmission
- JWT stateless authentication
- API Token long-term access (365 days)
- Role-based access control (RBAC)

### 📊 Test Management
- Test task CRUD operations
- Test case detail management
- Execution history tracking
- Statistics dashboard

</td>
<td width="50%">

### 🛠️ Developer Friendly
- RESTful API design
- Complete TypeScript type support
- Hot reload development mode
- Detailed API documentation

### 🌍 Internationalization
- Chinese/English language switch
- Responsive layout design
- Modern UI (shadcn/ui)
- Smooth user experience

</td>
</tr>
</table>

## 🏗️ Tech Stack

### Backend
| Technology | Purpose | Version |
|------------|---------|---------|
| [Go](https://golang.org) | Main Language | 1.25+ |
| [Gin](https://gin-gonic.com) | Web Framework | v1.12 |
| [GORM](https://gorm.io) | ORM Framework | v1.31 |
| [PostgreSQL](https://postgresql.org) | Database | 14+ |
| [Redis](https://redis.io) | Cache | 6+ |
| [JWT](https://jwt.io) | Authentication | v5 |

### Frontend
| Technology | Purpose | Version |
|------------|---------|---------|
| [React](https://react.dev) | UI Framework | 19 |
| [TypeScript](https://typescriptlang.org) | Type Safety | 5.9 |
| [Vite](https://vitejs.dev) | Build Tool | 7 |
| [Tailwind CSS](https://tailwindcss.com) | Styling | 4 |
| [shadcn/ui](https://ui.shadcn.com) | Component Library | latest |
| [Zustand](https://zustand-demo.pmnd.rs) | State Management | 5 |

## 🚀 Quick Start

### Prerequisites

- ✅ Go 1.25+
- ✅ Node.js 18+
- ✅ PostgreSQL 14+
- ✅ Redis 6+

### 1️⃣ Clone the Project

```bash
git clone https://github.com/tingxiuxiu/hermes-platform.git
cd hermes-platform
```

### 2️⃣ Start Backend

```bash
cd server

# Install dependencies
go mod download

# Create database
psql -U postgres -c "CREATE DATABASE hermes_dev;"
psql -U postgres -d hermes_dev -f init_database.sql

# Generate RSA key pair
openssl genrsa -out rsa_private_key.pem 2048
openssl rsa -in rsa_private_key.pem -pubout -out rsa_public_key.pem

# Start server
go run cmd/server/main.go
```

Server will start at `http://localhost:8080` 🎉

### 3️⃣ Start Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev
```

App will start at `http://localhost:5173` 🚀

## 📚 API Documentation

Full API docs at [server/docs/API.md](server/docs/API.md)

### Authentication

```http
POST   /api/auth/register      # User registration
POST   /api/auth/login         # User login (RSA encrypted)
POST   /api/auth/logout        # User logout
GET    /api/auth/public-key    # Get RSA public key
```

### Core Resources

```http
# Test Tasks
GET    /api/test-tasks              # List
POST   /api/test-tasks              # Create
POST   /api/test-tasks/:id/update   # Update
POST   /api/test-tasks/:id/delete   # Delete

# API Token Management
GET    /api/tokens                  # List
POST   /api/tokens                  # Create (365 days validity)
POST   /api/tokens/:id/revoke       # Revoke
```

## 🏛️ Project Structure

```
hermes-platform/
├── 🎨 frontend/              # React Frontend
│   ├── src/
│   │   ├── components/      # UI Components
│   │   ├── pages/           # Pages
│   │   ├── services/        # API Services
│   │   └── stores/          # State Management
│   └── package.json
│
├── ⚙️ server/                # Go Backend
│   ├── cmd/server/          # Entry Point
│   ├── internal/
│   │   ├── api/             # HTTP Handlers
│   │   ├── auth/            # Auth Middleware
│   │   ├── crypto/          # RSA Encryption
│   │   ├── models/          # Data Models
│   │   ├── repository/      # Data Access
│   │   └── services/        # Business Logic
│   └── docs/                # Documentation
│
└── 📖 README.md
```

## 🔒 Security Features

| Feature | Description |
|---------|-------------|
| 🔐 **RSA Encryption** | Login passwords encrypted with RSA public key |
| 🎫 **JWT Auth** | Stateless authentication with token expiration |
| 🔑 **API Token** | 365 days validity, revocable |
| 🛡️ **Password Hash** | bcrypt encrypted storage |
| 👥 **RBAC** | Role-based access control |
| ⚡ **Rate Limiting** | API request frequency limit |

## 🧪 Testing

```bash
# Backend testing
cd server
go test ./... -v

# Frontend testing
cd frontend
npm run lint
npm run build
```

## 📝 Configuration

```yaml
# server/internal/config/config.development.yaml
server:
  port: "8080"
  host: "0.0.0.0"

database:
  host: "127.0.0.1"
  port: "5432"
  user: "postgres"
  password: "postgres"
  dbname: "hermes_dev"

jwt:
  secret: "your-secret-key"
  expiration: 24

redis:
  host: "127.0.0.1"
  port: "6379"
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
