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

所有 API 响应都采用 JSON 格式，包含以下字段：

- `success`: 布尔值，表示请求是否成功
- `data`: 响应数据（成功时）
- `error`: 错误信息（失败时）
- `code`: 错误代码（失败时）

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

**响应示例**:

```json
{
  "success": true,
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

**响应示例**:

```json
{
  "success": true,
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

**响应示例**:

```json
{
  "success": true,
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

**响应示例**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "role": "user",
    "created_at": "2026-03-10T12:00:00Z"
  }
}
```

## 5. 测试任务相关接口

### 5.1 创建测试任务

```
POST /api/test-tasks
```

**功能**: 创建新的测试任务

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task_create` 权限

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `build_id` | string | 是 | 测试任务唯一标识符（UUID 格式） |
| `task_name` | string | 是 | 测试任务名称 |
| `worker_name` | string | 是 | 执行测试的节点名称 |
| `plan_key` | string | 是 | 测试所属的测试集 |
| `status` | string | 否 | 测试任务状态（默认：pending） |
| `start_time` | int64 | 否 | 开始时间戳 |
| `end_time` | int64 | 否 | 结束时间戳 |

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

**响应示例**:

```json
{
  "success": true,
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

### 5.2 列出测试任务

```
GET /api/test-tasks
```

**功能**: 获取测试任务列表

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task_read` 权限

**查询参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `page` | int | 否 | 页码（默认：1） |
| `page_size` | int | 否 | 每页数量（默认：10） |
| `status` | string | 否 | 按状态过滤 |
| `priority` | string | 否 | 按优先级过滤 |

**响应示例**:

```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "id": 1,
        "name": "API 测试任务",
        "description": "测试所有 API 接口",
        "status": "pending",
        "priority": "high",
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

### 5.3 获取测试任务详情

```
GET /api/test-tasks/:id
```

**功能**: 根据 ID 获取测试任务详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task_read` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试任务 ID |

**响应示例**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "API 测试任务",
    "description": "测试所有 API 接口",
    "status": "pending",
    "priority": "high",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 5.4 更新测试任务

```
PUT /api/test-tasks/:id
```

**功能**: 更新测试任务信息

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task_update` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试任务 ID |

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `name` | string | 否 | 测试任务名称 |
| `description` | string | 否 | 测试任务描述 |
| `status` | string | 否 | 测试任务状态 |
| `priority` | string | 否 | 测试任务优先级 |

**请求示例**:

```json
{
  "status": "in_progress",
  "priority": "medium"
}
```

**响应示例**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "API 测试任务",
    "description": "测试所有 API 接口",
    "status": "in_progress",
    "priority": "medium",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T13:00:00Z"
  }
}
```

### 5.5 删除测试任务

```
DELETE /api/test-tasks/:id
```

**功能**: 删除测试任务

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task_delete` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试任务 ID |

**响应示例**:

```json
{
  "success": true,
  "data": {
    "message": "Test task deleted successfully"
  }
}
```

### 5.6 获取测试任务的测试详情

```
GET /api/test-tasks/:id/details
```

**功能**: 获取指定测试任务的测试详情列表

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail_read` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试任务 ID |

**响应示例**:

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "test_task_id": 1,
      "name": "登录测试",
      "description": "测试用户登录功能",
      "created_at": "2026-03-10T12:00:00Z",
      "updated_at": "2026-03-10T12:00:00Z"
    }
  ]
}
```

### 5.7 获取测试任务的测试记录

```
GET /api/test-tasks/:id/records
```

**功能**: 获取指定测试任务的测试记录列表

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_record_read` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试任务 ID |

**响应示例**:

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "test_task_id": 1,
      "test_detail_id": 1,
      "status": "passed",
      "result": "测试通过",
      "created_at": "2026-03-10T12:00:00Z",
      "updated_at": "2026-03-10T12:00:00Z"
    }
  ]
}
```

### 5.8 获取测试任务进度（基于 UUID）

```
GET /api/test-tasks/uuid/:uuid/progress
```

**功能**: 根据 UUID 获取测试任务的运行进度和通过率

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_task_read` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `uuid` | string | 是 | 测试任务 UUID |

**响应示例**:

```json
{
  "success": true,
  "data": {
    "uuid": "12345678-1234-1234-1234-123456789012",
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

**响应字段说明**:

| 字段 | 类型 | 描述 |
|------|------|------|
| `uuid` | string | 测试任务 UUID |
| `task_name` | string | 测试任务名称 |
| `status` | string | 测试任务状态 |
| `progress` | float | 测试任务进度（百分比） |
| `pass_rate` | float | 测试任务通过率（百分比） |
| `total_tests` | int | 总测试数 |
| `passed_tests` | int | 通过测试数 |
| `failed_tests` | int | 失败测试数 |
| `completed_tests` | int | 已完成测试数 |

## 6. 测试详情相关接口

### 6.1 创建测试详情

```
POST /api/test-details
```

**功能**: 创建新的测试详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail_create` 权限

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `test_task_id` | int | 是 | 测试任务 ID |
| `name` | string | 是 | 测试详情名称 |
| `description` | string | 否 | 测试详情描述 |
| `expected_result` | string | 否 | 预期结果 |

**请求示例**:

```json
{
  "test_task_id": 1,
  "name": "登录测试",
  "description": "测试用户登录功能",
  "expected_result": "用户能够成功登录"
}
```

**响应示例**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "test_task_id": 1,
    "name": "登录测试",
    "description": "测试用户登录功能",
    "expected_result": "用户能够成功登录",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 6.2 获取测试详情

```
GET /api/test-details/:id
```

**功能**: 根据 ID 获取测试详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail_read` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试详情 ID |

**响应示例**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "test_task_id": 1,
    "name": "登录测试",
    "description": "测试用户登录功能",
    "expected_result": "用户能够成功登录",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 6.3 更新测试详情

```
PUT /api/test-details/:id
```

**功能**: 更新测试详情信息

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail_update` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试详情 ID |

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `name` | string | 否 | 测试详情名称 |
| `description` | string | 否 | 测试详情描述 |
| `expected_result` | string | 否 | 预期结果 |

**请求示例**:

```json
{
  "description": "测试用户登录和登出功能",
  "expected_result": "用户能够成功登录和登出"
}
```

**响应示例**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "test_task_id": 1,
    "name": "登录测试",
    "description": "测试用户登录和登出功能",
    "expected_result": "用户能够成功登录和登出",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T13:00:00Z"
  }
}
```

### 6.4 删除测试详情

```
DELETE /api/test-details/:id
```

**功能**: 删除测试详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail_delete` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试详情 ID |

**响应示例**:

```json
{
  "success": true,
  "data": {
    "message": "Test detail deleted successfully"
  }
}
```

## 7. 测试记录相关接口

### 7.1 创建测试记录

```
POST /api/test-records
```

**功能**: 创建新的测试记录

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_record_create` 权限

**请求参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `test_task_id` | int | 是 | 测试任务 ID |
| `test_detail_id` | int | 是 | 测试详情 ID |
| `status` | string | 是 | 测试状态（passed/failed/pending） |
| `result` | string | 否 | 测试结果 |
| `notes` | string | 否 | 测试备注 |

**请求示例**:

```json
{
  "test_task_id": 1,
  "test_detail_id": 1,
  "status": "passed",
  "result": "测试通过",
  "notes": "用户能够成功登录系统"
}
```

**响应示例**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "test_task_id": 1,
    "test_detail_id": 1,
    "status": "passed",
    "result": "测试通过",
    "notes": "用户能够成功登录系统",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 7.2 获取测试记录

```
GET /api/test-records/:id
```

**功能**: 根据 ID 获取测试记录

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_record_read` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试记录 ID |

**响应示例**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "test_task_id": 1,
    "test_detail_id": 1,
    "status": "passed",
    "result": "测试通过",
    "notes": "用户能够成功登录系统",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T12:00:00Z"
  }
}
```

### 7.3 更新测试记录

```
PUT /api/test-records/:id
```

**功能**: 更新测试记录信息

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_record_update` 权限

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

**响应示例**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "test_task_id": 1,
    "test_detail_id": 1,
    "status": "failed",
    "result": "测试失败",
    "notes": "用户登录时出现错误",
    "created_at": "2026-03-10T12:00:00Z",
    "updated_at": "2026-03-10T13:00:00Z"
  }
}
```

### 7.4 删除测试记录

```
DELETE /api/test-records/:id
```

**功能**: 删除测试记录

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_record_delete` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试记录 ID |

**响应示例**:

```json
{
  "success": true,
  "data": {
    "message": "Test record deleted successfully"
  }
}
```

## 8. 错误码

| 错误码 | 描述 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 8. 测试步骤详情相关接口

### 8.1 创建测试步骤详情

```
POST /api/test-step-details
```

**功能**: 创建新的测试步骤详情，关联到测试详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail_create` 权限

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

**响应示例**:

```json
{
  "success": true,
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

### 8.2 获取测试步骤详情

```
GET /api/test-step-details/:id
```

**功能**: 根据 ID 获取测试步骤详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail_read` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试步骤详情 ID |

**响应示例**:

```json
{
  "success": true,
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

### 8.3 更新测试步骤详情

```
PUT /api/test-step-details/:id
```

**功能**: 更新测试步骤详情，包括结束时间、是否通过、截图、验证区域等

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail_update` 权限

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

**响应示例**:

```json
{
  "success": true,
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

### 8.4 删除测试步骤详情

```
DELETE /api/test-step-details/:id
```

**功能**: 删除测试步骤详情

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail_delete` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试步骤详情 ID |

**响应示例**:

```json
{
  "success": true,
  "data": {
    "message": "Test step detail deleted successfully"
  }
}
```

### 8.5 获取测试详情的测试步骤详情

```
GET /api/test-details/:id/steps
```

**功能**: 获取指定测试详情的测试步骤详情列表

**认证**: 需要 JWT 令牌
**权限**: 需要 `test_detail_read` 权限

**路径参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `id` | int | 是 | 测试详情 ID |

**查询参数**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `page` | int | 否 | 页码（默认：1） |
| `page_size` | int | 否 | 每页数量（默认：10） |

**响应示例**:

```json
{
  "success": true,
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
        "screenshot": "screenshot.png",
        "verification_area": "{\"x\": 0, \"y\": 0, \"width\": 100, \"height\": 100}",
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

## 9. 错误码

| 错误码 | 描述 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 10. 最佳实践

1. **认证**: 所有需要认证的接口都需要在请求头中包含 `Authorization: Bearer <token>`
2. **错误处理**: 客户端应该处理 API 返回的错误信息
3. **参数验证**: 客户端应该在发送请求前验证参数的有效性
4. **速率限制**: 系统对请求有速率限制（每分钟 60 个请求），请合理控制请求频率
5. **权限管理**: 确保用户拥有足够的权限执行相应的操作