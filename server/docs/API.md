# API 文档

## 1. 概述

本文档描述了 Hermes Platform 的 API 接口，包括认证、测试任务、测试详情和测试记录的管理接口。

## 2. 基础信息

### 2.1 基本 URL

```
http://localhost:8080
```

### 2.2 认证方式

- **JWT 认证**: 使用 Bearer Token 进行认证
- **权限控制**: 基于角色的权限管理

### 2.3 响应格式

所有 API 响应都采用统一的 JSON 格式，无论请求成功或失败，HTTP 状态码始终为 **200**。

**响应结构**:
```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {}
}
```

**成功响应示例**:
```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "example"
  }
}
```

**失败响应示例**:
```json
{
  "success": false,
  "code": 400,
  "message": "bad request",
  "error": "Invalid parameter"
}
```

### 2.4 HTTP 方法

本 API 仅使用 **GET** 和 **POST** 两种 HTTP 方法：

- **GET**: 用于查询操作
- **POST**: 用于创建、更新、删除操作（更新和删除使用 URL 后缀区分）

| 操作类型 | 方法 | URL 模式 |
|---------|------|----------|
| 查询 | GET | /api/resource |
| 创建 | POST | /api/resource |
| 更新 | POST | /api/resource/:id/update |
| 删除 | POST | /api/resource/:id/delete |

## 3. 健康检查

### 3.1 健康检查接口

```
GET /health
```

**功能**: 检查服务是否正常运行

**响应示例**:

```json
{
  "status": "ok"
}
```

## 4. 认证相关接口

### 4.1 用户注册

```
POST /api/auth/register
```

**功能**: 注册新用户

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `username` | string | 是 | 用户名 |
| `email` | string | 是 | 邮箱地址 |
| `password` | string | 是 | 密码 |
| `role` | string | 否 | 角色（默认：user） |

**请求示例**:

```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "role": "user",
    "created_at": "2026-03-10T12:00:00Z"
  }
}
```

### 4.2 用户登录

```
POST /api/auth/login
```

**功能**: 用户登录并获取 JWT 令牌

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `email` | string | 是 | 邮箱地址 |
| `password` | string | 是 | 密码 |

**请求示例**:

```json
{
  "email": "test@example.com",
  "password": "password123"
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "role": "user"
    }
  }
}
```

### 4.3 修改密码

```
POST /api/auth/change-password
```

**功能**: 修改用户密码

**认证**: 需要 JWT 令牌

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `old_password` | string | 是 | 旧密码 |
| `new_password` | string | 是 | 新密码 |

**请求示例**:

```json
{
  "old_password": "password123",
  "new_password": "newpassword456"
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "message": "Password changed successfully"
  }
}
```

### 4.4 获取个人信息

```
GET /api/auth/profile
```

**功能**: 获取当前登录用户的个人信息

**认证**: 需要 JWT 令牌

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "role": "user",
    "created_at": "2026-03-10T12:00:00Z"
  }
}
```

### 4.5 更新个人信息

```
POST /api/auth/profile/update
```

**功能**: 更新当前登录用户的个人信息

**认证**: 需要 JWT 令牌

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `name` | string | 是 | 用户名 |

**请求示例**:

```json
{
  "name": "newusername"
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "message": "Profile updated successfully"
  }
}
```

## 5. 用户管理相关接口

### 5.1 获取用户列表

```
GET /api/users
```

**功能**: 获取用户列表

**认证**: 需要 JWT 令牌
**权限**: 需要 `user:manage` 权限

**查询参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `page` | int | 否 | 页码（默认：1） |
| `page_size` | int | 否 | 每页数量（默认：10） |
| `status` | string | 否 | 按状态过滤 |
| `name` | string | 否 | 按名称搜索 |
| `email` | string | 否 | 按邮箱搜索 |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "users": [
      {
        "id": 1,
        "name": "testuser",
        "email": "test@example.com",
        "roles": ["user"],
        "status": "active",
        "created_at": "2026-03-10T12:00:00Z",
        "updated_at": "2026-03-10T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

### 5.2 获取单个用户

```
GET /api/users/:id
```

**功能**: 根据 ID 获取用户详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `user:manage` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 用户 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "testuser",
    "email": "test@example.com",
    "roles": ["user"],
    "status": "active",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 5.3 更新用户

```
POST /api/users/:id/update
```

**功能**: 更新用户信息

**认证**: 需要 JWT 令牌
**权限**: 需要 `user:manage` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 用户 ID |

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `name` | string | 否 | 用户名 |
| `email` | string | 否 | 邮箱地址 |
| `status` | string | 否 | 用户状态（active/inactive） |

**请求示例**:

```json
{
  "name": "newusername",
  "status": "active"
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "message": "User updated successfully",
    "user": {
      "id": 1,
      "name": "newusername",
      "email": "test@example.com",
      "roles": ["user"],
      "status": "active"
    }
  }
}
```

### 5.4 删除用户

```
POST /api/users/:id/delete
```

**功能**: 删除用户

**认证**: 需要 JWT 令牌
**权限**: 需要 `user:manage` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 用户 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "message": "User deleted successfully"
  }
}
```

### 5.5 分配角色

```
POST /api/users/:id/assign-roles
```

**功能**: 为用户分配角色

**认证**: 需要 JWT 令牌
**权限**: 需要 `user:manage` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 用户 ID |

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `role_ids` | int[] | 是 | 角色 ID 列表 |

**请求示例**:

```json
{
  "role_ids": [1, 2]
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "message": "Roles assigned successfully"
  }
}
```

## 6. 测试任务相关接口

### 6.1 创建测试任务

```
POST /api/test-tasks
```

**功能**: 创建新的测试任务

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task:create` 权限

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `build_id` | string | 是 | 测试任务唯一标识符（UUID 格式） |
| `task_name` | string | 是 | 测试任务名称 |
| `worker_name` | string | 是 | 执行测试的节点名称 |
| `plan_key` | string | 是 | 测试所属的测试集 |
| `status` | string | 否 | 测试任务状态（默认：pending） |
| `start_time` | int64 | 否 | 开始时间戳 |

**请求示例**:

```json
{
  "build_id": "12345678-1234-1234-1234-123456789012",
  "task_name": "API 测试任务",
  "worker_name": "worker-node-01",
  "plan_key": "api-test-plan",
  "status": "pending",
  "start_time": 1609459200000
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "build_id": "12345678-1234-1234-1234-123456789012",
    "task_name": "API 测试任务",
    "worker_name": "worker-node-01",
    "plan_key": "api-test-plan",
    "status": "pending",
    "start_time": 1609459200000,
    "end_time": 0,
    "duration": 0,
    "total_tests": 0,
    "passed_tests": 0,
    "failed_tests": 0,
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 6.2 列出测试任务

```
GET /api/test-tasks
```

**功能**: 获取测试任务列表

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task:view` 权限

**查询参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `page` | int | 否 | 页码（默认：1） |
| `page_size` | int | 否 | 每页数量（默认：10） |
| `status` | string | 否 | 按状态过滤 |
| `worker_name` | string | 否 | 按工作节点过滤 |
| `plan_key` | string | 否 | 按测试计划过滤 |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "tasks": [
      {
        "id": 1,
        "build_id": "12345678-1234-1234-1234-123456789012",
        "task_name": "API 测试任务",
        "status": "completed",
        "worker_name": "worker-node-01",
        "plan_key": "api-test-plan"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

### 6.3 获取测试任务详情

```
GET /api/test-tasks/:id
```

**功能**: 根据 ID 获取测试任务详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task:view` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试任务 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "build_id": "12345678-1234-1234-1234-123456789012",
    "task_name": "API 测试任务",
    "status": "completed",
    "worker_name": "worker-node-01",
    "plan_key": "api-test-plan",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 6.4 更新测试任务

```
POST /api/test-tasks/:id/update
```

**功能**: 更新测试任务信息

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task:edit` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试任务 ID |

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `task_name` | string | 否 | 测试任务名称 |
| `status` | string | 否 | 测试任务状态 |

**请求示例**:

```json
{
  "status": "in_progress"
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "build_id": "12345678-1234-1234-1234-123456789012",
    "task_name": "API 测试任务",
    "status": "in_progress"
  }
}
```

### 6.5 删除测试任务

```
POST /api/test-tasks/:id/delete
```

**功能**: 删除测试任务

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task:delete` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试任务 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "message": "Test task deleted successfully"
  }
}
```

### 6.6 获取测试任务的测试详情

```
GET /api/test-tasks/:id/details
```

**功能**: 获取指定测试任务的测试详情列表

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail:view` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试任务 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "details": [
      {
        "id": 1,
        "test_task_id": 1,
        "test_name": "登录测试",
        "test_status": "passed",
        "created_at": "2026-03-10T12:00:00Z",
        "updated_at": "2026-03-10T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

### 6.7 获取测试任务的测试记录

```
GET /api/test-tasks/:id/records
```

**功能**: 获取指定测试任务的测试记录列表

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_record:view` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试任务 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "records": [
      {
        "id": 1,
        "test_task_id": 1,
        "test_detail_id": 1,
        "status": "passed",
        "result": "测试通过",
        "created_at": "2026-03-10T12:00:00Z",
        "updated_at": "2026-03-10T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

### 6.8 获取测试任务进度（基于 Build ID）

```
GET /api/test-tasks/buildid/:buildid/progress
```

**功能**: 根据 Build ID 获取测试任务的运行进度和通过率

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task:view` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `buildid` | string | 是 | Build ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "build_id": "12345678-1234-1234-1234-123456789012",
    "task_name": "API 测试任务",
    "status": "running",
    "progress": 60.0,
    "pass_rate": 66.67,
    "total_tests": 5,
    "passed_tests": 2,
    "failed_tests": 1,
    "completed_tests": 3
  }
}
```

## 7. 测试详情相关接口

### 7.1 创建测试详情

```
POST /api/test-details
```

**功能**: 创建新的测试详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail:create` 权限

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `test_task_id` | int | 是 | 测试任务 ID |
| `test_name` | string | 是 | 测试详情名称 |
| `error_message` | string | 否 | 错误信息 |
| `test_start_time` | int64 | 否 | 测试开始时间 |
| `test_end_time` | int64 | 否 | 测试结束时间 |
| `test_data` | string | 否 | 测试数据 |

**请求示例**:

```json
{
  "test_task_id": 1,
  "test_name": "登录测试",
  "test_start_time": 1609459200000
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "test_task_id": 1,
    "test_name": "登录测试",
    "test_status": "passed",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 7.2 获取测试详情

```
GET /api/test-details/:id
```

**功能**: 根据 ID 获取测试详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail:view` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试详情 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "test_task_id": 1,
    "test_name": "登录测试",
    "test_status": "passed",
    "error_message": "",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 7.3 更新测试详情

```
POST /api/test-details/:id/update
```

**功能**: 更新测试详情信息

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail:edit` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试详情 ID |

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `test_name` | string | 否 | 测试详情名称 |
| `test_status` | string | 否 | 测试状态（passed/failed/skipped） |
| `error_message` | string | 否 | 错误信息 |

**请求示例**:

```json
{
  "test_status": "failed",
  "error_message": "Login failed"
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "test_task_id": 1,
    "test_name": "登录测试",
    "test_status": "failed",
    "error_message": "Login failed"
  }
}
```

### 7.4 删除测试详情

```
POST /api/test-details/:id/delete
```

**功能**: 删除测试详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail:delete` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试详情 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "message": "Test detail deleted successfully"
  }
}
```

### 7.5 获取测试详情的测试步骤

```
GET /api/test-details/:id/steps
```

**功能**: 获取指定测试详情的测试步骤详情列表

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail:view` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试详情 ID |

**查询参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `page` | int | 否 | 页码（默认：1） |
| `page_size` | int | 否 | 每页数量（默认：10） |
| `passed` | bool | 否 | 按通过状态过滤 |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "steps": [
      {
        "id": 1,
        "test_detail_id": 1,
        "step_name": "输入用户名",
        "start_time": 1609459200000,
        "end_time": 1609459205000,
        "duration": 5000,
        "passed": true,
        "created_at": "2026-03-10T12:00:00Z",
        "updated_at": "2026-03-10T12:00:05Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

## 8. 测试记录相关接口

### 8.1 创建测试记录

```
POST /api/test-records
```

**功能**: 创建新的测试记录

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_record:create` 权限

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `test_task_id` | int | 是 | 测试任务 ID |
| `test_detail_id` | int | 是 | 测试详情 ID |

**请求示例**:

```json
{
  "test_task_id": 1,
  "test_detail_id": 1
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "test_task_id": 1,
    "test_detail_id": 1,
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 8.2 获取测试记录

```
GET /api/test-records/:id
```

**功能**: 根据 ID 获取测试记录

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_record:view` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试记录 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "test_task_id": 1,
    "test_detail_id": 1,
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 8.3 更新测试记录

```
POST /api/test-records/:id/update
```

**功能**: 更新测试记录信息

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_record:edit` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试记录 ID |

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `status` | string | 否 | 测试状态（passed/failed/pending） |
| `result` | string | 否 | 测试结果 |
| `notes` | string | 否 | 测试备注 |

**请求示例**:

```json
{
  "status": "failed",
  "result": "测试失败",
  "notes": "用户登录时出现错误"
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "test_task_id": 1,
    "test_detail_id": 1,
    "status": "failed",
    "result": "测试失败",
    "notes": "用户登录时出现错误"
  }
}
```

### 8.4 删除测试记录

```
POST /api/test-records/:id/delete
```

**功能**: 删除测试记录

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_record:delete` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试记录 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "message": "Test record deleted successfully"
  }
}
```

## 9. 测试步骤详情相关接口

### 9.1 创建测试步骤详情

```
POST /api/test-step-details
```

**功能**: 创建新的测试步骤详情，关联到测试详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail:create` 权限

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `test_detail_id` | int | 是 | 测试详情 ID |
| `step_name` | string | 是 | 步骤名称 |
| `start_time` | int64 | 是 | 开始时间（时间戳，毫秒） |
| `end_time` | int64 | 否 | 结束时间（时间戳，毫秒） |
| `passed` | bool | 否 | 是否通过 |
| `screenshot` | string | 否 | 截图（文件链接） |
| `verification_area` | string | 否 | 验证区域（JSON 格式） |

**请求示例**:

```json
{
  "test_detail_id": 1,
  "step_name": "输入用户名",
  "start_time": 1609459200000
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "test_detail_id": 1,
    "step_name": "输入用户名",
    "start_time": 1609459200000,
    "end_time": 0,
    "duration": 0,
    "passed": false,
    "screenshot": "",
    "verification_area": "",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 9.2 获取测试步骤详情

```
GET /api/test-step-details/:id
```

**功能**: 根据 ID 获取测试步骤详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail:view` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试步骤详情 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "test_detail_id": 1,
    "step_name": "输入用户名",
    "start_time": 1609459200000,
    "end_time": 1609459205000,
    "duration": 5000,
    "passed": true,
    "screenshot": "screenshot.png",
    "verification_area": "{\"x\": 0, \"y\": 0, \"width\": 100, \"height\": 100}",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:05Z"
  }
}
```

### 9.3 更新测试步骤详情

```
POST /api/test-step-details/:id/update
```

**功能**: 更新测试步骤详情，包括结束时间、是否通过、截图、验证区域等

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail:edit` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试步骤详情 ID |

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `end_time` | int64 | 否 | 结束时间（时间戳，毫秒） |
| `passed` | bool | 否 | 是否通过 |
| `screenshot` | string | 否 | 截图（文件链接） |
| `verification_area` | string | 否 | 验证区域（JSON 格式） |

**请求示例**:

```json
{
  "end_time": 1609459205000,
  "passed": true,
  "screenshot": "screenshot.png",
  "verification_area": "{\"x\": 0, \"y\": 0, \"width\": 100, \"height\": 100}"
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "test_detail_id": 1,
    "step_name": "输入用户名",
    "start_time": 1609459200000,
    "end_time": 1609459205000,
    "duration": 5000,
    "passed": true,
    "screenshot": "screenshot.png",
    "verification_area": "{\"x\": 0, \"y\": 0, \"width\": 100, \"height\": 100}",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:05Z"
  }
}
```

### 9.4 删除测试步骤详情

```
POST /api/test-step-details/:id/delete
```

**功能**: 删除测试步骤详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail:delete` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试步骤详情 ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "message": "Test step detail deleted successfully"
  }
}
```

## 10. 统计相关接口

### 10.1 获取仪表盘统计

```
GET /api/stats/dashboard
```

**功能**: 获取仪表盘统计数据

**认证**: 需要 JWT 令牌

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "total_tasks": 100,
    "running_tasks": 5,
    "completed_tasks": 80,
    "failed_tasks": 15,
    "total_tests": 1000,
    "passed_tests": 850,
    "failed_tests": 150
  }
}
```

### 10.2 获取趋势数据

```
GET /api/stats/trend
```

**功能**: 获取测试趋势数据

**认证**: 需要 JWT 令牌

**查询参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `days` | int | 否 | 查询天数（默认：7） |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "trend": [
      {
        "date": "2026-03-10",
        "total": 10,
        "passed": 8,
        "failed": 2
      }
    ]
  }
}
```

### 10.3 获取运行中的任务

```
GET /api/stats/running-tasks
```

**功能**: 获取当前正在运行的任务

**认证**: 需要 JWT 令牌

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "tasks": [
      {
        "id": 1,
        "build_id": "12345678-1234-1234-1234-123456789012",
        "task_name": "API 测试任务",
        "worker_name": "worker-node-01",
        "progress": 60.0,
        "start_time": 1609459200000
      }
    ]
  }
}
```

## 11. 错误码

所有 API 响应都使用 HTTP 200 状态码，错误信息通过响应体中的 `code` 和 `success` 字段表示：

| 错误码 | 描述 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（需要登录或 token 无效） |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

**错误响应示例**:

```json
{
  "success": false,
  "code": 400,
  "message": "bad request",
  "error": "Invalid parameter: email"
}
```

## 12. RSA 加密接口

### 12.1 获取 RSA 公钥

```
GET /api/auth/rsa-pubkey
```

**功能**: 获取 RSA 公钥用于密码加密

**认证**: 无需认证

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...\n-----END PUBLIC KEY-----"
  }
}
```

### 12.2 加密登录

```
POST /api/auth/login
```

**功能**: 用户登录，支持明文密码或加密密码

**认证**: 无需认证

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `email` | string | 是 | 邮箱地址 |
| `password` | string | 否 | 明文密码（当 encrypted_pwd 为空时使用） |
| `encrypted_pwd` | string | 否 | RSA 加密后的密码 |

**请求示例** (加密密码):

```json
{
  "email": "test@example.com",
  "password": "",
  "encrypted_pwd": "base64_encoded_encrypted_password"
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "message": "Login successful",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

## 13. API Token 管理接口

### 13.1 创建 API Token

```
POST /api/tokens
```

**功能**: 创建 365 天有效的 API Token

**认证**: 需要 JWT 令牌

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `name` | string | 是 | Token 名称 |

**请求示例**:

```json
{
  "name": "自动化测试 Token"
}
```

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "自动化测试 Token",
    "token": "abc123xyz789...",
    "expires_at": "2027-03-13 12:00:00",
    "is_revoked": false,
    "created_at": "2026-03-13 12:00:00"
  }
}
```

### 13.2 获取 API Token 列表

```
GET /api/tokens
```

**功能**: 获取当前用户的所有 API Token

**认证**: 需要 JWT 令牌

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "tokens": [
      {
        "id": 1,
        "name": "自动化测试 Token",
        "token": "abc***xyz",
        "expires_at": "2027-03-13 12:00:00",
        "is_revoked": false,
        "created_at": "2026-03-13 12:00:00"
      }
    ]
  }
}
```

### 13.3 撤销 API Token

```
POST /api/tokens/:id/delete
```

**功能**: 撤销指定的 API Token

**认证**: 需要 JWT 令牌

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | Token ID |

**响应示例** (成功):

```json
{
  "success": true,
  "code": 200,
  "message": "success",
  "data": {
    "message": "Token revoked successfully"
  }
}
```

## 14. 最佳实践

1. **认证**: 所有需要认证的接口都需要在请求头中包含 `Authorization: Bearer <token>`
2. **错误处理**: 客户端应该检查响应中的 `success` 字段来判断请求是否成功
3. **参数验证**: 客户端应该在发送请求前验证参数的有效性
4. **速率限制**: 系统对请求有速率限制（每分钟 60 个请求），请合理控制请求频率
5. **权限管理**: 确保用户拥有足够的权限执行相应的操作
6. **安全登录**: 登录时请使用 RSA 公钥加密密码后再传输（通过 `encrypted_pwd` 字段）
7. **Token 管理**: API Token 只会在创建时显示一次，请妥善保存
