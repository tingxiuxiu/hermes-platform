# Hermes Platform 后端接口文档 (v1)

> 本文档面向前端开发，用于对接当前后端已实现的全部接口。内容基于代码实际实现整理（非纯设计稿），如有出入以下列源码为准：
> `service/src/app/api/v1/{auth,users,roles,permissions,automation}.py`
>
> 交互式文档 (Swagger / ReDoc，可在线试调、查看实时 OpenAPI Schema)：
> - Swagger UI: `{BASE_URL}{API_V1_STR}/docs`
> - ReDoc: `{BASE_URL}{API_V1_STR}/redoc`

---

## 1. 基础约定

### 1.1 Base URL

所有接口均挂载在统一前缀下（见 `service/src/app/config.py` 的 `API_V1_STR`，默认值）：

```
/hermes-platform/api/v1
```

下文所有路径均为相对该前缀的路径，例如 "`POST /login`" 实际请求地址为 `POST /hermes-platform/api/v1/login`。

### 1.2 统一响应结构（成功响应）

所有**成功**响应（HTTP 200）均包裹为统一信封：

```typescript
interface NormalResponse<T = unknown> {
  code: number;      // 业务码，成功固定为 200
  message: string;   // 响应描述信息，如 "Get User List Success"
  data: T;           // 实际业务数据，具体结构见各接口
  success: true;
}
```

### 1.3 分页数据结构

所有分页列表接口的 `data` 字段均为如下结构（`items` 的类型随接口不同）：

```typescript
interface PageResult<T> {
  total: number;      // 总记录数
  page: number;        // 当前页码 (从 1 开始)
  page_size: number;    // 每页条数
  items: T[];
}
```

### 1.4 错误响应结构 ⚠️ 重要

**后端当前未注册统一异常处理中间件**，所有错误（包括业务异常、404、401/403 鉴权失败）均直接使用 **FastAPI/Starlette 默认错误结构**，*不是* 上面的 `NormalResponse` 信封：

```typescript
interface HttpErrorResponse {
  detail: string;
}
```

参数校验失败（如必填字段缺失、类型不对）由 FastAPI 自动返回 `422 Unprocessable Entity`，结构为：

```typescript
interface ValidationErrorResponse {
  detail: Array<{
    loc: (string | number)[];
    msg: string;
    type: string;
  }>;
}
```

前端统一错误处理建议：优先读取 `error.response.data.detail`；`detail` 为字符串时直接展示，为数组时取第一条 `msg` 展示。

常见 HTTP 状态码：

| 状态码 | 场景 |
| --- | --- |
| 400 | 业务校验失败（如用户名密码错误、账号已存在、账号被禁用） |
| 401 | 未提供 / 无效的服务令牌（自动化上报接口专用，见 1.6） |
| 403 | 未登录或 JWT 失效、或已登录但权限不足（非管理员访问管理员接口） |
| 404 | 目标资源不存在（用户/角色/execution/item/step 等） |
| 422 | 请求体/查询参数校验失败 |

### 1.5 鉴权方式一：用户登录态 (JWT)

除注册/登录接口和自动化**上报类**接口外，其余所有接口都需要携带登录态：

```
Authorization: Bearer <access_token>
```

`access_token` 通过 [2. 认证接口](#2-认证-auth) 的登录/注册接口获取。部分接口进一步要求当前用户拥有管理员角色（`admin` / `superadmin` / `administrator` 角色码，或用户名为 `admin`），下文用 **「需管理员权限」** 标注，权限不足返回 `403`。

### 1.6 鉴权方式二：自动化上报服务令牌 (Service Token)

自动化用例执行的**上报类接口**（`POST`/`PATCH`，供 pytest/CI Runner 调用，非人工浏览器场景）使用独立于用户登录态的服务令牌鉴权，请求头为：

```
X-Service-Token: <AUTOMATION_SERVICE_TOKEN>
```

该令牌为后端固定配置值（`.env` 中 `AUTOMATION_SERVICE_TOKEN`），**不是**从登录接口获取，也不依赖用户账号。前端 Dashboard 页面不需要关心此令牌，只需使用 1.5 的 JWT 登录态调用**查询类**接口（`GET`）即可。详见 [6. 自动化用例执行](#6-自动化用例执行-automation) 中每个接口的鉴权列。

---

## 2. 认证 Auth

无需鉴权（登录/注册本身），`/logout` 需要 JWT。

| Method | Path | 鉴权 | 说明 |
| --- | --- | --- | --- |
| POST | `/login/access-token` | 无 | OAuth2 表单登录（`application/x-www-form-urlencoded`，字段名 `username`/`password`），主要用于 Swagger 的 Authorize 按钮 |
| POST | `/login` | 无 | JSON 账号密码登录，**前端应使用此接口** |
| POST | `/register` | 无 | 注册新用户并自动登录，返回值与登录接口一致 |
| POST | `/logout` | JWT | 登出，清理该用户的 Redis 缓存 |

### POST /login （JSON 登录，推荐）

请求体：

```typescript
interface UserLogin {
  username: string;
  password: string;
}
```

响应 `data`：

```typescript
interface UserLoginData {
  access_token: string;
  token_type: "bearer";
  expires_in?: number;   // 过期时间(秒)
  user?: UserItem;       // 见 3.1 用户基础信息
}
```

失败：用户名或密码错误 / 账号被禁用 → `400`。

### POST /register

请求体：

```typescript
interface UserCreate {
  username: string;        // 3~64 字符
  password: string;        // 至少 6 位
  email: string;
  metadata?: {
    nickname?: string;
    phone?: string;
    avatar?: string;
    department?: string;
  };
  role_ids?: number[];     // 初始分配角色ID列表，可不传
}
```

响应结构与 `/login` 相同（`UserLoginData`）。失败：用户名/邮箱已被注册 → `400`。

### POST /logout

无请求体，需 `Authorization: Bearer`。响应 `data` 为空。

---

## 3. 用户管理 Users

前缀：`/users`。除 "获取角色下拉选项" 和 "获取用户详情" 外，**均需管理员权限**。

| Method | Path | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/users` | 需管理员权限 | 分页查询用户列表，支持多条件筛选 |
| POST | `/users` | 需管理员权限 | 创建新用户 |
| GET | `/users/role-options` | JWT | 获取可分配的角色下拉列表 |
| GET | `/users/{user_id}` | JWT | 获取指定用户详情 |
| PUT | `/users/{user_id}` | 需管理员权限 | 更新用户基本信息（用户名/邮箱/扩展信息） |
| PUT | `/users/{user_id}/roles` | 需管理员权限 | 覆盖式重新分配用户角色 |
| PUT | `/users/{user_id}/status` | 需管理员权限 | 启用/禁用用户 |
| POST | `/users/{user_id}/reset-password` | 需管理员权限 | 管理员强制重置指定用户密码 |
| POST | `/users/me/password` | JWT | 当前登录用户自行修改密码（需提供原密码） |
| POST | `/users/batch-status` | 需管理员权限 | 批量启用/禁用用户 |
| POST | `/users/batch-delete` | 需管理员权限 | 批量软删除用户 |

### 3.1 公共数据类型

```typescript
interface RoleItem {
  id: number;
  code: string;          // 如 admin, editor
  name: string;
  description?: string;
}

interface UserItem {
  id: number;
  username: string;
  email?: string;
  status: number;        // 0: 正常, 1: 禁用, 2: 已删除
  roles: RoleItem[];
  last_login_at?: string; // ISO datetime
  created_at?: string;
}

interface UserDetailData extends UserItem {
  last_login_ip?: string;
  metadata?: {
    nickname?: string;
    phone?: string;
    avatar?: string;
    department?: string;
  };
  updated_at?: string;
}
```

### GET /users

Query 参数：`page`(默认1), `page_size`(默认10, 最大100), `username`(模糊), `email`, `status`(0/1), `role_id`。

响应 `data`: `PageResult<UserItem>`

### POST /users

请求体：`UserCreate`（同 2. 注册接口）。响应 `data`: `UserItem`。失败：用户名/邮箱已存在 → `400`。

### GET /users/role-options

无参数。响应 `data`: `RoleItem[]`。

### GET /users/{user_id}

响应 `data`: `UserDetailData`。失败：用户不存在 → `404`。

### PUT /users/{user_id}

请求体：

```typescript
interface UserUpdate {
  user_id: number;
  username?: string;
  email?: string;
  metadata?: { nickname?: string; phone?: string; avatar?: string; department?: string };
}
```

响应 `data`: 空（仅 `message`）。

### PUT /users/{user_id}/roles

请求体：

```typescript
interface UserRoleAssign {
  user_id: number;
  role_ids: number[];   // 覆盖式设置，非增量
}
```

### PUT /users/{user_id}/status

请求体：`{ user_id: number; status: number }`（0: 启用, 1: 禁用）。

### POST /users/{user_id}/reset-password

请求体：`{ user_id: number; new_password: string }`（至少6位）。

### POST /users/me/password

请求体：

```typescript
interface UserChangePassword {
  old_password: string;
  new_password: string;  // 至少6位
}
```

失败：原密码错误 → `400`。

### POST /users/batch-status

请求体：`{ user_ids: number[]; status: number }`（`user_ids` 至少1个）。

### POST /users/batch-delete

请求体：`{ user_ids: number[] }`。

---

## 4. 角色管理 Roles

前缀：`/roles`，**全部接口需管理员权限**。

| Method | Path | 说明 |
| --- | --- | --- |
| GET | `/roles` | 获取全部角色列表（含权限关联） |
| POST | `/roles` | 创建角色并绑定初始权限 |
| GET | `/roles/{role_id}` | 获取角色详情 |
| PUT | `/roles/{role_id}` | 修改角色名称/描述 |
| DELETE | `/roles/{role_id}` | 删除角色（内置系统角色禁止删除） |
| PUT | `/roles/{role_id}/permissions` | 重新分配角色权限（覆盖式） |

### 公共数据类型

```typescript
interface PermissionItem {
  id: number;
  parent_id?: number;
  code: string;             // 如 user:create
  name: string;
  resource_type: number;    // 1-菜单, 2-按钮/功能, 3-API接口
  path?: string;
  method?: string;          // GET/POST/PUT/DELETE...
  created_at?: string;
}

interface RoleDetailData {
  id: number;
  code: string;
  name: string;
  description?: string;
  is_system: boolean;      // 内置角色不可删除
  permissions: PermissionItem[];
  created_at?: string;
  updated_at?: string;
}
```

### GET /roles

响应 `data`: `RoleDetailData[]`（非分页，返回全部）。

### POST /roles

请求体：

```typescript
interface RoleCreate {
  code: string;             // 2~64字符，如 editor
  name: string;             // 2~64字符
  description?: string;
  permission_ids?: number[];
}
```

响应 `data`: `RoleDetailData`。

### GET /roles/{role_id}

响应 `data`: `RoleDetailData`。失败：角色不存在 → `404`。

### PUT /roles/{role_id}

请求体：`{ name?: string; description?: string }`。响应 `data`: `RoleDetailData`。

### DELETE /roles/{role_id}

无请求体。响应 `data` 为空。失败：内置角色不可删除 / 角色不存在。

### PUT /roles/{role_id}/permissions

请求体：

```typescript
interface RolePermissionAssign {
  role_id: number;
  permission_ids: number[];  // 覆盖式设置
}
```

---

## 5. 权限管理 Permissions

前缀：`/permissions`，**全部接口需管理员权限**。

| Method | Path | 说明 |
| --- | --- | --- |
| GET | `/permissions/tree` | 获取树形结构权限数据（用于菜单/权限勾选组件） |
| GET | `/permissions` | 获取平铺权限列表 |
| POST | `/permissions` | 创建权限节点（菜单/按钮/接口） |
| DELETE | `/permissions/{permission_id}` | 删除权限节点 |

### GET /permissions/tree

响应 `data`:

```typescript
interface PermissionTreeItem extends PermissionItem {
  children: PermissionTreeItem[];
}
// data: PermissionTreeItem[]
```

### GET /permissions

响应 `data`: `PermissionItem[]`（平铺，无父子嵌套）。

### POST /permissions

请求体：

```typescript
interface PermissionCreate {
  parent_id?: number;
  code: string;             // 2~128字符，如 user:create
  name: string;             // 2~64字符
  resource_type?: number;   // 默认 1 (菜单)；1-菜单, 2-按钮/功能, 3-API接口
  path?: string;
  method?: string;
}
```

### DELETE /permissions/{permission_id}

无请求体。响应 `data` 为空。

---

## 6. 自动化用例执行 Automation

前缀：`/automation`。该模块用于记录 Jenkins 触发的每次 pytest 自动化执行（execution）、其下每个用例的每次尝试（item/attempt，**失败重试不会覆盖历史记录**）、每个用例内的步骤（step）以及步骤失败时的截图/日志等附件（attachment）。

> 详细数据模型设计背景见 [automation-execution-api-design.md](./automation-execution-api-design.md)（早期设计稿，个别字段以本篇 + 源码实现为准）。

### 6.1 鉴权分类 ⚠️

| 类型 | 接口 | 鉴权 | 调用方 |
| --- | --- | --- | --- |
| **上报类** | 下表中标注 `Service Token` 的所有 `POST`/`PATCH` 接口 | `X-Service-Token` Header（见 1.6） | pytest/CI Runner，**不是前端** |
| **查询类** | 下表中标注 `JWT` 的所有 `GET` 接口 | `Authorization: Bearer` | **前端 Dashboard** |

**前端页面只需对接“查询类”接口**（全部为 GET），无需关心服务令牌。

### 6.2 接口列表

| Method | Path | 鉴权 | 说明 |
| --- | --- | --- | --- |
| GET | `/automation/dashboard/overview` | JWT | 获取 dashboard 概览（卡片统计 + 趋势图 + 运行中 execution 卡片） |
| GET | `/automation/dashboard/running-executions/{execution_id}/cases` | JWT | 获取某个运行中 execution 的 case 快照明细 |
| GET | `/automation/jobs` | JWT | 分页获取 Jenkins 自动化任务列表 |
| POST | `/automation/executions` | Service Token | 创建一次 execution（pytest_sessionstart 上报） |
| PATCH | `/automation/executions/{execution_id}` | Service Token | 结束/更新一次 execution（pytest_sessionfinish 上报） |
| GET | `/automation/executions` | JWT | 分页获取 execution 列表 |
| GET | `/automation/executions/{execution_id}` | JWT | 获取单次 execution 详情 |
| POST | `/automation/executions/{execution_id}/items` | Service Token | 上报一次用例执行开始（含失败重试，自动分配 attempt_number） |
| PATCH | `/automation/executions/{execution_id}/items/{item_id}` | Service Token | 上报一次用例执行结束 |
| GET | `/automation/executions/{execution_id}/items` | JWT | 获取 execution 下用例列表（默认只返回每个 case 的最终结果） |
| GET | `/automation/executions/{execution_id}/items/attempts` | JWT | 获取某个用例在本次 execution 下的全部重试历史 |
| GET | `/automation/items/{item_id}` | JWT | 获取单次用例执行详情（含步骤与附件） |
| POST | `/automation/items/{item_id}/steps` | Service Token | 批量上报用例步骤（幂等，重复上报同一 step_index 会更新） |
| GET | `/automation/items/{item_id}/steps` | JWT | 获取用例执行的步骤列表 |
| POST | `/automation/steps/{step_id}/attachments` | Service Token | 批量登记步骤附件（截图/日志的外部 URL） |
| GET | `/automation/steps/{step_id}/attachments` | JWT | 获取步骤附件列表 |
| GET | `/automation/cases/history` | JWT | 获取某个用例跨 execution 的历史结果（稳定性/flaky 分析） |

### 6.3 公共数据类型 / 枚举

```typescript
type ExecutionStatus = "pending" | "running" | "completed" | "failed";
type CaseStatus = "pending" | "running" | "success" | "failure" | "skipped" | "error";
type StepStatus = "passed" | "failed" | "skipped" | "broken";
type AttachmentType = "screenshot" | "log" | "video" | "other";

interface ExecutionItem {
  id: number;
  job_id: number;              // Jenkins 任务 ID
  job_uid: string;             // 本次 build 唯一标识 (UUID)
  job_name: string;
  job_url: string;
  software_name?: string;
  software_version?: string;
  status: ExecutionStatus;
  start_at: string;            // ISO datetime
  end_at?: string;
  duration?: number;           // 秒
  pre_cases_count?: number;
  final_success_count?: number;
  final_failure_count?: number;
  final_skipped_count?: number;
  created_at?: string;
}

interface CaseItem {
  id: number;
  execution_id: number;
  case_key: string;            // pytest nodeid，同一用例的唯一标识
  case_id?: number;            // 外部用例管理系统 ID
  case_name: string;
  attempt_number: number;      // 1=首次，2=第一次重试...
  is_latest: boolean;          // 是否为该用例在本次 execution 下的最终结果
  status: CaseStatus;
  start_at?: string;
  end_at?: string;
  duration?: number;
  error_message?: string;
}

interface AttachmentItem {
  id: number;
  step_id: number;
  attachment_type: AttachmentType;
  file_url: string;
  content_type?: string;
  file_size?: number;
  created_at?: string;
}

interface StepItem {
  id: number;
  item_id: number;
  step_index: number;
  step_name: string;
  status: StepStatus;
  start_at?: string;
  end_at?: string;
  duration?: number;
  error_message?: string;
  attachment_count: number;
  attachments?: AttachmentItem[];  // 仅“用例执行详情”接口会填充，列表接口为 undefined
}

interface CaseHistoryItem {
  execution_id: number;
  job_name: string;
  item_id: number;
  attempt_number: number;      // 该 execution 内的最终 attempt 序号
  status: CaseStatus;
  start_at?: string;
  duration?: number;
}

interface ItemDetailData extends CaseItem {
  steps: StepItem[];
}

interface JobItem {
  id: number;
  job_name: string;
  job_url: string;
  status: string;
  last_build_number?: number;
  last_build_uid?: string;
  last_build_status?: string;
  last_build_timestamp?: string;
  last_build_duration?: number;
  pipeline_params?: Record<string, unknown>;
  sync_at?: string;
}

interface DashboardSummaryItem {
  active_jobs_count: number;
  total_executions_count: number;
  running_executions_count: number;
  success_cases_7d: number;
  failure_cases_7d: number;
  skipped_cases_7d: number;
  pass_rate_7d: number;
  avg_execution_duration_7d?: number;
  updated_at: string;
}

interface DashboardTrendItem {
  stat_date: string;
  execution_total: number;
  success_cases: number;
  failure_cases: number;
  skipped_cases: number;
  running_cases: number;
}

interface DashboardRunningExecutionItem {
  execution_id: number;
  job_id: number;
  job_name: string;
  job_url: string;
  status: ExecutionStatus;
  started_at: string;
  duration?: number;
  pre_cases_count: number;
  completed_cases_count: number;
  success_cases_count: number;
  failure_cases_count: number;
  skipped_cases_count: number;
  running_cases_count: number;
  progress_percent: number;
  running_case_names: string[];
  updated_at: string;
}

interface DashboardRunningCaseItem {
  execution_id: number;
  item_id: number;
  case_key: string;
  case_name: string;
  attempt_number: number;
  status: CaseStatus;
  start_at?: string;
  end_at?: string;
  duration?: number;
  error_message?: string;
  updated_at: string;
}

interface DashboardOverviewData {
  summary: DashboardSummaryItem;
  trends: DashboardTrendItem[];
  running_executions: DashboardRunningExecutionItem[];
}
```

### 6.4 逐接口详情（前端需对接的查询类接口）

#### GET /automation/dashboard/overview

Query:
- `days`（可选，默认 7，范围 1~30）
- `job_id`（可选，按 Jenkins 任务 ID 筛选）
- `job_name`（可选，按 Jenkins 任务名称模糊搜索）

响应 `data`: `DashboardOverviewData`

#### GET /automation/dashboard/running-executions/{execution_id}/cases

响应 `data`:

```typescript
{
  execution_id: number;
  items: DashboardRunningCaseItem[];
}
```

#### GET /automation/jobs

Query: `page`(默认1), `page_size`(默认10,最大100), `job_name?`, `status?`, `last_build_status?`。

响应 `data`: `PageResult<JobItem>`

#### GET /automation/executions

Query: `page`(默认1), `page_size`(默认10,最大100), `job_id?`, `job_name?`(模糊搜索), `status?`(ExecutionStatus)。

响应 `data`: `PageResult<ExecutionItem>`

#### GET /automation/executions/{execution_id}

响应 `data`: `ExecutionItem`。失败：execution 不存在 → `404`。

#### GET /automation/executions/{execution_id}/items

Query: `page`, `page_size`, `status?`(CaseStatus), `case_key?`(精确匹配), `case_name?`(模糊搜索)。

响应 `data`: `PageResult<CaseItem>`。**默认只返回每个 case_key 在本次 execution 下 `is_latest=true` 的最终结果**（不含被重试覆盖的历史尝试）。

#### GET /automation/executions/{execution_id}/items/attempts

Query: `case_key`（必填，pytest nodeid）。

响应 `data`: `PageResult<CaseItem>`，按 `attempt_number` 升序返回该用例在本次 execution 下的**全部**尝试记录（包括被覆盖的历史失败重试），用于展示"第几次重试才通过"等信息。

#### GET /automation/items/{item_id}

响应 `data`: `ItemDetailData`（含 `steps`，每个 step 含完整 `attachments` 数组）。失败：item 不存在 → `404`。

#### GET /automation/items/{item_id}/steps

响应 `data`: `StepItem[]`（按 `step_index` 升序，仅含 `attachment_count`，不含 `attachments` 详情，如需详情请调用上面的 item 详情接口或下面的附件列表接口）。

#### GET /automation/steps/{step_id}/attachments

响应 `data`: `AttachmentItem[]`。

#### GET /automation/cases/history

Query: `case_key`（必填），`page`，`page_size`。

响应 `data`: `PageResult<CaseHistoryItem>`，按 `start_at` 倒序返回该用例在各次 execution 中的最终结果（每个 execution 只取一条 `is_latest=true` 的记录），用于识别间歇性失败 (flaky) 用例的历史趋势。

### 6.5 上报类接口（供参考，前端一般无需调用）

以下接口由 pytest 插件 / CI Runner 在测试执行生命周期内自动调用，使用 `X-Service-Token` 鉴权：

- `POST /automation/executions`：`ExecutionCreate` → `ExecutionItem`
- `PATCH /automation/executions/{execution_id}`：`ExecutionUpdate` → `ExecutionItem`
- `POST /automation/executions/{execution_id}/items`：`ItemCreate` → `CaseItem`（重试时自动新建行、`attempt_number+1`，不覆盖旧数据）
- `PATCH /automation/executions/{execution_id}/items/{item_id}`：`ItemUpdate` → `CaseItem`
- `POST /automation/items/{item_id}/steps`：`StepsBatchCreate`（`steps: StepCreate[]`） → `StepItem[]`
- `POST /automation/steps/{step_id}/attachments`：`AttachmentsBatchCreate`（`attachments: AttachmentCreate[]`） → `AttachmentItem[]`

> 注：dashboard 快照数据由后端任务异步维护，前端只需要调用查询接口，不需要直接触发任务。

---

## 7. 前端对接建议

1. **统一响应拦截器**：对 2xx 响应直接取 `response.data.data`；对非 2xx 响应统一从 `error.response.data.detail` 提取错误信息（注意 `detail` 可能是字符串或校验错误数组，见 1.4）。
2. **鉴权头**：全局 axios/fetch 拦截器为除 `/login`、`/login/access-token`、`/register` 外的请求自动附加 `Authorization: Bearer <token>`；`X-Service-Token` 仅供 CI 侧使用，前端无需实现。
3. **401/403 处理**：`403` 常见于 JWT 过期/未登录（跳转登录页）或权限不足（提示无权限，不跳转登录页）；两者可通过 `detail` 文案区分（"Could not validate credentials" vs "权限不足..."）。`401` 目前仅出现在自动化服务令牌校验失败场景，前端一般不会遇到。
4. **分页组件**：所有分页列表均遵循同一个 `PageResult<T>` 结构，可封装通用分页 Hook/组件复用。
5. **自动化模块页面**：Dashboard 首屏建议调用 `/automation/dashboard/overview`，运行中执行卡片详情调用 `/automation/dashboard/running-executions/{execution_id}/cases`；任务页可调用 `/automation/jobs`；执行列表调用 `/automation/executions`；如需查看某条用例的重试历程，调用 `/automation/executions/{execution_id}/items/attempts`；如需查看某条用例跨多次 CI 构建的历史稳定性，调用 `/automation/cases/history`。
