from datetime import datetime
from typing import Optional, List, Any
from pydantic import BaseModel, Field
from app.core.base_response import NormalResponse, ErrorResponse
from app.domains.user.schema import UserItem, RoleItem, UserDetailData


class UserCreateResponse(NormalResponse):
    """创建用户成功响应"""
    message: str = Field(default="Create User Success", description="响应消息")
    data: Optional[UserItem] = Field(None, description="创建成功的用户信息")


class UserCreateErrorResponse(ErrorResponse):
    """创建用户失败响应"""
    message: str = Field(default="Create User Failed", description="响应消息")


class UserLoginData(BaseModel):
    """登录成功返回的数据"""
    access_token: str = Field(..., description="访问令牌")
    token_type: str = Field(default="bearer", description="令牌类型")
    expires_in: Optional[int] = Field(None, description="过期时间(秒)")
    user: Optional[UserItem] = Field(None, description="当前登录用户信息")


class UserLoginResponse(NormalResponse):
    """用户登录成功响应"""
    message: str = Field(default="User Login Success", description="响应消息")
    data: UserLoginData = Field(..., description="登录凭证与用户信息")


class UserLoginErrorResponse(ErrorResponse):
    """用户登录失败响应"""
    message: str = Field(default="User Login Failed", description="响应消息")


class UserLogoutResponse(NormalResponse):
    """用户登出成功响应"""
    message: str = Field(default="User Logout Success", description="响应消息")


class UserLogoutErrorResponse(ErrorResponse):
    """用户登出失败响应"""
    message: str = Field(default="User Logout Failed", description="响应消息")


class UserListData(BaseModel):
    """用户列表分页数据"""
    total: int = Field(..., description="总记录数")
    page: int = Field(..., description="当前页码")
    page_size: int = Field(..., description="每页条数")
    items: List[UserItem] = Field(..., description="用户列表项")


class UserListResponse(NormalResponse):
    """获取用户列表成功响应"""
    message: str = Field(default="Get User List Success", description="响应消息")
    data: UserListData = Field(..., description="分页列表数据")


class UserListErrorResponse(ErrorResponse):
    """获取用户列表失败响应"""
    message: str = Field(default="Get User List Failed", description="响应消息")


class UserRoleAssignResponse(NormalResponse):
    """分配用户角色成功响应"""
    message: str = Field(default="Assign User Roles Success", description="响应消息")


class UserRoleAssignErrorResponse(ErrorResponse):
    """分配用户角色失败响应"""
    message: str = Field(default="Assign User Roles Failed", description="响应消息")

class UserStatusUpdateResponse(NormalResponse):
    """更新用户状态成功响应"""
    message: str = Field(default="Update User Status Success", description="响应消息")


class UserStatusUpdateErrorResponse(ErrorResponse):
    """更新用户状态失败响应"""
    message: str = Field(default="Update User Status Failed", description="响应消息")

class UserInactiveResponse(NormalResponse):
    """禁用用户成功响应"""
    message: str = Field(default="User Inactive Success", description="响应消息")


class UserInactiveErrorResponse(ErrorResponse):
    """禁用用户失败响应"""
    message: str = Field(default="User Inactive Failed", description="响应消息")


class UserActivateResponse(NormalResponse):
    """启用用户成功响应"""
    message: str = Field(default="User Activate Success", description="响应消息")


class UserActivateErrorResponse(ErrorResponse):
    """启用用户失败响应"""
    message: str = Field(default="User Activate Failed", description="响应消息")


class UserDetailResponse(NormalResponse):
    """获取用户详情成功响应"""
    message: str = Field(default="Get User Detail Success", description="响应消息")
    data: UserDetailData = Field(..., description="用户详情数据")


class UserDetailErrorResponse(ErrorResponse):
    """获取用户详情失败响应"""
    message: str = Field(default="Get User Detail Failed", description="响应消息")

class UserUpdateResponse(NormalResponse):
    """修改用户信息成功响应"""
    message: str = Field(default="Update User Success", description="响应消息")


class UserUpdateErrorResponse(ErrorResponse):
    """修改用户信息失败响应"""
    message: str = Field(default="Update User Failed", description="响应消息")


class UserResetPasswordResponse(NormalResponse):
    """管理员重置密码成功响应"""
    message: str = Field(default="Reset Password Success", description="响应消息")


class UserResetPasswordErrorResponse(ErrorResponse):
    """管理员重置密码失败响应"""
    message: str = Field(default="Reset Password Failed", description="响应消息")


class UserChangePasswordResponse(NormalResponse):
    """用户修改密码成功响应"""
    message: str = Field(default="Change Password Success", description="响应消息")


class UserChangePasswordErrorResponse(ErrorResponse):
    """用户修改密码失败响应"""
    message: str = Field(default="Change Password Failed", description="响应消息")


class UserBatchResponse(NormalResponse):
    """批量操作成功响应"""
    message: str = Field(default="Batch Operation Success", description="响应消息")


class UserBatchErrorResponse(ErrorResponse):
    """批量操作失败响应"""
    message: str = Field(default="Batch Operation Failed", description="响应消息")


class RoleOptionResponse(NormalResponse):
    """角色选择列表响应 (用于前端配置用户角色时的下拉选框)"""
    message: str = Field(default="Get Role Options Success", description="响应消息")
    data: List[RoleItem] = Field(..., description="角色可用列表")
