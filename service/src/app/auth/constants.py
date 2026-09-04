from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field
from app.core.base_response import BaseResponse


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


class UserCreateResponse(BaseResponse):
    """创建用户成功响应"""

    data: Optional[UserItem] = Field(None, description="创建成功的用户信息")


class UserLoginData(BaseModel):
    """登录成功返回的数据"""

    access_token: str = Field(..., description="访问令牌")
    token_type: str = Field(default="bearer", description="令牌类型")
    expires_in: Optional[int] = Field(None, description="过期时间(秒)")
    user: Optional[UserItem] = Field(None, description="当前登录用户信息")


class UserDetailData(UserItem):
    """用户详情数据对象"""

    last_login_ip: Optional[str] = Field(None, description="最后登录IP")
    metadata: Optional[UserMetadata] = Field(None, description="用户扩展信息")
    updated_at: Optional[datetime] = Field(None, description="更新时间")


class UserLoginResponse(BaseResponse):
    """用户登录成功响应"""

    data: UserLoginData = Field(..., description="登录凭证与用户信息")


class UserLogoutResponse(BaseResponse):
    """用户登出成功响应"""

    ...


class UserListData(BaseModel):
    """用户列表分页数据"""

    total: int = Field(..., description="总记录数")
    page: int = Field(..., description="当前页码")
    page_size: int = Field(..., description="每页条数")
    items: List[UserItem] = Field(..., description="用户列表项")


class UserListResponse(BaseResponse):
    """获取用户列表成功响应"""

    data: UserListData = Field(..., description="分页列表数据")


class UserRoleAssignResponse(BaseResponse):
    """分配用户角色成功响应"""

    ...


class UserStatusUpdateResponse(BaseResponse):
    """更新用户状态成功响应"""

    ...


class UserInactiveResponse(BaseResponse):
    """禁用用户成功响应"""

    ...


class UserActivateResponse(BaseResponse):
    """启用用户成功响应"""

    ...


class UserDetailResponse(BaseResponse):
    """获取用户详情成功响应"""

    data: UserDetailData = Field(..., description="用户详情数据")


class UserUpdateResponse(BaseResponse):
    """修改用户信息成功响应"""

    message: str = Field(default="Update User Success", description="响应消息")


class UserResetPasswordResponse(BaseResponse):
    """管理员重置密码成功响应"""

    message: str = Field(default="Reset Password Success", description="响应消息")


class UserChangePasswordResponse(BaseResponse):
    """用户修改密码成功响应"""

    message: str = Field(default="Change Password Success", description="响应消息")


class UserBatchResponse(BaseResponse):
    """批量操作成功响应"""

    message: str = Field(default="Batch Operation Success", description="响应消息")


class RoleOptionResponse(BaseResponse):
    """角色选择列表响应 (用于前端配置用户角色时的下拉选框)"""

    data: List[RoleItem] = Field(..., description="角色可用列表")


class RoleDetailData(BaseModel):
    """角色详情与所拥有的权限关联数据"""

    id: int = Field(..., description="角色ID")
    code: str = Field(..., description="角色标识")
    name: str = Field(..., description="角色名称")
    description: Optional[str] = Field(None, description="角色描述")
    is_system: bool = Field(default=False, description="是否为内置系统角色")
    permissions: List[PermissionItem] = Field(
        default_factory=list, description="权限列表"
    )
    created_at: Optional[datetime] = Field(None, description="创建时间")
    updated_at: Optional[datetime] = Field(None, description="更新时间")


class RoleListResponse(BaseResponse):
    """角色列表响应"""

    data: List[RoleDetailData] = Field(..., description="角色列表数据")


class RoleDetailResponse(BaseResponse):
    """角色详情响应"""

    data: RoleDetailData = Field(..., description="角色详情数据")


class PermissionTreeResponse(BaseResponse):
    """权限树型结构响应"""

    data: List[PermissionTreeItem] = Field(..., description="树形权限列表")


class PermissionListResponse(BaseResponse):
    """权限平铺列表响应"""

    data: List[PermissionItem] = Field(..., description="平铺权限列表")


class PermissionAssignResponse(BaseResponse):
    """分配权限成功响应"""

    ...


class RoleDeleteResponse(BaseResponse):
    """删除角色成功响应"""

    ...


class PermissionDeleteResponse(BaseResponse):
    """删除权限成功响应"""

    ...


class PermissionCreateResponse(BaseResponse):
    """创建权限成功响应"""

    data: PermissionItem = Field(..., description="创建成功的权限信息")
