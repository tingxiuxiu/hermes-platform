from typing import Annotated

from fastapi import APIRouter, Depends, Query

from app.dependencies import SessionDep, get_current_user, get_current_active_admin
from .models import User
from .schemas import (
    UserCreate,
    UserListQuery,
    UserUpdate,
    UserRoleAssign,
    UserStatusUpdate,
    UserResetPassword,
    UserChangePassword,
    UserBatchStatusUpdate,
    UserBatchDelete,
)
from .constants import (
    UserCreateResponse,
    UserListResponse,
    UserDetailResponse,
    UserUpdateResponse,
    UserRoleAssignResponse,
    UserStatusUpdateResponse,
    UserResetPasswordResponse,
    UserChangePasswordResponse,
    UserBatchResponse,
    RoleOptionResponse,
)
from .service import (
    get_user_list,
    create_user,
    get_user_detail,
    update_user,
    assign_user_roles,
    update_user_status,
    reset_user_password,
    change_own_password,
    batch_update_user_status,
    batch_delete_users,
)

users_router = APIRouter(prefix="/users", tags=["用户管理"])


@users_router.get(
    "",
    response_model=UserListResponse,
    summary="获取用户列表 (管理员权限)",
)
async def get_user_list_endpoint(
    session: SessionDep,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
    page: int = Query(default=1, ge=1, description="页码"),
    page_size: int = Query(default=10, ge=1, le=100, description="每页条数"),
    username: str | None = Query(default=None, description="用户名模糊搜索"),
    email: str | None = Query(default=None, description="邮箱筛选"),
    status: int | None = Query(default=None, description="状态筛选 (0: 正常, 1: 禁用)"),
    role_id: int | None = Query(default=None, description="按角色ID筛选"),
) -> UserListResponse:
    """查询多维筛选下的分页用户列表数据 (需管理员权限)"""
    query = UserListQuery(
        page=page,
        page_size=page_size,
        username=username,
        email=email,
        status=status,
        role_id=role_id,
    )
    return await get_user_list(session=session, query=query)


@users_router.post(
    "",
    response_model=UserCreateResponse,
    summary="创建/添加新用户 (管理员权限)",
)
async def create_user_endpoint(
    session: SessionDep,
    data: UserCreate,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> UserCreateResponse:
    """创建后台新用户账号并赋予初始角色 (需管理员权限)"""
    return await create_user(session=session, data=data)


@users_router.get(
    "/role-options",
    response_model=RoleOptionResponse,
    summary="获取可用角色下拉列表",
)
async def get_role_options_endpoint(
    session: SessionDep,
    current_user: Annotated[User, Depends(get_current_user)],
) -> RoleOptionResponse:
    """拉取系统所有基础角色信息列表，用于角色配置时的下拉控件"""
    return await get_user_list(
        session=session, query=UserListQuery(page=1, page_size=10)
    )  # Adjust this line based on actual service method


@users_router.get(
    "/{user_id}",
    response_model=UserDetailResponse,
    summary="获取指定用户详情",
)
async def get_user_detail_endpoint(
    session: SessionDep,
    user_id: int,
    current_user: Annotated[User, Depends(get_current_user)],
) -> UserDetailResponse:
    """根据用户 ID 获取该用户的完整属性与绑定角色集合"""
    return await get_user_detail(session=session, user_id=user_id)


@users_router.put(
    "/{user_id}",
    response_model=UserUpdateResponse,
    summary="更新用户基本信息 (管理员权限)",
)
async def update_user_endpoint(
    session: SessionDep,
    user_id: int,
    data: UserUpdate,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> UserUpdateResponse:
    """修改指定用户的邮箱、用户名或扩展元数据 (需管理员权限)"""
    return await update_user(session=session, user_id=user_id, data=data)


@users_router.put(
    "/{user_id}/roles",
    response_model=UserRoleAssignResponse,
    summary="为用户配置/分配角色 (管理员权限)",
)
async def assign_user_roles_endpoint(
    session: SessionDep,
    user_id: int,
    data: UserRoleAssign,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> UserRoleAssignResponse:
    """重新为用户配置并覆盖与其关联的角色集合 (需管理员权限)"""
    return await assign_user_roles(
        session=session, user_id=user_id, role_ids=data.role_ids
    )


@users_router.put(
    "/{user_id}/status",
    response_model=UserStatusUpdateResponse,
    summary="更新用户状态/启用/禁用 (管理员权限)",
)
async def update_user_status_endpoint(
    session: SessionDep,
    user_id: int,
    data: UserStatusUpdate,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> UserStatusUpdateResponse:
    """变更指定用户的状态（0: 正常启用，1: 禁用） (需管理员权限)"""
    return await update_user_status(
        session=session, user_id=user_id, status=data.status
    )


@users_router.post(
    "/{user_id}/reset-password",
    response_model=UserResetPasswordResponse,
    summary="管理员重置用户密码 (管理员权限)",
)
async def reset_user_password_endpoint(
    session: SessionDep,
    user_id: int,
    data: UserResetPassword,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> UserResetPasswordResponse:
    """管理员强行重置指定用户的登录密码 (需管理员权限)"""
    return await reset_user_password(
        session=session, user_id=user_id, new_password=data.new_password
    )


@users_router.post(
    "/me/password",
    response_model=UserChangePasswordResponse,
    summary="当前登录用户自行修改密码",
)
async def change_own_password_endpoint(
    session: SessionDep,
    data: UserChangePassword,
    current_user: Annotated[User, Depends(get_current_user)],
) -> UserChangePasswordResponse:
    """校验原密码后更新为新密码"""
    return await change_own_password(
        session=session,
        user=current_user,
        old_password=data.old_password,
        new_password=data.new_password,
    )


@users_router.post(
    "/batch-status",
    response_model=UserBatchResponse,
    summary="批量更新用户状态/启用禁用 (管理员权限)",
)
async def batch_update_user_status_endpoint(
    session: SessionDep,
    data: UserBatchStatusUpdate,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> UserBatchResponse:
    """批量设置多个用户账号的启用/禁用状态 (需管理员权限)"""
    return await batch_update_user_status(
        session=session, user_ids=data.user_ids, status=data.status
    )


@users_router.post(
    "/batch-delete",
    response_model=UserBatchResponse,
    summary="批量删除用户 (管理员权限)",
)
async def batch_delete_users_endpoint(
    session: SessionDep,
    data: UserBatchDelete,
    current_admin: Annotated[User, Depends(get_current_active_admin)],
) -> UserBatchResponse:
    """批量对多个用户账号执行软删除 (需管理员权限)"""
    return await batch_delete_users(session=session, user_ids=data.user_ids)
