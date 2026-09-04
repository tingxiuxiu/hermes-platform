from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field


# ==============================================================================
# 1. 基础组件与关联实体模型 (Base & Nested Entities)
# ==============================================================================
class RoleItem(BaseModel):
    """角色简要信息"""

    id: int = Field(..., description="角色ID")
    code: str = Field(..., description="角色标识 (如: admin, editor)")
    name: str = Field(..., description="角色名称 (如: 管理员, 编辑)")
    description: Optional[str] = Field(None, description="角色描述")


class UserMetadata(BaseModel):
    """用户扩展信息模型 (存储在 metadata 字段中)"""

    nickname: Optional[str] = Field(None, description="用户昵称")
    phone: Optional[str] = Field(None, description="手机号码")
    avatar: Optional[str] = Field(None, description="头像地址")
    department: Optional[str] = Field(None, description="所属部门")


class UserItem(BaseModel):
    """用户列表/基础信息数据项"""

    id: int = Field(..., description="用户ID")
    username: str = Field(..., description="用户名")
    email: Optional[str] = Field(None, description="用户邮箱")
    status: int = Field(..., description="状态 (0: 正常, 1: 禁用, 2: 已删除)")
    roles: List[RoleItem] = Field(default_factory=list, description="所属角色列表")
    last_login_at: Optional[datetime] = Field(None, description="最后登录时间")
    created_at: Optional[datetime] = Field(None, description="创建时间")


# ==============================================================================
# 2. 账号注册 / 登录 / 登出模型 (Authentication & Registration)
# ==============================================================================
class UserCreate(BaseModel):
    """创建/注册用户请求"""

    username: str = Field(..., min_length=3, max_length=64, description="用户名")
    password: str = Field(..., min_length=6, description="密码")
    email: str = Field(..., description="邮箱")
    metadata: Optional[UserMetadata] = Field(
        None, description="扩展字段(昵称、手机号等)"
    )
    role_ids: Optional[List[int]] = Field(
        default_factory=list, description="初始分配的角色ID列表"
    )


class UserLogin(BaseModel):
    """用户登录请求"""

    username: str = Field(..., description="用户名")
    password: str = Field(..., description="密码")


class UserLogout(BaseModel):
    """用户登出请求"""

    id: int = Field(..., description="用户ID")


# ==============================================================================
# 3. 用户列表与分页模型 (User List & Pagination)
# ==============================================================================
class UserListQuery(BaseModel):
    """用户列表查询参数"""

    page: int = Field(default=1, ge=1, description="页码")
    page_size: int = Field(default=10, ge=1, le=100, description="每页条数")
    username: Optional[str] = Field(None, description="用户名模糊匹配")
    email: Optional[str] = Field(None, description="邮箱筛选")
    status: Optional[int] = Field(None, description="状态筛选 (0: 正常, 1: 禁用)")
    role_id: Optional[int] = Field(None, description="按角色ID筛选")


# ==============================================================================
# 4. 用户角色配置模型 (User Role Assignment)
# ==============================================================================
class UserRoleAssign(BaseModel):
    """分配/变更用户角色请求"""

    user_id: int = Field(..., description="用户ID")
    role_ids: List[int] = Field(..., description="角色ID列表")


# ==============================================================================
# 5. 用户状态管理模型 (Enable / Disable / Activate / Inactive)
# ==============================================================================
class UserStatusUpdate(BaseModel):
    """更新用户状态请求 (通用启用/禁用)"""

    user_id: int = Field(..., description="用户ID")
    status: int = Field(..., description="目标状态 (0: 启用/正常, 1: 禁用)")


class UserInactive(BaseModel):
    """禁用用户请求 (兼容单项语法)"""

    id: int = Field(..., description="用户ID")


class UserActivate(BaseModel):
    """启用用户请求 (兼容单项语法)"""

    id: int = Field(..., description="用户ID")


# ==============================================================================
# 6. 用户详情模型 (User Detail)
# ==============================================================================
class UserDetailData(UserItem):
    """用户详情数据对象"""

    last_login_ip: Optional[str] = Field(None, description="最后登录IP")
    metadata: Optional[UserMetadata] = Field(None, description="用户扩展信息")
    updated_at: Optional[datetime] = Field(None, description="更新时间")


# ==============================================================================
# 7. 用户基本信息修改模型 (User Profile Update)
# ==============================================================================
class UserUpdate(BaseModel):
    """修改用户信息请求"""

    user_id: int = Field(..., description="用户ID")
    username: Optional[str] = Field(None, description="用户名")
    email: Optional[str] = Field(None, description="邮箱")
    metadata: Optional[UserMetadata] = Field(
        None, description="扩展字段(昵称、手机号等)"
    )


# ==============================================================================
# 8. 密码管理模型 (Password Management)
# ==============================================================================
class UserResetPassword(BaseModel):
    """管理员重置用户密码请求"""

    user_id: int = Field(..., description="用户ID")
    new_password: str = Field(..., min_length=6, description="新密码")


class UserChangePassword(BaseModel):
    """用户自行修改密码请求"""

    old_password: str = Field(..., description="原密码")
    new_password: str = Field(..., min_length=6, description="新密码")


# ==============================================================================
# 9. 批量操作模型 (Batch Operations)
# ==============================================================================
class UserBatchStatusUpdate(BaseModel):
    """批量启用/禁用用户请求"""

    user_ids: List[int] = Field(..., min_length=1, description="目标用户ID列表")
    status: int = Field(..., description="目标状态 (0: 启用, 1: 禁用)")


class UserBatchDelete(BaseModel):
    """批量删除用户请求"""

    user_ids: List[int] = Field(..., min_length=1, description="待删除的用户ID列表")


# ==============================================================================
# 10. 角色/权限辅助下拉选择模型 (Auxiliary Role Selection)
# ==============================================================================


class PermissionItem(BaseModel):
    """权限基础视图项"""

    id: int = Field(..., description="权限ID")
    parent_id: Optional[int] = Field(None, description="父权限ID")
    code: str = Field(..., description="权限唯一标识 (如: user:create, order:read)")
    name: str = Field(..., description="权限显示名称 (如: 创建用户)")
    resource_type: int = Field(
        ..., description="资源类型 (1-菜单, 2-按钮/功能, 3-API接口)"
    )
    path: Optional[str] = Field(None, description="前端路由或 API 路径")
    method: Optional[str] = Field(None, description="HTTP 请求方式 (GET, POST等)")
    created_at: Optional[datetime] = Field(None, description="创建时间")


class PermissionTreeItem(PermissionItem):
    """树形结构权限数据项"""

    children: List["PermissionTreeItem"] = Field(
        default_factory=list, description="子权限列表"
    )


class PermissionCreate(BaseModel):
    """创建权限节点请求"""

    parent_id: Optional[int] = Field(None, description="父权限ID")
    code: str = Field(
        ..., min_length=2, max_length=128, description="权限唯一标识 (如 user:create)"
    )
    name: str = Field(..., min_length=2, max_length=64, description="权限名称")
    resource_type: int = Field(
        default=1, description="资源类型 (1-菜单, 2-按钮/功能, 3-API接口)"
    )
    path: Optional[str] = Field(None, description="前端路径或 API Path")
    method: Optional[str] = Field(
        None, description="HTTP 请求方式 (GET, POST, PUT, DELETE等)"
    )


class PermissionUpdate(BaseModel):
    """更新权限节点请求"""

    name: Optional[str] = Field(None, description="权限名称")
    path: Optional[str] = Field(None, description="路径")
    method: Optional[str] = Field(None, description="HTTP请求方式")


# ==============================================================================
# 角色实体及请求模型
# ==============================================================================


class RoleCreate(BaseModel):
    """创建角色请求"""

    code: str = Field(
        ..., min_length=2, max_length=64, description="角色唯一标识 (如: editor, admin)"
    )
    name: str = Field(
        ..., min_length=2, max_length=64, description="角色显示名称 (如: 编辑人员)"
    )
    description: Optional[str] = Field(None, description="角色描述信息")
    permission_ids: Optional[List[int]] = Field(
        default_factory=list, description="初始绑定的权限ID列表"
    )


class RoleUpdate(BaseModel):
    """修改角色基本信息请求"""

    name: Optional[str] = Field(None, description="角色名称")
    description: Optional[str] = Field(None, description="角色描述")


class RolePermissionAssign(BaseModel):
    """配置/分配角色权限请求"""

    role_id: int = Field(..., description="目标角色ID")
    permission_ids: List[int] = Field(..., description="绑定的权限ID列表")
